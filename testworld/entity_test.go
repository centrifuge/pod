// +build testworld

package testworld

import (
	"fmt"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	entitypb2 "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestHost_BasicEntity(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// alice shares a document with bob and charlie
	res := createDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusOK, defaultEntityPayload(alice.id.String(), []string{bob.id.String(), charlie.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)

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

	time.Sleep(3000)
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

	// todo replace with Share endpoint
	ctxAlice := testingconfig.CreateAccountContext(t, alice.host.config)
	erData := &entitypb2.RelationshipData{
		EntityIdentifier: entityIdentifier,
		OwnerIdentity:    alice.id.String(),
		TargetIdentity:   bob.id.String(),
	}

	er := entityrelationship.EntityRelationship{}
	err := er.InitEntityRelationshipInput(ctxAlice, entityIdentifier, erData)
	assert.NoError(t, err)

	erModel, _, isDone, err := alice.host.erService.Create(ctxAlice, &er)
	assert.NoError(t, err)
	done := <-isDone
	assert.True(t, done)
	cd, err := erModel.PackCoreDocument()
	assert.Nil(t, err)
	// todo end

	erIdentifier := cd.DocumentIdentifier

	params := map[string]interface{}{
		"entity_identifier": entityIdentifier,
		"er_identifier":  	hexutil.Encode(erIdentifier),
	}

	response := getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, params)
	fmt.Println(response)


}
