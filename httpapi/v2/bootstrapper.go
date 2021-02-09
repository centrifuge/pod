package v2

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/oracle"
	"github.com/centrifuge/go-centrifuge/pending"
)

// BootstrappedService key maps to the Service implementation in Bootstrap context.
const BootstrappedService = "V2 Service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	pendingDocSrv := ctx[pending.BootstrappedPendingDocumentService].(pending.Service)
	nftSrv := ctx[bootstrap.BootstrappedNFTService].(nft.Service)
	oracleService := ctx[oracle.BootstrappedOracleService].(oracle.Service)
	accountsSrv := ctx[config.BootstrappedConfigStorage].(config.Service)
	dispatcher := ctx[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
	entitySrv := ctx[entity.BootstrappedEntityService].(entity.Service)
	erSrv := ctx[entityrelationship.BootstrappedEntityRelationshipService].(entityrelationship.Service)
	ctx[BootstrappedService] = Service{
		pendingDocSrv: pendingDocSrv,
		tokenRegistry: nftSrv.(documents.TokenRegistry),
		oracleService: oracleService,
		accountSrv:    accountsSrv,
		dispatcher:    dispatcher,
		nftSrv:        nftSrv,
		entitySrv:     entitySrv,
		erSrv:         erSrv,
	}
	return nil
}
