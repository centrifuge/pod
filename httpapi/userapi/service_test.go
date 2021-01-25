// +build unit

package userapi

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newCoreAPIService(docSrv documents.Service) coreapi.Service {
	return coreapi.NewService(docSrv, nil, nil, nil)
}

func TestService_CreateEntity(t *testing.T) {
	ctx := context.Background()
	did := testingidentity.GenerateRandomDID()
	req := CreateEntityRequest{
		WriteAccess: []identity.DID{did},
		Data: entity.Data{
			Identity:  &did,
			LegalName: "John Doe",
			Addresses: []entity.Address{
				{
					IsMain:  true,
					Country: "Germany",
					Label:   "home",
				},
			},
		},
		Attributes: coreapi.AttributeMapRequest{
			"string_test": {
				Type:  "invalid",
				Value: "hello, world!",
			},

			"decimal_test": {
				Type:  "decimal",
				Value: "100001.001",
			},
		},
	}
	s := Service{}

	// invalid attribute map
	_, _, err := s.CreateEntity(ctx, req)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrNotValidAttrType, err))

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	docSrv.On("CreateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil)
	s.coreAPISrv = newCoreAPIService(docSrv)
	strAttr := req.Attributes["string_test"]
	strAttr.Type = "string"
	req.Attributes["string_test"] = strAttr
	_, _, err = s.CreateEntity(ctx, req)
	assert.NoError(t, err)
	docSrv.AssertExpectations(t)
}

func TestService_UpdateEntity(t *testing.T) {
	ctx := context.Background()
	did := testingidentity.GenerateRandomDID()
	req := CreateEntityRequest{
		WriteAccess: []identity.DID{did},
		Data: entity.Data{
			Identity:  &did,
			LegalName: "John Doe",
			Addresses: []entity.Address{
				{
					IsMain:  true,
					Country: "Germany",
					Label:   "home",
				},
			},
		},
		Attributes: coreapi.AttributeMapRequest{
			"string_test": {
				Type:  "invalid",
				Value: "hello, world!",
			},

			"decimal_test": {
				Type:  "decimal",
				Value: "100001.001",
			},
		},
	}
	s := Service{}

	docID := utils.RandomSlice(32)

	// invalid attribute map
	_, _, err := s.UpdateEntity(ctx, docID, req)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrNotValidAttrType, err))

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	docSrv.On("UpdateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil)
	s.coreAPISrv = newCoreAPIService(docSrv)
	strAttr := req.Attributes["string_test"]
	strAttr.Type = "string"
	req.Attributes["string_test"] = strAttr
	_, _, err = s.UpdateEntity(ctx, docID, req)
	assert.NoError(t, err)
	docSrv.AssertExpectations(t)
}

func TestService_ShareEntity(t *testing.T) {
	// failed to convert
	ctx := context.Background()
	s := Service{}
	_, _, err := s.ShareEntity(ctx, nil, ShareEntityRequest{})
	assert.Error(t, err)

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	s.coreAPISrv = newCoreAPIService(docSrv)
	did := testingidentity.GenerateRandomDID()
	did1 := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, did.String())
	docSrv.On("CreateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil).Once()
	docID := byteutils.HexBytes(utils.RandomSlice(32))
	m1, _, err := s.ShareEntity(ctx, docID, ShareEntityRequest{TargetIdentity: did1})
	assert.NoError(t, err)
	assert.Equal(t, m, m1)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestService_RevokeRelationship(t *testing.T) {
	// failed to convert
	ctx := context.Background()
	s := Service{}
	_, _, err := s.RevokeRelationship(ctx, nil, ShareEntityRequest{})
	assert.Error(t, err)

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	s.coreAPISrv = newCoreAPIService(docSrv)
	did := testingidentity.GenerateRandomDID()
	did1 := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, did.String())
	docSrv.On("UpdateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil).Once()
	docID := byteutils.HexBytes(utils.RandomSlice(32))
	m1, _, err := s.RevokeRelationship(ctx, docID, ShareEntityRequest{TargetIdentity: did1})
	assert.NoError(t, err)
	assert.Equal(t, m, m1)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}
