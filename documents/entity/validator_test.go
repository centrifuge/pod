//go:build unit

package entity

import (
	"testing"

	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/stretchr/testify/assert"
)

func TestFieldValidator_Validate(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	fv := fieldValidator(nil)

	err := fv.Validate(nil, nil)
	assert.ErrorIs(t, err, documents.ErrDocumentNil)

	documentMock := documents.NewDocumentMock(t)

	err = fv.Validate(documentMock, nil)
	assert.ErrorIs(t, err, documents.ErrDocumentNil)

	entity := &Entity{}

	err = fv.Validate(nil, entity)
	assert.ErrorIs(t, err, ErrEntityDataNoIdentity)

	fv = fieldValidator(identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entity.Data = Data{Identity: accountID}

	identityServiceMock.On("ValidateAccount", accountID).
		Return(errors.New("error")).
		Once()

	err = fv.Validate(nil, entity)
	assert.ErrorIs(t, err, documents.ErrIdentityInvalid)

	identityServiceMock.On("ValidateAccount", accountID).
		Return(nil).
		Once()

	err = fv.Validate(nil, entity)
	assert.Nil(t, err)
}
