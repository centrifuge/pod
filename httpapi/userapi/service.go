package userapi

// Service provides functionality for User APIs.
type Service struct {
	//fundService funding.Service
	// paymentDetailsService
}

// Sign adds a signature to an existing document
//func (s Service) Sign(ctx context.Context, fundingID string, identifier []byte) (documents.Model, error) {
//	return s.fundService.Sign(ctx, fundingID, identifier)
//}
//
//// DeriveFromUpdatePayload derives Funding from clientUpdatePayload
//func (s Service) DeriveFromUpdatePayload(ctx context.Context, req *clientfunpb.FundingUpdatePayload, identifier []byte) (documents.Model, error) {
//	return s.fundService.DeriveFromUpdatePayload(ctx, req, identifier)
//}
//
//// DeriveFromPayload derives Funding from clientPayload
//func (s Service) DeriveFromPayload(ctx context.Context, req *clientfunpb.FundingCreatePayload, identifier []byte) (documents.Model, error) {
//	return s.fundService.DeriveFromPayload(ctx, req, identifier)
//}
//
//// DeriveFundingResponse returns a funding in client format
//func (s Service) DeriveFundingResponse(ctx context.Context, model documents.Model, fundingID string) (*clientfunpb.FundingResponse, error) {
//	return s.DeriveFundingResponse(ctx, model, fundingID)
//}
//
//// DeriveFundingListResponse returns a funding list in client format
//func (s Service) DeriveFundingListResponse(ctx context.Context, model documents.Model) (*clientfunpb.FundingListResponse, error) {
//	return s.DeriveFundingListResponse(ctx, model)
//}