package v2

import (
	"context"
	"encoding/gob"
	"errors"
	"net/url"
	"time"

	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	podErrors "github.com/centrifuge/go-centrifuge/errors"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/validation"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	logging "github.com/ipfs/go-log"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/protocol"
)

func init() {
	gob.Register(&configstore.Account{})
	gob.Register([]*keystoreType.AddKey{{}})
}

type CreateIdentityRequest struct {
	Identity         *types.AccountID
	WebhookURL       string
	PrecommitEnabled bool
}

//go:generate mockery --name Service --structname ServiceMock --filename service_mock.go --inpackage

type Service interface {
	CreateIdentity(ctx context.Context, req *CreateIdentityRequest) (config.Account, error)

	ValidateKey(ctx context.Context, accountID *types.AccountID, pubKey []byte, keyPurpose keystoreType.KeyPurpose, validationTime time.Time) error
	ValidateSignature(ctx context.Context, accountID *types.AccountID, pubKey []byte, signature []byte, message []byte, validationTime time.Time) error
	ValidateAccount(ctx context.Context, accountID *types.AccountID) error

	GetLastKeyByPurpose(ctx context.Context, accountID *types.AccountID, keyPurpose keystoreType.KeyPurpose) (*types.Hash, error)
}

type service struct {
	configService        config.Service
	centAPI              centchain.API
	keystoreAPI          keystore.API
	proxyAPI             proxy.API
	protocolIDDispatcher dispatcher.Dispatcher[protocol.ID]
	log                  *logging.ZapEventLogger
}

func NewService(
	configService config.Service,
	centAPI centchain.API,
	keystoreAPI keystore.API,
	proxyAPI proxy.API,
	protocolIDDispatcher dispatcher.Dispatcher[protocol.ID],
) Service {
	log := logging.Logger("identity-service-v2")

	return &service{
		configService,
		centAPI,
		keystoreAPI,
		proxyAPI,
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

func (s *service) ValidateKey(
	ctx context.Context,
	accountID *types.AccountID,
	pubKey []byte,
	keyPurpose keystoreType.KeyPurpose,
	validationTime time.Time,
) error {
	err := validation.Validate(
		validation.NewValidator(accountID, accountIDValidatorFn),
		validation.NewValidator(pubKey, publicKeyValidatorFn),
	)

	if err != nil {
		s.log.Errorf("Invalid args: %s", err)

		return err
	}

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

	return s.validateKey(key, validationTime)
}

func (s *service) validateKey(key *keystoreType.Key, validationTime time.Time) error {
	ok, revokedAt := key.RevokedAt.Unwrap()

	if !ok {
		return nil
	}

	blockHash, err := s.centAPI.GetBlockHash(uint64(revokedAt))

	if err != nil {
		s.log.Errorf("Couldn't retrieve block hash: %s", err)

		return ErrBlockHashRetrieval
	}

	block, err := s.centAPI.GetBlock(blockHash)

	if err != nil {
		s.log.Errorf("Couldn't retrieve block: %s", err)

		return ErrBlockRetrieval
	}

	meta, err := s.centAPI.GetMetadataLatest()

	if err != nil {
		s.log.Errorf("Couldn't retrieve metadata: %s", err)

		return podErrors.ErrMetadataRetrieval
	}

	timestamp, err := getBlockTimestamp(meta, block)

	if err != nil {
		s.log.Errorf("Couldn't retrieve metadata: %s", err)

		return ErrBlockTimestampRetrieval
	}

	if validationTime.After(*timestamp) {
		s.log.Error("Key is revoked")

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
	validationTime time.Time,
) error {
	if err := s.ValidateKey(ctx, accountID, pubKey, keystoreType.KeyPurposeP2PDocumentSigning, validationTime); err != nil {
		s.log.Errorf("Couldn't validate key: %s", err)

		return err
	}

	if !crypto.VerifyMessage(pubKey, message, signature, crypto.CurveEd25519) {
		s.log.Error("Couldn't verify message - invalid signature")

		return ErrInvalidSignature
	}

	return nil
}

func (s *service) ValidateAccount(ctx context.Context, accountID *types.AccountID) error {
	if err := validation.Validate(validation.NewValidator(accountID, accountIDValidatorFn)); err != nil {
		s.log.Errorf("Invalid account ID: %s", err)

		return err
	}

	meta, err := s.centAPI.GetMetadataLatest()

	if err != nil {
		s.log.Errorf("Couldn't get latest metadata: %s", err)

		return ErrMetadataRetrieval
	}

	accountStorageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())

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

	if ok {
		return nil
	}

	// Account info not found, check if account is a valid anonymous proxy i.e. has at least one proxy with type Any.
	return s.isValidAnonymousProxy(ctx, accountID)
}

func (s *service) isValidAnonymousProxy(ctx context.Context, accountID *types.AccountID) error {
	res, err := s.proxyAPI.GetProxies(ctx, accountID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve account proxies: %s", err)

		if errors.Is(err, proxy.ErrProxiesNotFound) {
			return ErrAccountNotAnonymousProxy
		}

		return ErrAccountProxiesRetrieval
	}

	for _, proxyDefinition := range res.ProxyDefinitions {
		if uint8(proxyDefinition.ProxyType) == uint8(proxyType.Any) {
			return nil
		}
	}

	return ErrAccountNotAnonymousProxy
}

func (s *service) GetLastKeyByPurpose(ctx context.Context, accountID *types.AccountID, keyPurpose keystoreType.KeyPurpose) (*types.Hash, error) {
	if err := validation.Validate(validation.NewValidator(accountID, accountIDValidatorFn)); err != nil {
		s.log.Errorf("Invalid account ID: %s", err)

		return nil, err
	}

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
	if err := s.ValidateAccount(ctx, req.Identity); err != nil {
		s.log.Errorf("Invalid identity - %s: %s", req.Identity.ToHexString(), err)

		return ErrInvalidAccount
	}

	if req.WebhookURL == "" {
		return nil
	}

	if _, err := url.ParseRequestURI(req.WebhookURL); err != nil {
		s.log.Errorf("Invalid webhook URL: %s", err)

		return ErrInvalidWebhookURL
	}

	return nil
}

func getBlockTimestamp(meta *types.Metadata, block *types.SignedBlock) (*time.Time, error) {
	for _, extrinsic := range block.Block.Extrinsics {
		if isTimestampSetExtrinsic(meta, extrinsic.Method.CallIndex) {
			var timestamp types.UCompact

			if err := codec.Decode(extrinsic.Method.Args, &timestamp); err != nil {
				return nil, err
			}

			t := time.UnixMilli(timestamp.Int64())

			return &t, nil
		}
	}

	return nil, errors.New("timestamp extrinsic not found")
}

const (
	timestampPalletName      = "pallet_timestamp"
	timestampPalletSetMethod = "set"
)

func isTimestampSetExtrinsic(meta *types.Metadata, callIndex types.CallIndex) bool {
	for _, pallet := range meta.AsMetadataV14.Pallets {
		if pallet.Index != types.U8(callIndex.SectionIndex) {
			continue
		}

		metaType := meta.AsMetadataV14.EfficientLookup[pallet.Calls.Type.Int64()]

		if metaType.Path[0] != timestampPalletName {
			continue
		}

		for _, variant := range metaType.Def.Variant.Variants {
			if variant.Index != types.U8(callIndex.MethodIndex) {
				continue
			}

			if variant.Name == timestampPalletSetMethod {
				return true
			}
		}
	}

	return false
}
