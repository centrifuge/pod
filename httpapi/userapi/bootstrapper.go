package userapi

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions/funding"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
)

// BootstrappedUserAPIService key maps to the Service implementation in Bootstrap context.
const BootstrappedUserAPIService = "UserAPI Service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	coreAPISrv, ok := ctx[coreapi.BootstrappedCoreAPIService].(coreapi.Service)
	if !ok {
		return errors.New("failed to get %s", coreapi.BootstrappedCoreAPIService)
	}

	tdSrv, ok := ctx[transferdetails.BootstrappedTransferDetailService].(transferdetails.Service)
	if !ok {
		return errors.New("failed to get %s", transferdetails.BootstrappedTransferDetailService)
	}

	erSrv, ok := ctx[entityrelationship.BootstrappedEntityRelationshipService].(entityrelationship.Service)
	if !ok {
		return errors.New("failed to get %s", entityrelationship.BootstrappedEntityRelationshipService)
	}

	eSrv, ok := ctx[entity.BootstrappedEntityService].(entity.Service)
	if !ok {
		return errors.New("failed to get %s", entity.BootstrappedEntityService)
	}

	fundingSrv, ok := ctx[funding.BootstrappedFundingService].(funding.Service)
	if !ok {
		return errors.New("failed to get %s", funding.BootstrappedFundingService)
	}

	configSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("failed to get %s", config.BootstrappedConfigStorage)
	}

	ctx[BootstrappedUserAPIService] = Service{
		coreAPISrv:             coreAPISrv,
		transferDetailsService: tdSrv,
		entityRelationshipSrv:  erSrv,
		entitySrv:              eSrv,
		fundingSrv:             fundingSrv,
		config:                 configSrv,
	}
	return nil
}
