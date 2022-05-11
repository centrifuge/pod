package v3

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"go.uber.org/multierr"
)

const (
	ErrRequestNil        = errors.Error("request nil")
	ErrMissingDocumentID = errors.Error("missing document ID")
	ErrInvalidClassID    = errors.Error("invalid class ID")
	ErrInvalidInstanceID = errors.Error("invalid instance ID")
	ErrMissingMetadata   = errors.Error("missing metadata")
	ErrMetadataTooBig    = errors.Error("metadata too big")
)

type validator struct {
	errs []error
}

func newValidator() *validator {
	return &validator{}
}

func (v *validator) error() error {
	return multierr.Combine(v.errs...)
}

func (v *validator) addError(err error) *validator {
	v.errs = append(v.errs, err)

	return v
}

func (v *validator) validateBool(expr bool, err error) *validator {
	if expr {
		v.addError(err)
	}

	return v
}

func (v *validator) validateMintRequest(req *MintNFTRequest) *validator {
	if req == nil {
		return v.addError(ErrRequestNil)
	}

	return v.validateBool(len(req.DocumentID) == 0, ErrMissingDocumentID).
		validateClassID(req.ClassID)
}

func (v *validator) validateOwnerOfRequest(req *OwnerOfRequest) *validator {
	if req == nil {
		return v.addError(ErrRequestNil)
	}

	return v.validateClassID(req.ClassID).
		validateInstanceID(req.InstanceID)
}

func (v *validator) validateCreateNFTClassRequest(req *CreateNFTClassRequest) *validator {
	if req == nil {
		return v.addError(ErrRequestNil)
	}

	return v.validateClassID(req.ClassID)
}

func (v *validator) validateInstanceMetadataOfRequest(req *InstanceMetadataOfRequest) *validator {
	if req == nil {
		return v.addError(ErrRequestNil)
	}

	return v.validateClassID(req.ClassID).
		validateInstanceID(req.InstanceID)
}

func (v *validator) validateInstanceID(instanceID types.U128) *validator {
	return v.validateBool(instanceID.BitLen() == 0, ErrInvalidInstanceID)
}

func (v *validator) validateClassID(classID types.U64) *validator {
	return v.validateBool(classID == 0, ErrInvalidClassID)
}

func (v *validator) validateMetadata(data []byte) *validator {
	return v.validateBool(len(data) > StringLimit, ErrMetadataTooBig).
		validateBool(len(data) == 0, ErrMissingMetadata)
}
