package coreapi

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
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
)

// Register registers the core apis to the router.
func Register(r chi.Router,
	nftSrv nft.Service,
	accountSrv config.Service,
	docSrv documents.Service,
	jobsSrv jobs.Manager) {
	h := handler{
		srv:           Service{docService: docSrv, jobsService: jobsSrv, nftService: nftSrv, accountsService: accountSrv},
		tokenRegistry: nftSrv.(documents.TokenRegistry),
	}
	r.Post("/documents", h.CreateDocument)
	r.Put("/documents", h.UpdateDocument)
	r.Get("/documents/{"+documentIDParam+"}", h.GetDocument)
	r.Get("/documents/{"+documentIDParam+"}/versions/{"+versionIDParam+"}", h.GetDocumentVersion)
	r.Post("/documents/{"+documentIDParam+"}/proofs", h.GenerateProofs)
	r.Post("/documents/{"+documentIDParam+"}/versions/{"+versionIDParam+"}/proofs", h.GenerateProofsForVersion)
	r.Get("/jobs/{"+jobIDParam+"}", h.GetJobStatus)
	// TODO change these paths
	r.Post("/nfts/mint", h.MintNFT)
	r.Post("/nfts/{"+tokenIDParam+"}/transfer", h.TransferNFT)
	r.Get("/nfts/{"+tokenIDParam+"}/registry/{"+registryAddressParam+"}/owner", h.OwnerOfNFT)
	r.Post("/accounts/{"+accountIDParam+"}/sign", h.SignPayload)
}
