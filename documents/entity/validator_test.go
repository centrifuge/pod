//go:build unit

package entity

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFieldValidator_Validate(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	fv := fieldValidator(nil)

	err := fv.Validate(nil, nil)
	assert.ErrorIs(t, err, documents.ErrDocumentNil)

	documentMock := documents.NewDocumentMock(t)

	err = fv.Validate(documentMock, nil)
	assert.ErrorIs(t, err, documents.ErrDocumentInvalidType)

	entity := &Entity{}

	err = fv.Validate(entity, nil)
	assert.ErrorIs(t, err, ErrEntityDataNoIdentity)

	fv = fieldValidator(identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entity.Data = Data{Identity: accountID}

	identityServiceMock.On("ValidateAccount", mock.Anything, accountID).
		Return(errors.New("error")).
		Once()

	err = fv.Validate(entity, nil)
	assert.ErrorIs(t, err, documents.ErrIdentityInvalid)

	identityServiceMock.On("ValidateAccount", mock.Anything, accountID).
		Return(nil).
		Once()

	err = fv.Validate(entity, nil)
	assert.Nil(t, err)
}
