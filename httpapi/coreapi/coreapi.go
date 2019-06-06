package coreapi

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/go-chi/chi"
)

// Register registers the core apis to the router.
func Register(r *chi.Mux,
	registry documents.TokenRegistry,
	docSrv documents.Service,
	jobsSrv jobs.Manager) {
	h := handler{srv: Service{docService: docSrv, jobsService: jobsSrv}, tokenRegistry: registry}
	r.Post("/documents", h.CreateDocument)
	r.Put("/documents", h.UpdateDocument)
	r.Get("/documents/{document_id}", h.GetDocument)
	r.Get("/documents/{document_id}/versions/{version_id}", h.GetDocumentVersion)
	r.Get("/jobs/{job_id}", h.GetJobStatus)
}
