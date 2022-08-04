package v2

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"time"

	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"

	"github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/libp2p/go-libp2p-core/protocol"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
)

//go:generate mockery --name Service --structname ServiceMock --filename service_mock.go --inpackage

type Service interface {
	CreateIdentity(ctx context.Context, req *CreateIdentityRequest) (*CreateIdentityResponse, error)

	ValidateKey(ctx context.Context, accountID *types.AccountID, pubKey []byte, keyPurpose types.KeyPurpose) error
	ValidateSignature(ctx context.Context, accountID *types.AccountID, pubKey []byte, signature []byte, message []byte) error
	ValidateAccount(ctx context.Context, accountID *types.AccountID) error

	GetLastKeyByPurpose(ctx context.Context, accountID *types.AccountID, keyPurpose types.KeyPurpose) (*types.Hash, error)
}

type service struct {
	configService        config.Service
	centAPI              centchain.API
	dispatcher           jobs.Dispatcher
	keystoreAPI          keystore.API
	protocolIDDispatcher dispatcher.Dispatcher[protocol.ID]
	log                  *logging.ZapEventLogger
}

func NewService(
	configService config.Service,
	centAPI centchain.API,
	dispatcher jobs.Dispatcher,
	keystoreAPI keystore.API,
	protocolIDDispatcher dispatcher.Dispatcher[protocol.ID],
) Service {
	log := logging.Logger("identity-service-v2")

	return &service{
		configService,
		centAPI,
		dispatcher,
		keystoreAPI,
		protocolIDDispatcher,
		log,
	}
}

func (s *service) CreateIdentity(ctx context.Context, req *CreateIdentityRequest) (*CreateIdentityResponse, error) {
	if err := s.validateCreateIdentityRequest(ctx, req); err != nil {
		return nil, err
	}

	cfg, err := s.configService.GetConfig()

	if err != nil {
		s.log.Errorf("Couldn't retrieve node config: %s", err)

		return nil, ErrNodeConfigRetrieval
	}

	p2pPrivateKey, p2pPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		s.log.Errorf("Couldn't retrieve p2p key pair: %s", err)

		return nil, ErrP2PKeyPairRetrieval
	}

	signingPrivateKey, signingPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetSigningKeyPair())

	if err != nil {
		s.log.Errorf("Couldn't retrieve signing key pair: %s", err)

		return nil, ErrSigningKeyPairRetrieval
	}

	acc, err := configstore.NewAccount(
		req.Identity,
		p2pPublicKey,
		p2pPrivateKey,
		signingPublicKey,
		signingPrivateKey,
		req.WebhookURL,
		req.PrecommitEnabled,
		req.AccountProxies,
	)

	if err != nil {
		s.log.Errorf("Couldn't create account: %s", err)

		return nil, ErrAccountCreation
	}

	if err := s.configService.CreateAccount(acc); err != nil {
		s.log.Errorf("Couldn't store account: %s", err)

		return nil, ErrAccountStorage
	}

	protocolID := p2pcommon.ProtocolForIdentity(acc.GetIdentity())

	err = s.protocolIDDispatcher.Dispatch(ctx, protocolID)

	if err != nil {
		s.log.Errorf("Couldn't dispatch protocol ID: %s", err)

		return nil, ErrProtocolIDDispatch
	}

	keys, err := createKeystoreKeys(p2pPublicKey, signingPublicKey)

	if err != nil {
		s.log.Errorf("Couldn't create keystore keys: %s", err)

		return nil, ErrKeystoreKeysCreation
	}

	ctx = contextutil.WithAccount(ctx, acc)

	jobID, err := s.dispatchAddKeysJob(ctx, acc.GetIdentity(), keys)

	if err != nil {
		s.log.Errorf("Couldn't dispatch job: %s", err)

		return nil, ErrJobDispatch
	}

	return &CreateIdentityResponse{
		Identity: acc.GetIdentity(),
		JobID:    jobID.Hex(),
	}, nil
}

func createKeystoreKeys(p2pPublicKey libp2pcrypto.PubKey, signingPublicKey libp2pcrypto.PubKey) ([]*types.AddKey, error) {
	p2pPublickKeyRaw, err := p2pPublicKey.Raw()

	if err != nil {
		return nil, fmt.Errorf("couldn't get raw p2p public key: %w", err)
	}

	signingPublicKeyRaw, err := signingPublicKey.Raw()

	if err != nil {
		return nil, fmt.Errorf("couldn't get raw signing public key: %w", err)
	}

	p2pKeyHash, err := crypto.Blake2bHash(p2pPublickKeyRaw)

	if err != nil {
		return nil, fmt.Errorf("couldn't hash p2p public key: %s", err)
	}

	signingKeyHash, err := crypto.Blake2bHash(signingPublicKeyRaw)

	if err != nil {
		return nil, fmt.Errorf("couldn't hash signing public key: %s", err)
	}

	keys := []*types.AddKey{
		{
			Key:     types.NewHash(p2pKeyHash),
			Purpose: types.KeyPurposeP2PDiscovery,
			KeyType: types.KeyTypeECDSA,
		},
		{
			Key:     types.NewHash(signingKeyHash),
			Purpose: types.KeyPurposeP2PDocumentSigning,
			KeyType: types.KeyTypeECDSA,
		},
	}

	return keys, nil
}

