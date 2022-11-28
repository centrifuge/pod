//go:build unit

package entityrelationship

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/documents"

	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestFieldValidator_Validate(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	fv := fieldValidator(identityServiceMock)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	doc := &EntityRelationship{
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: utils.RandomSlice(32),
			TargetIdentity:   targetIdentity,
		},
	}

	identityServiceMock.On("ValidateAccount", ownerIdentity).
		Return(nil).Once()

	identityServiceMock.On("ValidateAccount", targetIdentity).
		Return(nil).Once()

	err = fv.Validate(nil, doc)
	assert.NoError(t, err)
}

func TestFieldValidator_Validate_NilDocument(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	fv := fieldValidator(identityServiceMock)

	err := fv.Validate(nil, nil)
	assert.ErrorIs(t, err, documents.ErrDocumentNil)
}

func TestFieldValidator_Validate_InvalidDocumentType(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	fv := fieldValidator(identityServiceMock)

	doc := documents.NewDocumentMock(t)

	err := fv.Validate(nil, doc)
	assert.ErrorIs(t, err, documents.ErrDocumentInvalidType)
}

func TestFieldValidator_Validate_InvalidOwnerIdentity(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	fv := fieldValidator(identityServiceMock)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	doc := &EntityRelationship{
		Data: Data{
			OwnerIdentity:    nil,
			EntityIdentifier: utils.RandomSlice(32),
			TargetIdentity:   targetIdentity,
		},
	}

	err = fv.Validate(nil, doc)
	assert.True(t, errors.IsOfType(documents.ErrIdentityInvalid, err))
}

func TestFieldValidator_Validate_InvalidTargetIdentity(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	fv := fieldValidator(identityServiceMock)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	doc := &EntityRelationship{
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: utils.RandomSlice(32),
			TargetIdentity:   nil,
		},
	}

	err = fv.Validate(nil, doc)
	assert.True(t, errors.IsOfType(documents.ErrIdentityInvalid, err))
}

func TestFieldValidator_Validate_AccountValidationError(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	fv := fieldValidator(identityServiceMock)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	doc := &EntityRelationship{
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: utils.RandomSlice(32),
			TargetIdentity:   targetIdentity,
		},
	}

	identityServiceMock.On("ValidateAccount", ownerIdentity).
		Return(errors.New("error")).Once()

	err = fv.Validate(nil, doc)
	assert.True(t, errors.IsOfType(documents.ErrIdentityInvalid, err))
}
