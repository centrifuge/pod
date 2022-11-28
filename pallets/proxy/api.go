package proxy

import (
	"context"

	"github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/validation"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("proxy_api")
)

const (
	ErrAccountIDEncoding          = errors.Error("couldn't encode account ID")
	ErrProxyStorageEntryRetrieval = errors.Error("couldn't retrieve proxy storage entry")
	ErrProxiesNotFound            = errors.Error("account proxies not found")
	ErrMultiAddressCreation       = errors.Error("couldn't create multi address")
)

const (
	PalletName = "Proxy"

	ProxyCall      = PalletName + ".proxy"
	ProxyAdd       = PalletName + ".add_proxy"
	ProxyAnonymous = PalletName + ".anonymous"

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
		forcedProxyType types.Option[proxy.CentrifugeProxyType],
		proxiedCall types.Call,
	) (*centchain.ExtrinsicInfo, error)

	GetProxies(accountID *types.AccountID) (*types.ProxyStorageEntry, error)
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
	err := validation.Validate(
		validation.NewValidator(delegate, validation.AccountIDValidatorFn),
	)

	if err != nil {
		log.Errorf("Validation error: %s", err)

		return err
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return errors.ErrMetadataRetrieval
	}

	delegateMultiAddress, err := types.NewMultiAddressFromAccountID(delegate.ToBytes())

	if err != nil {
		log.Errorf("Couldn't create multi address for delegate: %s", err)

		return ErrMultiAddressCreation
	}

	call, err := types.NewCall(
		meta,
		ProxyAdd,
		delegateMultiAddress,
		proxyType,
		delay,
	)

	if err != nil {
		log.Errorf("Couldn't create call: %s", err)

		return errors.ErrCallCreation
	}

	_, _, _, err = a.api.SubmitExtrinsic(ctx, meta, call, krp)

	if err != nil {
		log.Errorf("Couldn't submit extrinsic: %s", err)

		return errors.ErrExtrinsicSubmission
	}

	return nil
}

func (a *api) ProxyCall(
	ctx context.Context,
	delegator *types.AccountID,
	proxyKeyringPair signature.KeyringPair,
	forcedProxyType types.Option[proxy.CentrifugeProxyType],
	proxiedCall types.Call,
) (*centchain.ExtrinsicInfo, error) {
	err := validation.Validate(
		validation.NewValidator(delegator, validation.AccountIDValidatorFn),
	)

	if err != nil {
		log.Errorf("Validation error: %s", err)

		return nil, err
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	delegatorMultiAddress, err := types.NewMultiAddressFromAccountID(delegator.ToBytes())

	if err != nil {
		log.Errorf("Couldn't create multi address for delegator: %s", err)

		return nil, ErrMultiAddressCreation
	}

	call, err := types.NewCall(
		meta,
		ProxyCall,
		delegatorMultiAddress,
		forcedProxyType,
		proxiedCall,
	)

	if err != nil {
		log.Errorf("Couldn't create call: %s", err)

		return nil, errors.ErrCallCreation
	}

	extInfo, err := a.api.SubmitAndWatch(ctx, meta, call, proxyKeyringPair)

	if err != nil {
		log.Errorf("Couldn't submit and watch extrinsic: %s", err)

		return nil, errors.ErrExtrinsicSubmitAndWatch
	}

	return &extInfo, nil
}

func (a *api) GetProxies(accountID *types.AccountID) (*types.ProxyStorageEntry, error) {
	err := validation.Validate(
		validation.NewValidator(accountID, validation.AccountIDValidatorFn),
	)

	if err != nil {
		log.Errorf("Validation error: %s", err)

		return nil, err
	}

	meta, err := a.api.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	encodedAccountID, err := codec.Encode(accountID)

	if err != nil {
		log.Errorf("Couldn't encode account ID: %s", err)

		return nil, ErrAccountIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, PalletName, ProxiesStorageName, encodedAccountID)

	if err != nil {
		log.Errorf("Couldn't create storage key: %s", err)

		return nil, errors.ErrStorageKeyCreation
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
