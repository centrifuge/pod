package v2

import (
	"context"
	"encoding/gob"
	"net/url"

	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/protocol"
)

func init() {
	gob.Register(&configstore.Account{})
	gob.Register([]*keystoreType.AddKey{{}})
}

//go:generate mockery --name Service --structname ServiceMock --filename service_mock.go --inpackage

type Service interface {
	CreateIdentity(ctx context.Context, req *CreateIdentityRequest) (config.Account, error)

	ValidateKey(ctx context.Context, accountID *types.AccountID, pubKey []byte, keyPurpose keystoreType.KeyPurpose) error
	ValidateSignature(ctx context.Context, accountID *types.AccountID, pubKey []byte, signature []byte, message []byte) error
	ValidateAccount(ctx context.Context, accountID *types.AccountID) error

	GetLastKeyByPurpose(ctx context.Context, accountID *types.AccountID, keyPurpose keystoreType.KeyPurpose) (*types.Hash, error)
}

type service struct {
	configService        config.Service
	centAPI              centchain.API
	keystoreAPI          keystore.API
	protocolIDDispatcher dispatcher.Dispatcher[protocol.ID]
	log                  *logging.ZapEventLogger
}

func NewService(
	configService config.Service,
	centAPI centchain.API,
	keystoreAPI keystore.API,
	protocolIDDispatcher dispatcher.Dispatcher[protocol.ID],
) Service {
	log := logging.Logger("identity-service-v2")

	return &service{
		configService,
		centAPI,
		keystoreAPI,
		protocolIDDispatcher,
		log,
	}
}

func (s *service) CreateIdentity(ctx context.Context, req *CreateIdentityRequest) (config.Account, error) {
	if err := s.validateCreateIdentityRequest(ctx, req); err != nil {
		s.log.Errorf("Invalid request: %s", err)

		return nil, err
	}

	signingPublicKey, signingPrivateKey, err := generateDocumentSigningKeys()

	if err != nil {
		s.log.Errorf("Couldn't generate document signing key pair: %s", err)

		return nil, ErrSigningKeyPairGeneration
	}

	acc, err := configstore.NewAccount(
		req.Identity,
		signingPublicKey,
		signingPrivateKey,
		req.WebhookURL,
		req.PrecommitEnabled,
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

	return acc, nil
}

func generateDocumentSigningKeys() (libp2pcrypto.PubKey, libp2pcrypto.PrivKey, error) {
	signingPublicKey, signingPrivateKey, err := ed25519.GenerateSigningKeyPair()

	if err != nil {
		return nil, nil, err
	}

	privateKey, err := libp2pcrypto.UnmarshalEd25519PrivateKey(signingPrivateKey)

	if err != nil {
		return nil, nil, err
	}

	publicKey, err := libp2pcrypto.UnmarshalEd25519PublicKey(signingPublicKey)

	if err != nil {
		return nil, nil, err
	}

	return publicKey, privateKey, nil
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

	return nil
}

func (s *service) ValidateKey(
	ctx context.Context,
	accountID *types.AccountID,
	pubKey []byte,
	keyPurpose keystoreType.KeyPurpose,
) error {
	// TODO(cdamian): Add validation from the NFT branch

	keyID := &keystoreType.KeyID{
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

	if !key.RevokedAt.HasValue() {
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

	if types.U32(latestBlock.Block.Header.Number) >= revokedAt {
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
	if err := s.ValidateKey(ctx, accountID, pubKey, keystoreType.KeyPurposeP2PDocumentSigning); err != nil {
		s.log.Errorf("Couldn't validate key: %s", err)

		return err
	}

	if !crypto.VerifyMessage(pubKey, message, signature, crypto.CurveEd25519) {
		s.log.Error("Couldn't verify message - invalid signature")

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

	ok, err := s.centAPI.GetStorageLatest(accountStorageKey, &accountInfo)

	if err != nil {
		s.log.Errorf("Couldn't retrieve account from storage: %s", err)

		return ErrAccountStorageRetrieval
	}

	if !ok {
		s.log.Errorf("Account not found")

		return ErrAccountNotFound
	}

	return nil
}

func (s *service) GetLastKeyByPurpose(ctx context.Context, accountID *types.AccountID, keyPurpose keystoreType.KeyPurpose) (*types.Hash, error) {
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
