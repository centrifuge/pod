package coreapi

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/go-chi/chi"
)

const (
	// DocumentIDParam for document_id in api path.
	DocumentIDParam = "document_id"

	// VersionIDParam for version_id in api path.
	VersionIDParam = "version_id"

	jobIDParam           = "job_id"
	tokenIDParam         = "token_id"
	registryAddressParam = "registry_address"
	// AccountIDParam for accounts api
	AccountIDParam = "account_id"
)

// Register registers the core apis to the router.
func Register(ctx map[string]interface{}, r chi.Router) {
	coreAPISrv := ctx[BootstrappedCoreAPIService].(Service)
	tokenRegistry := ctx[bootstrap.BootstrappedNFTService].(documents.TokenRegistry)
	h := handler{
		srv:           coreAPISrv,
		tokenRegistry: tokenRegistry,
	}

	r.Post("/documents", h.CreateDocument)
	r.Put("/documents/{"+DocumentIDParam+"}", h.UpdateDocument)
	r.Get("/documents/{"+DocumentIDParam+"}", h.GetDocument)
	r.Get("/documents/{"+DocumentIDParam+"}/versions/{"+VersionIDParam+"}", h.GetDocumentVersion)
	r.Post("/documents/{"+DocumentIDParam+"}/proofs", h.GenerateProofs)
	r.Post("/documents/{"+DocumentIDParam+"}/versions/{"+VersionIDParam+"}/proofs", h.GenerateProofsForVersion)
	r.Get("/jobs/{"+jobIDParam+"}", h.GetJobStatus)
	r.Post("/nfts/registries/{"+registryAddressParam+"}/mint", h.MintNFT)
	r.Post("/nfts/registries/{"+registryAddressParam+"}/tokens/{"+tokenIDParam+"}/transfer", h.TransferNFT)
	r.Get("/nfts/registries/{"+registryAddressParam+"}/tokens/{"+tokenIDParam+"}/owner", h.OwnerOfNFT)
	r.Get("/accounts/{"+AccountIDParam+"}", h.GetAccount)
	r.Get("/accounts", h.GetAccounts)
	r.Post("/accounts", h.CreateAccount)
	r.Put("/accounts/{"+AccountIDParam+"}", h.UpdateAccount)
}
