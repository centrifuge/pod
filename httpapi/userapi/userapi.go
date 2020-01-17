package userapi

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/go-chi/chi"
)

const (
	transferIDParam      = "transfer_id"
	agreementIDParam     = "agreement_id"
	registryAddressParam = "registry_address"
	tokenIDParam         = "token_id"
)

// Register registers the core apis to the router.
func Register(ctx map[string]interface{}, r chi.Router) {
	tokenRegistry := ctx[bootstrap.BootstrappedInvoiceUnpaid].(documents.TokenRegistry)
	userAPISrv := ctx[BootstrappedUserAPIService].(Service)
	h := handler{
		tokenRegistry: tokenRegistry,
		srv:           userAPISrv,
	}

	// transfer details api
	r.Post("/documents/{"+coreapi.DocumentIDParam+"}/transfer_details", h.CreateTransferDetail)
	r.Put("/documents/{"+coreapi.DocumentIDParam+"}/transfer_details/{"+transferIDParam+"}", h.UpdateTransferDetail)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/transfer_details", h.GetTransferDetailList)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/transfer_details/{"+transferIDParam+"}", h.GetTransferDetail)

	// entity api
	r.Post("/entities", h.CreateEntity)
	r.Put("/entities/{"+coreapi.DocumentIDParam+"}", h.UpdateEntity)
	r.Get("/entities/{"+coreapi.DocumentIDParam+"}", h.GetEntity)
	r.Post("/entities/{"+coreapi.DocumentIDParam+"}/share", h.ShareEntity)
	r.Post("/entities/{"+coreapi.DocumentIDParam+"}/revoke", h.RevokeEntity)
	r.Get("/relationships/{"+coreapi.DocumentIDParam+"}/entity", h.GetEntityThroughRelationship)

	// funding api
	r.Post("/documents/{"+coreapi.DocumentIDParam+"}/funding_agreements", h.CreateFundingAgreement)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/funding_agreements", h.GetFundingAgreements)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/funding_agreements/{"+agreementIDParam+"}", h.GetFundingAgreement)
	r.Put("/documents/{"+coreapi.DocumentIDParam+"}/funding_agreements/{"+agreementIDParam+"}", h.UpdateFundingAgreement)
	r.Post("/documents/{"+coreapi.DocumentIDParam+"}/funding_agreements/{"+agreementIDParam+"}/sign", h.SignFundingAgreement)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/versions/{"+coreapi.VersionIDParam+"}/funding_agreements/{"+agreementIDParam+"}", h.GetFundingAgreementFromVersion)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/versions/{"+coreapi.VersionIDParam+"}/funding_agreements", h.GetFundingAgreementsFromVersion)
}

// RegisterBeta registers the core apis to the router that are not production ready
func RegisterBeta(ctx map[string]interface{}, r chi.Router) {
	tokenRegistry := ctx[bootstrap.BootstrappedInvoiceUnpaid].(documents.TokenRegistry)
	userAPISrv := ctx[BootstrappedUserAPIService].(Service)
	h := handler{
		tokenRegistry: tokenRegistry,
		srv:           userAPISrv,
	}

	// beta
	r.Post("/nfts/registries/{"+registryAddressParam+"}/mint", h.MintNFT)
	r.Post("/nfts/registries/{"+registryAddressParam+"}/tokens/{"+tokenIDParam+"}/transfer", h.TransferNFT)
	r.Get("/nfts/registries/{"+registryAddressParam+"}/tokens/{"+tokenIDParam+"}/owner", h.OwnerOfNFT)
}
