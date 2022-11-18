package utility

import (
	"context"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("utility_api")
)

const (
	PalletName = "Utility"

	BatchAllCall = PalletName + ".batch_all"
)

type API interface {
	BatchAll(ctx context.Context, calls ...types.Call) (*centchain.ExtrinsicInfo, error)
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

func (a *api) BatchAll(ctx context.Context, calls ...types.Call) (*centchain.ExtrinsicInfo, error) {
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

	call, err := types.NewCall(
		meta,
		BatchAllCall,
		calls,
	)

	if err != nil {
		log.Errorf("Couldn't create call: %s", err)

		return nil, errors.ErrCallCreation
	}

	extInfo, err := a.proxyAPI.ProxyCall(
		ctx,
		identity,
		a.podOperator.ToKeyringPair(),
		types.NewOption(proxyType.PodOperation),
		call,
	)

	if err != nil {
		log.Errorf("Couldn't perform proxy call: %s", err)

		return nil, errors.ErrProxyCall
	}

	return extInfo, nil
}
