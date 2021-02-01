// +build unit

package userapi

import (
	"context"
	"encoding/json"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_getEntityRelationships(t *testing.T) {
	// missing did
	ctx := context.Background()
	_, err := getEntityRelationships(ctx, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get self ID")

	// failed to check collaborator
	m := new(testingdocuments.MockModel)
	collab := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, collab.String())
	m.On("IsDIDCollaborator", collab).Return(false, errors.New("failed to check DID")).Once()
	_, err = getEntityRelationships(ctx, nil, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check DID")

	// not a collaborator
	m.On("IsDIDCollaborator", collab).Return(false, nil).Once()
	rs, err := getEntityRelationships(ctx, nil, m)
	assert.NoError(t, err)
	assert.Nil(t, rs)

	// failed to get relationships
	eid := utils.RandomSlice(32)
	m.On("IsDIDCollaborator", collab).Return(true, nil)
	m.On("ID").Return(eid)
	erSrv := new(entity.MockEntityRelationService)
	erSrv.On("GetEntityRelationships", ctx, eid).Return(nil, errors.New("failed to get relationships")).Once()
	_, err = getEntityRelationships(ctx, erSrv, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get relationships")

	// failed to get access tokens
	m.On("GetAccessTokens").Return(nil, errors.New("failed to get access tokens")).Once()
	erSrv.On("GetEntityRelationships", ctx, eid).Return([]documents.Document{m}, nil).Once()
	_, err = getEntityRelationships(ctx, erSrv, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get access tokens")

	//success
	er := &entityrelationship.EntityRelationship{
		CoreDocument: &documents.CoreDocument{
			Document: coredocumentpb.CoreDocument{},
		},

		Data: entityrelationship.Data{
			TargetIdentity:   &collab,
			OwnerIdentity:    &collab,
			EntityIdentifier: eid,
		},
	}
	erSrv.On("GetEntityRelationships", ctx, eid).Return([]documents.Document{er}, nil).Once()
	rs, err = getEntityRelationships(ctx, erSrv, m)
	assert.NoError(t, err)
	assert.Len(t, rs, 1)
	assert.Equal(t, rs[0].TargetIdentity, collab)
	assert.False(t, rs[0].Active)
	m.AssertExpectations(t)
	erSrv.AssertExpectations(t)
}

func TestTypes_toEntityShareResponse(t *testing.T) {
	// failed to derive header
	model := new(testingdocuments.MockModel)
	model.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("error fetching collaborators")).Once()
	_, err := toEntityShareResponse(model, nil, jobs.NewJobID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error fetching collaborators")
	model.AssertExpectations(t)

	// success
	id := byteutils.HexBytes(utils.RandomSlice(32))
	did1 := testingidentity.GenerateRandomDID()
	did2 := testingidentity.GenerateRandomDID()
	er := &entityrelationship.EntityRelationship{
		CoreDocument: &documents.CoreDocument{
			Document: coredocumentpb.CoreDocument{},
		},

		Data: entityrelationship.Data{
			TargetIdentity:   &did1,
			OwnerIdentity:    &did2,
			EntityIdentifier: id,
		},
	}

	resp, err := toEntityShareResponse(er, nil, jobs.NewJobID())
	assert.NoError(t, err)
	assert.Equal(t, resp.Relationship.EntityIdentifier, id)
	assert.Equal(t, resp.Relationship.OwnerIdentity, did2)
	assert.Equal(t, resp.Relationship.TargetIdentity, did1)
	assert.True(t, resp.Relationship.Active)
}

func TestTypes_convertEntityShareRequest(t *testing.T) {
	// failed context
	ctx := context.Background()
	_, err := convertShareEntityRequest(ctx, nil, identity.DID{})
	assert.Error(t, err)

	// success
	did := testingidentity.GenerateRandomDID()
	did1 := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, did.String())
	docID := byteutils.HexBytes(utils.RandomSlice(32))
	req, err := convertShareEntityRequest(ctx, docID, did1)
	assert.NoError(t, err)
	assert.Equal(t, req.Scheme, entityrelationship.Scheme)
	var r entityrelationship.Data
	assert.NoError(t, json.Unmarshal(req.Data, &r))
	assert.Equal(t, r.EntityIdentifier, docID)
	assert.Equal(t, r.OwnerIdentity, &did)
	assert.Equal(t, r.TargetIdentity, &did1)
}
