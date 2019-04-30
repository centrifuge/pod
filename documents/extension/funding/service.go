package funding

import (
	"context"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
)

// Service defines specific functions for extension funding
type Service interface {
	documents.Service
	DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error)
	DeriveFundingResponse(model documents.Model, payload *clientfundingpb.FundingCreatePayload) (*clientfundingpb.FundingResponse, error)

}

// service implements Service and handles all funding related persistence and validations
type service struct {
	documents.Service
	tokenRegFinder func() documents.TokenRegistry
}

func (s service) DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error) {
	current, err := s.GetCurrentVersion(ctx, identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "document not found")
	}

	new, err := current.PrepareNewVersionWithExistingData()
	if err != nil {
		return nil, err
	}

	// todo validate funding payload


	// todo add custom attributes to model


	return new, nil
}

// DeriveInvoiceResponse returns create response from invoice model
func (s service) DeriveFundingResponse(model documents.Model, payload *clientfundingpb.FundingCreatePayload) (*clientfundingpb.FundingResponse, error) {
	h, err := documents.DeriveResponseHeader(s.tokenRegFinder(), model)
	if err != nil {
		return nil, errors.New("failed to derive response: %v", err)
	}

	return &clientfundingpb.FundingResponse{
		Header: h,
		Data:   payload.Data,
	}, nil

}

