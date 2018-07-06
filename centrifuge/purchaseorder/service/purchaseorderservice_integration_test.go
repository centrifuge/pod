// +build ethereum

package purchaseorderservice_test

import (
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/service"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cc.TestBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

func generateEmptyPurchaseOrderForProcessing() (doc *purchaseorder.PurchaseOrder) {
	identifier := testingutils.Rand32Bytes()
	doc = purchaseorder.NewEmptyPurchaseOrder()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		DataRoot:           testingutils.Rand32Bytes(),
	}
	return
}

func TestPurchaseOrderDocumentService_HandleAnchorPurchaseOrderDocument_Integration(t *testing.T) {
	s := purchaseorderservice.PurchaseOrderDocumentService{
		PurchaseOrderRepository: purchaseorderrepository.GetPurchaseOrderRepository(),
		CoreDocumentProcessor:   coredocument.GetDefaultCoreDocumentProcessor(),
	}
	doc := generateEmptyPurchaseOrderForProcessing()
	doc.Document.Data.OrderCountry = "DE"

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &purchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	//Call overall worked well and receive roughly sensical data back
	assert.Nil(t, err)
	assert.Equal(t, anchoredDoc.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument.DocumentIdentifier,
		"DocumentIdentifier doesn't match")

	//PurchaseOrder document got stored in the DB
	loadedPurchaseOrder, _ := purchaseorderrepository.GetPurchaseOrderRepository().FindById(doc.Document.CoreDocument.DocumentIdentifier)
	assert.Equal(t, "DE", loadedPurchaseOrder.Data.OrderCountry,
		"Didn't save the purchaseorder data correctly")
}
