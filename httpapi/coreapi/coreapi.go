package coreapi

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/go-chi/chi"
)

const (
	documentIDParam = "document_id"
	versionIDParam  = "version_id"
	jobIDParam      = "job_id"
	accountIDParam  = "account_id"
)

// Register registers the core apis to the router.
func Register(r *chi.Mux,
	registry documents.TokenRegistry,
	accountSrv config.Service,
	docSrv documents.Service,
	jobsSrv jobs.Manager) {
	h := handler{srv: Service{docService: docSrv, jobsService: jobsSrv, accountsService: accountSrv}, tokenRegistry: registry}
	r.Post("/documents", h.CreateDocument)
	r.Put("/documents", h.UpdateDocument)
	r.Get("/documents/{"+documentIDParam+"}", h.GetDocument)
	r.Get("/documents/{"+documentIDParam+"}/versions/{"+versionIDParam+"}", h.GetDocumentVersion)
	r.Post("/documents/{"+documentIDParam+"}/proofs", h.GenerateProofs)
	r.Post("/documents/{"+documentIDParam+"}/versions/{"+versionIDParam+"}/proofs", h.GenerateProofsForVersion)
	r.Get("/jobs/{"+jobIDParam+"}", h.GetJobStatus)
	r.Post("/accounts/{"+accountIDParam+"}/sign", h.SignPayload)
}
