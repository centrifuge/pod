// +build unit

package entityrelationship

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateValidator(t *testing.T) {
	cv := CreateValidator(nil)
	assert.Len(t, cv, 1)
}

func TestUpdateValidator(t *testing.T) {
	uv := UpdateValidator(nil, nil)
	assert.Len(t, uv, 2)
}

func TestFieldValidator_Validate(t *testing.T) {
	fv := fieldValidator(nil)

	//  nil error
	err := fv.Validate(nil, nil)
	assert.Error(t, err)
	errs := errors.GetErrs(err)
	assert.Len(t, errs, 1, "errors length must be one")
	assert.Contains(t, errs[0].Error(), "no(nil) document provided")

	// unknown type
	err = fv.Validate(nil, &mockModel{})
	assert.Error(t, err)
	errs = errors.GetErrs(err)
	assert.Len(t, errs, 1, "errors length must be one")
	assert.Contains(t, errs[0].Error(), "document is of invalid type")

	// identity not created from the identity factory
	idFactory := new(identity.MockFactory)
	relationship := CreateRelationship(t, testingconfig.CreateAccountContext(t, cfg))
	idFactory.On("IdentityExists", *relationship.Data.OwnerIdentity).Return(false, nil).Once()
	fv = fieldValidator(idFactory)
	err = fv.Validate(nil, relationship)
	assert.Error(t, err)
	idFactory.AssertExpectations(t)

	// identity created from identity factory
	idFactory = new(identity.MockFactory)
	idFactory.On("IdentityExists", *relationship.Data.TargetIdentity).Return(true, nil).Once()
	idFactory.On("IdentityExists", *relationship.Data.OwnerIdentity).Return(true, nil).Once()
	fv = fieldValidator(idFactory)
	err = fv.Validate(nil, relationship)
	assert.NoError(t, err)
	idFactory.AssertExpectations(t)
}
