package userapi

import (
	"context"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/funding"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
)

// Service provides functionality for User APIs.
type Service struct {
	fundService funding.Service
	transferDetailsService transferdetails.Service
}

func (s Service) CreateTransferDetails(ctx context.Context, req CreateTransferDetailRequest, identifier []byte) (documents.Model, error) {

}

func (s Service) UpdateTransferDetails(ctx context.Context, req UpdateTransferDetailRequest, identifier []byte) (documents.Model, error) {

}

func (s Service) GetTransferDetails(ctx context.Context, model documents.Model, transferID string) (*TransferDetailResponse, error){

}

func (s Service) GetTransferDetailsList(ctx context.Context, model documents.Model) (*TransferDetailListResponse, error) {

}
