// +build unit

package purchaseorderservice

import (
	"context"
	"crypto/sha256"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cc.TestUnitBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

func TestPurchaseOrderDocumentService_SendReceive(t *testing.T) {
	s := PurchaseOrderDocumentService{}
	identifier := testingutils.Rand32Bytes()
	identifierIncorrect := testingutils.Rand32Bytes()
	doc := purchaseorder.NewEmptyPurchaseOrder()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		DataMerkleRoot:     testingutils.Rand32Bytes(),
	}

	sentDoc, err := s.HandleSendPurchaseOrderDocument(context.Background(), &purchaseorderpb.SendPurchaseOrderEnvelope{Recipients: [][]byte{}, Document: doc.Document})
	assert.Nil(t, err, "Error in RPC Call")

	assert.Equal(t, sentDoc.CoreDocument.DocumentIdentifier, identifier,
		"DocumentIdentifier doesn't match")

	receivedDoc, err := s.HandleGetPurchaseOrderDocument(context.Background(),
		&purchaseorderpb.GetPurchaseOrderDocumentEnvelope{DocumentIdentifier: identifier})
	assert.Nil(t, err, "Error in RPC Call")
	assert.Equal(t, receivedDoc.CoreDocument.DocumentIdentifier, identifier,
		"DocumentIdentifier doesn't match")

	_, err = s.HandleGetPurchaseOrderDocument(context.Background(),
		&purchaseorderpb.GetPurchaseOrderDocumentEnvelope{DocumentIdentifier: identifierIncorrect})
	assert.NotNil(t, err,
		"RPC call should have raised exception")

}

func TestPurchaseOrderDocumentService_HandleCreatePurchaseOrderProof(t *testing.T) {
	s := PurchaseOrderDocumentService{}

	identifier := testingutils.Rand32Bytes()
	inv := purchaseorder.NewEmptyPurchaseOrder()
	inv.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		// TODO: below should be actual merkle root
		DataMerkleRoot: testingutils.Rand32Bytes(),
	}
	inv.CalculateMerkleRoot()
	err := purchaseorderrepository.GetPurchaseOrderRepository().Store(inv.Document)
	assert.Nil(t, err)

	proofRequest := &purchaseorderpb.CreatePurchaseOrderProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "country", "amount"},
	}

	purchaseorderProof, err := s.HandleCreatePurchaseOrderProof(context.Background(), proofRequest)
	assert.Nil(t, err)
	assert.Equal(t, identifier, purchaseorderProof.DocumentIdentifier)
	assert.Equal(t, len(proofRequest.Fields), len(purchaseorderProof.FieldProofs))
	assert.Equal(t, proofRequest.Fields[0], purchaseorderProof.FieldProofs[0].Property)
	sha256Hash := sha256.New()
	valid, err := proofs.ValidateProof(purchaseorderProof.FieldProofs[0], inv.Document.CoreDocument.DocumentRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)
}
