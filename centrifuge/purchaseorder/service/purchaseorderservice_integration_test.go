// +build ethereum

package purchaseorderservice

import (
	"testing"
	"os"
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/stretchr/testify/assert"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
)


func TestMain(m *testing.M) {
	cc.TestBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

func generateEmptyPurchaseOrderForProcessing() (doc *purchaseorder.PurchaseOrder){
	identifier := testingutils.Rand32Bytes()
	doc = purchaseorder.NewEmptyPurchaseOrder()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		DataMerkleRoot:     testingutils.Rand32Bytes(),
	}
	return
}

func TestPurchaseOrderDocumentService_HandleAnchorPurchaseOrderDocument_Integration(t *testing.T) {
	s := PurchaseOrderDocumentService{}
	doc := generateEmptyPurchaseOrderForProcessing()
	doc.Document.Data.Country = "DE"

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &purchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	//Call overall worked well and receive roughly sensical data back
	assert.Nil(t, err)
	assert.Equal(t, anchoredDoc.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument.DocumentIdentifier,
		"DocumentIdentifier doesn't match")

	//PurchaseOrder document got stored in the DB
	loadedPurchaseOrder, _ := purchaseorderrepository.GetPurchaseOrderRepository().FindById(doc.Document.CoreDocument.DocumentIdentifier)
	assert.Equal(t, "DE", loadedPurchaseOrder.Data.Country,
		"Didn't save the purchaseorder data correctly")
}