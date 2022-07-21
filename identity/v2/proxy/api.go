package proxy

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

const (
	ErrContextAccountRetrieval    = errors.Error("couldn't retrieve account from context")
	ErrAccountProxyRetrieval      = errors.Error("couldn't retrieve account proxy")
	ErrKeyringPairRetrieval       = errors.Error("couldn't retrieve keyring pair")
	ErrMetadataRetrieval          = errors.Error("couldn't retrieve metadata")
	ErrAccountIDEncoding          = errors.Error("couldn't encode account ID")
	ErrStorageKeyCreation         = errors.Error("couldn't create storage key")
	ErrProxyStorageEntryRetrieval = errors.Error("couldn't retrieve proxy storage entry")
	ErrCallCreation               = errors.Error("couldn't create call")
	ErrSubmitAndWatchExtrinsic    = errors.Error("couldn't submit and watch extrinsic")
)

const (
	PalletName = "Proxy"

	ProxyCall = PalletName + ".proxy"

	ProxiesStorageName = "Proxies"
)

type API interface {
	GetProxies(ctx context.Context, accountID *types.AccountID) (*types.ProxyStorageEntry, error)

	ProxyCall(
		ctx context.Context,
		delegator *types.AccountID,
		accountProxy *config.AccountProxy,
		proxiedCall types.Call,
	) (*centchain.ExtrinsicInfo, error)
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

func (a *api) ProxyCall(
	ctx context.Context,
	delegator *types.AccountID,
	accountProxy *config.AccountProxy,
	proxiedCall types.Call,
) (*centchain.ExtrinsicInfo, error) {
	proxyKeyringPair, err := accountProxy.ToKeyringPair()

	if err != nil {
		a.log.Errorf("Couldn't get key ring pair for account proxy: %s", err)

		return nil, ErrKeyringPairRetrieval
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	call, err := types.NewCall(
		meta,
		ProxyCall,
		delegator,
		types.NewOptionU8Empty(),
		proxiedCall,
	)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, ErrCallCreation
	}

	extInfo, err := a.api.SubmitAndWatch(ctx, meta, call, *proxyKeyringPair)

	if err != nil {
		a.log.Errorf("Couldn't submit and watch extrinsic: %s", err)

		return nil, ErrSubmitAndWatchExtrinsic
	}

	return &extInfo, nil
}

func (a *api) GetProxies(ctx context.Context, accountID *types.AccountID) (*types.ProxyStorageEntry, error) {
	acc, err := contextutil.Account(ctx)

	if err != nil {
		a.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrContextAccountRetrieval
	}

	accProxy, err := acc.GetAccountProxies().GetDefault()

	if err != nil {
		a.log.Errorf("Couldn't get default account proxy: %s", err)

		return nil, ErrAccountProxyRetrieval
	}

	krp, err := accProxy.ToKeyringPair()

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
