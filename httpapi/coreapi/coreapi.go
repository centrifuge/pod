package coreapi

import (
	"context"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/go-chi/chi"
)

const (
	documentIDParam      = "document_id"
	versionIDParam       = "version_id"
	jobIDParam           = "job_id"
	tokenIDParam         = "token_id"
	registryAddressParam = "registry_address"
	accountIDParam       = "account_id"

	// BootstrappedCoreService key
	BootstrappedCoreService = "coreapiservice"
)

// Register registers the core apis to the router.
// !!!IMPORTANT - this code must NOT be executed in separate go routine, because of bootstrap.NodeObjRegistry injection.
func Register(ctx context.Context, r chi.Router) error {
	// node object registry
	nodeObjReg, ok := ctx.Value(bootstrap.NodeObjRegistry).(map[string]interface{})
	if !ok {
		return errors.New("failed to get %s", bootstrap.NodeObjRegistry)
	}

	accountSrv, ok := nodeObjReg[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("failed to get %s", config.BootstrappedConfigStorage)
	}

	nftSrv, ok := nodeObjReg[bootstrap.BootstrappedInvoiceUnpaid].(nft.InvoiceUnpaid)
	if !ok {
		return errors.New("failed to get %s", bootstrap.BootstrappedInvoiceUnpaid)
	}

	docSrv, ok := nodeObjReg[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("failed to get %s", documents.BootstrappedDocumentService)
	}

	jobsSrv, ok := nodeObjReg[jobs.BootstrappedService].(jobs.Manager)
	if !ok {
		return errors.New("failed to get %s", jobs.BootstrappedService)
	}

	cs := Service{docService: docSrv, jobsService: jobsSrv, nftService: nftSrv, accountsService: accountSrv}
	nodeObjReg[BootstrappedCoreService] = cs

	h := handler{
		srv:           Service{docService: docSrv, jobsService: jobsSrv, nftService: nftSrv, accountsService: accountSrv},
		tokenRegistry: nftSrv.(documents.TokenRegistry),
	}
	r.Post("/documents", h.CreateDocument)
	r.Put("/documents/{"+documentIDParam+"}", h.UpdateDocument)
	r.Get("/documents/{"+documentIDParam+"}", h.GetDocument)
	r.Get("/documents/{"+documentIDParam+"}/versions/{"+versionIDParam+"}", h.GetDocumentVersion)
	r.Post("/documents/{"+documentIDParam+"}/proofs", h.GenerateProofs)
	r.Post("/documents/{"+documentIDParam+"}/versions/{"+versionIDParam+"}/proofs", h.GenerateProofsForVersion)
	r.Get("/jobs/{"+jobIDParam+"}", h.GetJobStatus)
	r.Post("/nfts/registries/{"+registryAddressParam+"}/mint", h.MintNFT)
	r.Post("/nfts/registries/{"+registryAddressParam+"}/tokens/{"+tokenIDParam+"}/transfer", h.TransferNFT)
	r.Get("/nfts/registries/{"+registryAddressParam+"}/tokens/{"+tokenIDParam+"}/owner", h.OwnerOfNFT)
	r.Post("/accounts/{"+accountIDParam+"}/sign", h.SignPayload)

	return nil
}
