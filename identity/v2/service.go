package v2

import (
	"context"
	"math/big"

	"github.com/centrifuge/go-centrifuge/centchain"

	"github.com/centrifuge/go-centrifuge/crypto"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

type Service interface {
	ValidateKey(ctx context.Context, accountID *types.AccountID, pubKey []byte, keyPurpose types.KeyPurpose) error
	ValidateSignature(ctx context.Context, accountID *types.AccountID, pubKey []byte, signature []byte, message []byte) error
	ValidateIdentity(ctx context.Context, accountID *types.AccountID) error

	GetLastKeyByPurpose(ctx context.Context, accountID *types.AccountID, keyPurpose types.KeyPurpose) (*types.Hash, error)
}

type service struct {
	accountSrv  config.Service
	centAPI     centchain.API
	keystoreAPI keystore.API
	log         *logging.ZapEventLogger
}

func NewService(
	accountSrv config.Service,
	centAPI centchain.API,
	keystoreAPI keystore.API,
) Service {
	log := logging.Logger("identity-service-v2")

	return &service{
		accountSrv,
		centAPI,
		keystoreAPI,
		log,
	}
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

	account, err := s.accountSrv.GetAccount(accountID.ToBytes())

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

func (s *service) ValidateIdentity(_ context.Context, accountID *types.AccountID) error {
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
	acc, err := s.accountSrv.GetAccount(accountID.ToBytes())

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
