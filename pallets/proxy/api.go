package proxy

import (
	"context"

	"github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("proxy_api")
)

const (
	ErrMetadataRetrieval          = errors.Error("couldn't retrieve metadata")
	ErrAccountIDEncoding          = errors.Error("couldn't encode account ID")
	ErrStorageKeyCreation         = errors.Error("couldn't create storage key")
	ErrProxyStorageEntryRetrieval = errors.Error("couldn't retrieve proxy storage entry")
	ErrCallCreation               = errors.Error("couldn't create call")
	ErrSubmitAndWatchExtrinsic    = errors.Error("couldn't submit and watch extrinsic")
	ErrProxiesNotFound            = errors.Error("account proxies not found")
)

const (
	PalletName = "Proxy"

	ProxyCall = PalletName + ".proxy"
	ProxyAdd  = PalletName + ".add_proxy"

	ProxiesStorageName = "Proxies"
)

//go:generate mockery --name API --structname APIMock --filename api_mock.go --inpackage

type API interface {
	AddProxy(
		ctx context.Context,
		delegate *types.AccountID,
		proxyType proxy.CentrifugeProxyType,
		delay types.U32,
		krp signature.KeyringPair,
	) error

	ProxyCall(
		ctx context.Context,
		delegator *types.AccountID,
		proxyKeyringPair signature.KeyringPair,
		forceProxyType types.Option[proxy.CentrifugeProxyType],
		proxiedCall types.Call,
	) (*centchain.ExtrinsicInfo, error)

	GetProxies(ctx context.Context, accountID *types.AccountID) (*types.ProxyStorageEntry, error)
}

type api struct {
	api centchain.API
}

func NewAPI(centAPI centchain.API) API {
	return &api{
		api: centAPI,
	}
}

func (a *api) AddProxy(
	ctx context.Context,
	delegate *types.AccountID,
	proxyType proxy.CentrifugeProxyType,
	delay types.U32,
	krp signature.KeyringPair,
) error {
	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return ErrMetadataRetrieval
	}

	call, err := types.NewCall(
		meta,
		ProxyAdd,
		delegate,
		proxyType,
		delay,
	)

	if err != nil {
		log.Errorf("Couldn't create call: %s", err)

		return ErrCallCreation
	}

	_, _, _, err = a.api.SubmitExtrinsic(ctx, meta, call, krp)

	if err != nil {
		log.Errorf("Couldn't submit extrinsic: %s", err)

		return ErrSubmitAndWatchExtrinsic
	}

	return nil
}

func (a *api) ProxyCall(
	ctx context.Context,
	delegator *types.AccountID,
	proxyKeyringPair signature.KeyringPair,
	forceProxyType types.Option[proxy.CentrifugeProxyType],
	proxiedCall types.Call,
) (*centchain.ExtrinsicInfo, error) {
	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	call, err := types.NewCall(
		meta,
		ProxyCall,
		delegator,
		forceProxyType,
		proxiedCall,
	)

	if err != nil {
		log.Errorf("Couldn't create call: %s", err)

		return nil, ErrCallCreation
	}

	extInfo, err := a.api.SubmitAndWatch(ctx, meta, call, proxyKeyringPair)

	if err != nil {
		log.Errorf("Couldn't submit and watch extrinsic: %s", err)

		return nil, ErrSubmitAndWatchExtrinsic
	}

	return &extInfo, nil
}

func (a *api) GetProxies(_ context.Context, accountID *types.AccountID) (*types.ProxyStorageEntry, error) {
	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedAccountID, err := codec.Encode(accountID)

	if err != nil {
		log.Errorf("Couldn't encode account ID: %s", err)

		return nil, ErrAccountIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, PalletName, ProxiesStorageName, encodedAccountID)

	if err != nil {
		log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var proxyStorageEntry types.ProxyStorageEntry

	ok, err := a.api.GetStorageLatest(storageKey, &proxyStorageEntry)

	if err != nil {
		log.Errorf("Couldn't retrieve proxy storage entry from storage: %s", err)

		return nil, ErrProxyStorageEntryRetrieval
	}

	if !ok {
		log.Error("Account proxies not found")

		return nil, ErrProxiesNotFound
	}

	return &proxyStorageEntry, nil
}
