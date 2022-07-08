package v2

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

type Service interface {
	ValidateKey(ctx context.Context, accountID *types.AccountID, key *types.Hash, keyPurpose *types.KeyPurpose) error
}

type service struct {
	accountSrv config.Service
	api        keystore.API
	log        *logging.ZapEventLogger
}

func NewService(api keystore.API, accountSrv config.Service) Service {
	log := logging.Logger("identity-service-v2")

	return &service{
		accountSrv,
		api,
		log,
	}
}

func (s *service) ValidateKey(
	ctx context.Context,
	accountID *types.AccountID,
	keyHash *types.Hash,
	keyPurpose *types.KeyPurpose,
) error {
	// TODO(cdamian): Add validation from the NFT branch

	keyID := &types.KeyID{
		Hash:       *keyHash,
		KeyPurpose: *keyPurpose,
	}

	account, err := s.accountSrv.GetAccount(accountID.ToBytes())

	if err != nil {
		s.log.Errorf("Couldn't retrieve account: %s", err)

		return ErrAccountRetrieval
	}

	ctx = contextutil.WithAccount(ctx, account)

	key, err := s.api.GetKey(ctx, keyID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve key: %s", err)

		return ErrKeyRetrieval
	}

	if key == nil {
		s.log.Errorf("Key not found")

		return ErrKeyNotFound
	}

	if key.RevokedAt.IsSome() {
		s.log.Errorf("Key is revoked")

		return ErrKeyRevoked
	}

	return nil
}
