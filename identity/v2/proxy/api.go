package proxy

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/centchain"
	logging "github.com/ipfs/go-log"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	ErrContextAccountRetrieval    = errors.Error("couldn't retrieve account from context")
	ErrKeyringPairRetrieval       = errors.Error("couldn't retrieve keyring pair")
	ErrMetadataRetrieval          = errors.Error("couldn't retrieve metadata")
	ErrAccountIDEncoding          = errors.Error("couldn't encode account ID")
	ErrStorageKeyCreation         = errors.Error("couldn't create storage key")
	ErrProxyStorageEntryRetrieval = errors.Error("couldn't retrieve proxy storage entry")
)

const (
	PalletName = "Proxy"

	ProxiesStorageName = "Proxies"
)

type API interface {
	GetProxies(ctx context.Context, accountID *types.AccountID) (*types.ProxyStorageEntry, error)
}

type api struct {
	api centchain.API
	log *logging.ZapEventLogger
}

func NewAPI(centAPI centchain.API) API {
	return &api{
		api: centAPI,
		log: logging.Logger("proxy_api"),
	}
}

func (a *api) GetProxies(ctx context.Context, accountID *types.AccountID) (*types.ProxyStorageEntry, error) {
	acc, err := contextutil.Account(ctx)

	if err != nil {
		a.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrContextAccountRetrieval
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()

	if err != nil {
		a.log.Errorf("Couldn't retrieve key ring pair from account: %s", err)

		return nil, ErrKeyringPairRetrieval
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedAccountID, err := types.Encode(accountID)

	if err != nil {
		a.log.Errorf("Couldn't encode account ID: %s", err)

		return nil, ErrAccountIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, PalletName, ProxiesStorageName, krp.PublicKey, encodedAccountID)

	if err != nil {
		a.log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var proxyStorageEntry types.ProxyStorageEntry

	// TODO(cdamian): Use the OK from the NFT branch.
	if err := a.api.GetStorageLatest(storageKey, &proxyStorageEntry); err != nil {
		a.log.Errorf("Couldn't retrieve proxy storage entry from storage: %s", err)

		return nil, ErrProxyStorageEntryRetrieval
	}

	return &proxyStorageEntry, nil
}
