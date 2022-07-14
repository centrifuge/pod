//go:build unit
// +build unit

package entity

//
//func TestFieldValidator_Validate(t *testing.T) {
//	fv := fieldValidator(nil)
//
//	//  nil error
//	err := fv.Validate(nil, nil)
//	assert.Error(t, err)
//	errs := errors.GetErrs(err)
//	assert.Len(t, errs, 1, "errors length must be one")
//	assert.Contains(t, errs[0].Error(), "no(nil) document provided")
//
//	// unknown type
//	err = fv.Validate(nil, &mockModel{})
//	assert.Error(t, err)
//	errs = errors.GetErrs(err)
//	assert.Len(t, errs, 1, "errors length must be one")
//	assert.Contains(t, errs[0].Error(), "document is of invalid type")
//
//	// identity not created from the identity factory
//	idFactory := new(identity.MockFactory)
//	entity, _ := CreateEntityWithEmbedCD(t, testingconfig.CreateAccountContext(t, cfg), did, nil)
//	idFactory.On("IdentityExists", *entity.Data.Identity).Return(false, nil).Once()
//	fv = fieldValidator(idFactory)
//	err = fv.Validate(nil, entity)
//	assert.Error(t, err)
//	idFactory.AssertExpectations(t)
//
//	// identity created from identity factory
//	idFactory = new(identity.MockFactory)
//	idFactory.On("IdentityExists", *entity.Data.Identity).Return(true, nil).Once()
//	fv = fieldValidator(idFactory)
//	err = fv.Validate(nil, entity)
//	assert.NoError(t, err)
//	idFactory.AssertExpectations(t)
//}
