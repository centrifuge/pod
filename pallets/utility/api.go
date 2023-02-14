package utility

import (
	"context"
	"fmt"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

const (
	ErrBatchCallCreation = errors.Error("couldn't create batch call")
)

var (
	log = logging.Logger("utility_api")
)

const (
	PalletName = "Utility"

	BatchAllCall = PalletName + ".batch_all"
)

//go:generate mockery --name API --structname APIMock --filename api_mock.go --inpackage

type API interface {
	BatchAll(ctx context.Context, callProviderFns ...centchain.CallProviderFn) (*centchain.ExtrinsicInfo, error)
}

type api struct {
	centAPI  centchain.API
	proxyAPI proxy.API

	podOperator config.PodOperator
}

func NewAPI(centAPI centchain.API, proxyAPI proxy.API, podOperator config.PodOperator) API {
	return &api{
		centAPI:     centAPI,
		proxyAPI:    proxyAPI,
		podOperator: podOperator,
	}
}

func (a *api) BatchAll(ctx context.Context, callProviderFns ...centchain.CallProviderFn) (*centchain.ExtrinsicInfo, error) {
	identity, err := contextutil.Identity(ctx)

	if err != nil {
		log.Errorf("Couldn't retrieve identity from context: %s", err)

		return nil, errors.ErrContextIdentityRetrieval
	}

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	batchCall, err := BatchCalls(callProviderFns...)(meta)

	if err != nil {
		log.Errorf("Couldn't create batch call: %s", err)

		return nil, ErrBatchCallCreation
	}

	extInfo, err := a.proxyAPI.ProxyCall(
		ctx,
		identity,
		a.podOperator.ToKeyringPair(),
		types.NewOption(proxyType.PodOperation),
		*batchCall,
	)

	if err != nil {
		log.Errorf("Couldn't perform proxy call: %s", err)

		return nil, errors.ErrProxyCall
	}

	return extInfo, nil
}

func BatchCalls(callCreationFns ...centchain.CallProviderFn) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		var calls []*types.Call

		for _, callCreationFn := range callCreationFns {
			call, err := callCreationFn(meta)

			if err != nil {
				return nil, fmt.Errorf("couldn't create call: %w", err)
			}

			calls = append(calls, call)
		}

		batchCall, err := types.NewCall(meta, BatchAllCall, calls)

		if err != nil {
			return nil, err
		}

		return &batchCall, nil
	}
}
