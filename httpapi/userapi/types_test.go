// +build unit

package userapi

import (
	"context"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestTypes_toTransferDetailCreatePayload(t *testing.T) {
	req := CreateTransferDetailRequest{
		DocumentID: "test_id",
		Data: transferdetails.Data{
			TransferID: "test",
		},
	}
	// success
	payload, err := toTransferDetailCreatePayload(req)
	assert.NoError(t, err)
	assert.Equal(t, payload.DocumentID, "test_id")
	assert.NotNil(t, payload.Data)
	assert.Equal(t, payload.Data.TransferID, req.Data.TransferID)
}

func TestTypes_toTransferDetailUpdatePayload(t *testing.T) {
	req := UpdateTransferDetailRequest{
		DocumentID: "test_id",
		TransferID: "transfer_id",
		Data: transferdetails.Data{
			TransferID: "test",
		},
	}
	// success
	payload, err := toTransferDetailUpdatePayload(req)
	assert.NoError(t, err)
	assert.Equal(t, "test_id", payload.DocumentID)
	assert.Equal(t, "transfer_id", payload.TransferID)
	assert.NotNil(t, payload.Data)
	assert.Equal(t, payload.Data.TransferID, req.Data.TransferID)
}

func transferData() map[string]interface{} {
	return map[string]interface{}{
		"status":       "unpaid",
		"amount":       "300",
		"sender_id":    "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"recipient_id": "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_due":     "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"currency":     "EUR",
	}
}

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
	erSrv.On("GetEntityRelationships", ctx, eid).Return([]documents.Model{m}, nil).Once()
	_, err = getEntityRelationships(ctx, erSrv, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get access tokens")

	//success
	er := new(entityrelationship.EntityRelationship)
	er.CoreDocument = &documents.CoreDocument{
		Document: coredocumentpb.CoreDocument{},
	}
	er.TargetIdentity = &collab
	erSrv.On("GetEntityRelationships", ctx, eid).Return([]documents.Model{er}, nil).Once()
	rs, err = getEntityRelationships(ctx, erSrv, m)
	assert.NoError(t, err)
	assert.Len(t, rs, 1)
	assert.Equal(t, rs[0].Identity, collab)
	assert.False(t, rs[0].Active)
	m.AssertExpectations(t)
	erSrv.AssertExpectations(t)
}
