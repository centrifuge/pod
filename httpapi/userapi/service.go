package userapi

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service provides functionality for User APIs.
type Service struct {
	srv                    documents.Service
	transferDetailsService transferdetails.Service
}

// TODO: this can be refactored into a generic Service which handles all kinds of custom attributes

// CreateTransferDetailsModel
func (s Service) CreateTransferDetailsModel(ctx context.Context, req transferdetails.CreateTransferDetailRequest) (documents.Model, error) {
	model, err := s.transferDetailsService.CreateTransferDetail(ctx, req)
	if err != nil {
		return nil, err
	}
	return model, nil
}

// UpdateTransferDetailsModel
func (s Service) UpdateTransferDetailsModel(ctx context.Context, req transferdetails.UpdateTransferDetailRequest) (documents.Model, error) {
	model, err := s.transferDetailsService.UpdateTransferDetail(ctx, req)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// GetCurrentTransferDetails
func (s Service) GetCurrentTransferDetail(ctx context.Context, docID, transferID string) (*transferdetails.TransferDetail, error) {
	identifier, err := hexutil.Decode(docID)
	if err != nil {
		return nil, err
	}

	model, err := s.srv.GetCurrentVersion(ctx, identifier)
	if err != nil {
		return nil, err
	}

	return s.transferDetailsService.DeriveTransferResponse(ctx, model, transferID)
}

// GetCurrentTransferDetailsList
func (s Service) GetCurrentTransferDetailsList(ctx context.Context, docID string) (*transferdetails.TransferDetailList, error) {
	identifier, err := hexutil.Decode(docID)
	if err != nil {
		return nil, err
	}

	model, err := s.srv.GetCurrentVersion(ctx, identifier)
	if err != nil {
		return nil, err
	}

	return s.transferDetailsService.DeriveTransferListResponse(ctx, model)
}

// GetVersionTransferDetail
func (s Service) GetVersionTransferDetail(ctx context.Context, docID, versionID, transferID string) (*transferdetails.TransferDetail, error) {
	identifier, err := hexutil.Decode(docID)
	if err != nil {
		return nil, err
	}

	version, err := hexutil.Decode(versionID)
	if err != nil {
		return nil, err
	}

	model, err := s.srv.GetVersion(ctx, identifier, version)
	if err != nil {
		return nil, err
	}

	return s.transferDetailsService.DeriveTransferResponse(ctx, model, transferID)
}

// GetVersionTransferDetailsList
func (s Service) GetVersionTransferDetailsList(ctx context.Context, docID, versionID string) (*transferdetails.TransferDetailList, error) {
	identifier, err := hexutil.Decode(docID)
	if err != nil {
		return nil, err
	}

	version, err := hexutil.Decode(versionID)
	if err != nil {
		return nil, err
	}

	model, err := s.srv.GetVersion(ctx, identifier, version)
	if err != nil {
		return nil, err
	}

	return s.transferDetailsService.DeriveTransferListResponse(ctx, model)
}
