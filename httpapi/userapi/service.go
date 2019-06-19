package userapi

import (
	"context"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
)

// Service provides functionality for User APIs.
type Service struct {
	transferDetailsService transferdetails.Service
}

func (s Service) CreateTransferDetails(ctx context.Context, req transferdetails.CreateTransferDetailRequest) (documents.Model, error) {
	return s.transferDetailsService.DeriveFromPayload(ctx, req)
}

func (s Service) UpdateTransferDetails(ctx context.Context, req transferdetails.UpdateTransferDetailRequest) (documents.Model, error) {
	return s.transferDetailsService.DeriveFromUpdatePayload(ctx, req)
}

func (s Service) GetTransferDetails(ctx context.Context, model documents.Model, transferID string) (*transferdetails.TransferDetail, error){
	return s.transferDetailsService.DeriveTransferResponse(ctx, model, transferID)
}

func (s Service) GetTransferDetailsList(ctx context.Context, model documents.Model) (*transferdetails.TransferDetailList, error) {
	return s.transferDetailsService.DeriveTransferListResponse(ctx, model)
}
