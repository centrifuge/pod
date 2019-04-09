// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	entitypb2 "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestHost_P2PGetDocumentWithToken(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// alice anchors entity
	res := createDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusOK, defaultEntityPayload(alice.id.String(), []string{}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}
	entityIdentifier := getDocumentIdentifier(t, res)

	// alice creates a entity-relationship with Bob
	// by directly using alice service
	ctxAlice := testingconfig.CreateAccountContext(t, alice.host.config)
	erData := &entitypb2.RelationshipData{
		OwnerIdentity:  alice.id.String(),
		TargetIdentity: bob.id.String(),
	}

	er := entityrelationship.EntityRelationship{}
	er.InitEntityRelationshipInput(ctxAlice, entityIdentifier, erData, alice.id)

	erModel, _, isDone, err := alice.host.erService.Create(ctxAlice, &er)
	assert.NoError(t, err)
	done := <-isDone
	assert.True(t, done)
	cd, err := erModel.PackCoreDocument()
	assert.Nil(t, err)

	erIdentifier := cd.DocumentIdentifier

	// Bob should have the entityRelationship
	ctxBob := testingconfig.CreateAccountContext(t, bob.host.config)
	bobModel, err := bob.host.erService.GetCurrentVersion(ctxBob, erIdentifier)
	assert.NoError(t, err)

	assert.Equal(t, erModel.CurrentVersion(), bobModel.CurrentVersion())

	// Bob access Entity directly on p2p
	accessTokenRequest := &p2ppb.AccessTokenRequest{DelegatingDocumentIdentifier: erIdentifier, AccessTokenId: cd.AccessTokens[0].Identifier}
	entityIdentifierByte, err := hexutil.Decode(entityIdentifier) // remove 0x
	assert.NoError(t, err)
	request := &p2ppb.GetDocumentRequest{DocumentIdentifier: entityIdentifierByte,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: accessTokenRequest,
	}

	response, err := bob.host.p2pClient.GetDocumentRequest(ctxBob, alice.id, request)
	assert.NoError(t, err)
	assert.Equal(t, response.Document.DocumentIdentifier, entityIdentifierByte)

}
