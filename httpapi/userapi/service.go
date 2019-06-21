package userapi

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/jobs"
)

// Service provides functionality for User APIs.
type Service struct {
	docSrv                 documents.Service
	transferDetailsService transferdetails.Service
}

// TODO: this can be refactored into a generic Service which handles all kinds of custom attributes

// CreateTransferDetail creates and anchors a Transfer Detail
func (s Service) CreateTransferDetail(ctx context.Context, req transferdetails.CreateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	fmt.Println(ctx, req)
	model, jobID, err := s.transferDetailsService.CreateTransferDetail(ctx, req)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}
	return model, jobID, nil
}

// UpdateTransferDetail updates and anchors a Transfer Detail
func (s Service) UpdateTransferDetail(ctx context.Context, req transferdetails.UpdateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	model, jobID, err := s.transferDetailsService.UpdateTransferDetail(ctx, req)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return model, jobID, nil
}

// GetCurrentTransferDetail returns the current version on a Transfer Detail
func (s Service) GetCurrentTransferDetail(ctx context.Context, docID, transferID []byte) (*transferdetails.TransferDetail, documents.Model, error) {
	model, err := s.docSrv.GetCurrentVersion(ctx, docID)
	if err != nil {
		return nil, nil, err
	}

	data, model, err := s.transferDetailsService.DeriveTransferDetail(ctx, model, transferID)
	if err != nil {
		return nil, nil, err
	}

	return data, model, nil
}

// GetCurrentTransferDetailsList returns a list of Transfer Details on the current version of a document
func (s Service) GetCurrentTransferDetailsList(ctx context.Context, docID []byte) (*transferdetails.TransferDetailList, documents.Model, error) {
	model, err := s.docSrv.GetCurrentVersion(ctx, docID)
	if err != nil {
		return nil, nil, err
	}

	data, model, err := s.transferDetailsService.DeriveTransferList(ctx, model)
	if err != nil {
		return nil, nil, err
	}

	return data, model, nil
}

// GetVersionTransferDetail returns a Transfer Detail on a particular version of a Document
func (s Service) GetVersionTransferDetail(ctx context.Context, docID, versionID, transferID []byte) (*transferdetails.TransferDetail, documents.Model, error) {
	model, err := s.docSrv.GetVersion(ctx, docID, versionID)
	if err != nil {
		return nil, nil, err
	}

	data, model, err := s.transferDetailsService.DeriveTransferDetail(ctx, model, transferID)
	if err != nil {
		return nil, nil, err
	}

	return data, model, nil
}

// GetVersionTransferDetailsList returns a list of Transfer Details on a particular version of a Document
func (s Service) GetVersionTransferDetailsList(ctx context.Context, docID, versionID []byte) (*transferdetails.TransferDetailList, documents.Model, error) {
	model, err := s.docSrv.GetVersion(ctx, docID, versionID)
	if err != nil {
		return nil, nil, err
	}

	data, model, err := s.transferDetailsService.DeriveTransferList(ctx, model)
	if err != nil {
		return nil, nil, err
	}

	return data, model, nil
}
