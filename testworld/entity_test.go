// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	entitypb2 "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestHost_BasicEntity(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares a document with Bob and Charlie
	res := createDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusAccepted, defaultEntityPayload(alice.id.String(), []string{bob.id.String(), charlie.id.String()}))
	docIdentifier := getDocumentIdentifier(t, res)
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"legal_name":  "test company",
	}
	getEntityAndCheck(alice.httpExpect, alice.id.String(), typeEntity, params)
	getEntityAndCheck(bob.httpExpect, bob.id.String(), typeEntity, params)
	getEntityAndCheck(charlie.httpExpect, charlie.id.String(), typeEntity, params)
	fmt.Println("Host test success")
}

func TestHost_EntityShareGet(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice anchors Entity
	res := createDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusAccepted, defaultEntityPayload(alice.id.String(), []string{}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}
	entityIdentifier := getDocumentIdentifier(t, res)

	// Alice creates an EntityRelationship with Bob
	ctxAlice := testingconfig.CreateAccountContext(t, alice.host.config)
	relationshipData := &entitypb2.RelationshipData{
		EntityIdentifier: entityIdentifier,
		OwnerIdentity:    alice.id.String(),
		TargetIdentity:   bob.id.String(),
	}

	relationship := entityrelationship.EntityRelationship{}
	err := relationship.InitEntityRelationshipInput(ctxAlice, entityIdentifier, relationshipData)
	assert.NoError(t, err)

	relationshipModel, _, isDone, err := alice.host.entityService.Share(ctxAlice, &relationship)
	assert.NoError(t, err)
	done := <-isDone
	assert.True(t, done)
	cd, err := relationshipModel.PackCoreDocument()
	assert.NoError(t, err)

	relationshipIdentifier := cd.DocumentIdentifier
	params := map[string]interface{}{
		"r_identifier": hexutil.Encode(relationshipIdentifier),
	}
	response := getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, params)
	response.Path("$.data.entity.legal_name").String().Equal("test company")
}
