package v2

import (
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/go-chi/chi"
	logging "github.com/ipfs/go-log"
)

// handler implements the API handlers.
type handler struct {
	srv Service
}

var log = logging.Logger("v2_api")

// Register registers the core apis to the router.
func Register(ctx map[string]interface{}, r chi.Router) {
	srv := ctx[BootstrappedService].(Service)
	h := handler{srv: srv}

	r.Post("/documents", h.CreateDocument)
	r.Post("/documents/clone", h.CloneDocument)
	r.Patch("/documents/{"+coreapi.DocumentIDParam+"}", h.UpdateDocument)
	r.Post("/documents/{"+coreapi.DocumentIDParam+"}/commit", h.Commit)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/pending", h.GetPendingDocument)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/committed", h.GetCommittedDocument)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/versions/{"+coreapi.VersionIDParam+"}", h.GetDocumentVersion)
	r.Post("/documents/{"+coreapi.DocumentIDParam+"}/signed_attribute", h.AddSignedAttribute)
	r.Delete("/documents/{"+coreapi.DocumentIDParam+"}/collaborators", h.RemoveCollaborators)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/roles/{"+RoleIDParam+"}", h.GetRole)
	r.Post("/documents/{"+coreapi.DocumentIDParam+"}/roles", h.AddRole)
	r.Patch("/documents/{"+coreapi.DocumentIDParam+"}/roles/{"+RoleIDParam+"}", h.UpdateRole)
	r.Post("/documents/{"+coreapi.DocumentIDParam+"}/transition_rules", h.AddTransitionRules)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/transition_rules/{"+RuleIDParam+"}", h.GetTransitionRule)
	r.Delete("/documents/{"+coreapi.DocumentIDParam+"}/transition_rules/{"+RuleIDParam+"}", h.DeleteTransitionRule)
}
