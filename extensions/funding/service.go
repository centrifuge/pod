package funding

import (
	"context"
	"reflect"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

// Service defines specific functions for extension funding
type Service interface {
	// Sign adds a signature to an existing document
	Sign(ctx context.Context, fundingID string, identifier []byte) (documents.Model, error)

	// CreateFundingAgreement creates a new funding agreement and anchors the document.
	CreateFundingAgreement(ctx context.Context, docID []byte, data *Data) (documents.Model, jobs.JobID, error)

	// UpdateFundingAgreement updates a given funding agreement with the data passed.
	UpdateFundingAgreement(ctx context.Context, docID, fundingID []byte, data *Data) (documents.Model, jobs.JobID, error)

	// GetDataAndSignatures return the funding Data and Signatures associated with the FundingID or funding index.
	GetDataAndSignatures(ctx context.Context, model documents.Model, fundingID string, idx string) (Data, []Signature, error)

	// SignFundingAgreement adds the signature to the given funding agreement.
	SignFundingAgreement(ctx context.Context, docID, fundingID []byte) (documents.Model, jobs.JobID, error)
}

// service implements Service and handles all funding related persistence and validations
type service struct {
	docSrv        documents.Service
	tokenRegistry documents.TokenRegistry
	idSrv         identity.Service
}

var log = logging.Logger("funding_agreement")

const (
	// AttrFundingLabel is the funding agreement label
	AttrFundingLabel          = "funding_agreement"
	fundingFieldKey           = "funding_agreement[{IDX}]."
	agreementIDLabel          = "agreement_id"
	fundingSignatures         = "signatures"
	fundingSignaturesFieldKey = "signatures[{IDX}]"
)

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	tokenRegistry documents.TokenRegistry,
) Service {
	return service{
		docSrv:        srv,
		tokenRegistry: tokenRegistry,
	}
}

// TODO: Move to attribute utils
func (s service) findFunding(model documents.Model, fundingID string) (data Data, err error) {
	idx, err := extensions.FindAttributeSetIDX(model, fundingID, AttrFundingLabel, agreementIDLabel, fundingFieldKey)
	if err != nil {
		return data, err
	}
	return s.deriveFundingData(model, idx)
}

func (s service) deriveFundingData(model documents.Model, idx string) (data Data, err error) {
	d := new(Data)
	types := reflect.TypeOf(*d)
	for i := 0; i < types.NumField(); i++ {
		// generate attr key
		jsonKey := types.Field(i).Tag.Get("json")
		label := extensions.LabelFromJSONTag(idx, jsonKey, fundingFieldKey)

		attrKey, err := documents.AttrKeyFromLabel(label)
		if err != nil {
			return data, err
		}

		if model.AttributeExists(attrKey) {
			attr, err := model.GetAttribute(attrKey)
			if err != nil {
				return data, err
			}

			// set field in data
			n := types.Field(i).Name

			v, err := attr.Value.String()
			if err != nil {
				return data, err
			}

			reflect.ValueOf(d).Elem().FieldByName(n).SetString(v)
		}
	}
	return *d, nil
}

// CreateFundingAgreement creates a funding agreement and anchors the document update
func (s service) CreateFundingAgreement(ctx context.Context, docID []byte, data *Data) (documents.Model, jobs.JobID, error) {
	model, err := s.docSrv.GetCurrentVersion(ctx, docID)
	if err != nil {
		return nil, jobs.NilJobID(), documents.ErrDocumentNotFound
	}

	data.AgreementID = extensions.NewAttributeSetID()
	attributes, err := extensions.CreateAttributesList(model, *data, fundingFieldKey, AttrFundingLabel)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	var collabs []identity.DID
	for _, id := range []string{data.BorrowerID, data.FunderID} {
		did, err := identity.NewDIDFromString(id)
		if err != nil {
			return nil, jobs.NilJobID(), err
		}

		collabs = append(collabs, did)
	}

	err = model.AddAttributes(
		documents.CollaboratorsAccess{
			ReadWriteCollaborators: collabs,
		},
		true,
		attributes...,
	)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	model, jobID, _, err := s.docSrv.Update(ctx, model)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return model, jobID, nil
}

// UpdateFundingAgreement updates a given funding agreement with the data passed.
func (s service) UpdateFundingAgreement(ctx context.Context, docID, fundingID []byte, data *Data) (documents.Model, jobs.JobID, error) {
	model, err := s.docSrv.GetCurrentVersion(ctx, docID)
	if err != nil {
		log.Error(err)
		return nil, jobs.NilJobID(), documents.ErrDocumentNotFound
	}

	data.AgreementID = hexutil.Encode(fundingID)
	idx, err := extensions.FindAttributeSetIDX(model, data.AgreementID, AttrFundingLabel, agreementIDLabel, fundingFieldKey)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	var collabs []identity.DID
	for _, id := range []string{data.BorrowerID, data.FunderID} {
		did, err := identity.NewDIDFromString(id)
		if err != nil {
			return nil, jobs.NilJobID(), err
		}

		collabs = append(collabs, did)
	}

	// overwriting is not enough because it is not required that
	// the funding payload contains all funding attributes
	model, err = extensions.DeleteAttributesSet(model, Data{}, idx, fundingFieldKey)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	attributes, err := extensions.FillAttributeList(*data, idx, fundingFieldKey)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	err = model.AddAttributes(
		documents.CollaboratorsAccess{
			ReadWriteCollaborators: collabs,
		},
		true,
		attributes...,
	)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	model, jobID, _, err := s.docSrv.Update(ctx, model)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return model, jobID, nil
}

// SignFundingAgreement adds the signature to the given funding agreement.
func (s service) SignFundingAgreement(ctx context.Context, docID, fundingID []byte) (documents.Model, jobs.JobID, error) {
	m, err := s.Sign(ctx, hexutil.Encode(fundingID), docID)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	m, jobID, _, err := s.docSrv.Update(ctx, m)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return m, jobID, nil
}

// GetDataAndSignatures return the funding Data and Signatures associated with the FundingID.
func (s service) GetDataAndSignatures(ctx context.Context, model documents.Model, fundingID string, idx string) (data Data, sigs []Signature, err error) {
	if idx == "" {
		idx, err = extensions.FindAttributeSetIDX(model, fundingID, AttrFundingLabel, agreementIDLabel, fundingFieldKey)
		if err != nil {
			return data, sigs, err
		}
	}

	data, err = s.deriveFundingData(model, idx)
	if err != nil {
		return data, sigs, err
	}

	sigs, err = s.deriveFundingSignatures(ctx, model, data, idx)
	if err != nil {
		return data, sigs, errors.NewTypedError(extensions.ErrAttrSetSignature, err)
	}

	return data, sigs, nil
}