func (s *service) dispatchAddKeysJob(ctx context.Context, identity *types.AccountID, keys []*types.AddKey) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Add keys to keystore",
		addKeysJob,
		"add_keys_to_keystore",
		[]interface{}{
			ctx,
			keys,
		},
		make(map[string]interface{}),
		time.Time{},
	)

	if _, err := s.dispatcher.Dispatch(identity, job); err != nil {
		s.log.Errorf("Couldn't dispatch add keys job: %s", err)

		return nil, fmt.Errorf("failed to dispatch add keys job: %w", err)
	}

	return job.ID, nil
}

func (s *service) validateCreateIdentityRequest(ctx context.Context, req *CreateIdentityRequest) error {
	accID, err := types.NewAccountID(req.Identity[:])

	if err != nil {
		s.log.Errorf("Couldn't create account ID: %s", err)

		return ErrAccountIDCreation
	}

	if err := s.ValidateAccount(ctx, accID); err != nil {
		s.log.Errorf("Invalid identity - %s: %s", accID.ToHexString(), err)

		return ErrInvalidAccount
	}

	if _, err := url.ParseRequestURI(req.WebhookURL); err != nil {
		s.log.Errorf("Invalid webhook URL: %s", err)

		return ErrInvalidWebhookURL
	}

	for _, accountProxy := range req.AccountProxies {
		proxyAccID, err := types.NewAccountID(accountProxy.AccountID[:])

		if err != nil {
			s.log.Errorf("Couldn't create account ID for proxy: %s", err)

			return ErrAccountIDCreation
		}

		if err := s.ValidateAccount(ctx, proxyAccID); err != nil {
			s.log.Errorf("Invalid proxy account - %s: %s", proxyAccID.ToHexString(), err)

			return ErrInvalidProxyAccount
		}
	}

	return nil
}

func (s *service) ValidateKey(
	ctx context.Context,
	accountID *types.AccountID,
	pubKey []byte,
	keyPurpose types.KeyPurpose,
) error {
	// TODO(cdamian): Add validation from the NFT branch

	// TODO(cdamian): Do we need to hash this?
	keyID := &types.KeyID{
		Hash:       types.NewHash(pubKey),
		KeyPurpose: keyPurpose,
	}

	account, err := s.configService.GetAccount(accountID.ToBytes())

	if err != nil {
		s.log.Errorf("Couldn't retrieve account: %s", err)

		return ErrAccountRetrieval
	}

	ctx = contextutil.WithAccount(ctx, account)

	key, err := s.keystoreAPI.GetKey(ctx, keyID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve key: %s", err)

		return ErrKeyRetrieval
	}

	if key == nil {
		s.log.Errorf("Key not found")

		return ErrKeyNotFound
	}

	if key.RevokedAt.IsNone() {
		return nil
	}

	ok, revokedAt := key.RevokedAt.Unwrap()

	if !ok {
		s.log.Errorf("Invalid key data - revoked block number should be present")

		return ErrInvalidKeyData
	}

	latestBlock, err := s.centAPI.GetBlockLatest()

	if err != nil {
		s.log.Errorf("Couldn't retrieve latest block")

		return ErrLatestBlockRetrieval
	}

	if latestBlock.Block.Header.Number >= revokedAt {
		s.log.Errorf("Key is revoked")

		return ErrKeyRevoked
	}

	return nil
}

func (s *service) ValidateSignature(
	ctx context.Context,
	accountID *types.AccountID,
	pubKey []byte,
	message []byte,
	signature []byte,
) error {
	if err := s.ValidateKey(ctx, accountID, pubKey, types.KeyPurposeP2PDocumentSigning); err != nil {
		s.log.Errorf("Couldn't validate key: %s", err)

		return err
	}

	if !crypto.VerifyMessage(pubKey, message, signature, crypto.CurveEd25519) {
		s.log.Error("Couldn't verify message")

		return ErrInvalidSignature
	}

	return nil
}

func (s *service) ValidateAccount(_ context.Context, accountID *types.AccountID) error {
	meta, err := s.centAPI.GetMetadataLatest()

	if err != nil {
		s.log.Errorf("Couldn't get latest metadata: %s", err)

		return ErrMetadataRetrieval
	}

	accountStorageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID[:])

	if err != nil {
		s.log.Errorf("Couldn't create storage key for account: %s", err)

		return ErrAccountStorageKeyCreation
	}

	var accountInfo types.AccountInfo

	// TODO(cdamian): Use OK from NFT branch
	if err = s.centAPI.GetStorageLatest(accountStorageKey, &accountInfo); err != nil {
		s.log.Errorf("Couldn't retrieve account from storage: %s", err)

		return ErrAccountStorageRetrieval
	}

	// TODO(cdamian): Remove this check when above TODO is taken care of
	if accountInfo.Data.Free == types.NewU128(*big.NewInt(0)) {
		s.log.Errorf("Invalid account")

		return ErrInvalidAccount
	}

	return nil
}

func (s *service) GetLastKeyByPurpose(ctx context.Context, accountID *types.AccountID, keyPurpose types.KeyPurpose) (*types.Hash, error) {
	acc, err := s.configService.GetAccount(accountID.ToBytes())

	if err != nil {
		s.log.Errorf("Couldn't retrieve account: %s", err)

		return nil, ErrAccountRetrieval
	}

	ctx = contextutil.WithAccount(ctx, acc)

	key, err := s.keystoreAPI.GetLastKeyByPurpose(ctx, keyPurpose)

	if err != nil {
		s.log.Errorf("Couldn't retrieve key: %s", err)

		return nil, ErrKeyRetrieval
	}

	return key, nil
}
