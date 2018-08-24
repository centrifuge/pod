// +build ethereum

package purchaseorderservice_test

import (
	"context"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	clientpurchaseorderpb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/service"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	purchaseorderrepository.InitLevelDBRepository(cc.GetLevelDBStorage())

	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func generateEmptyPurchaseOrderForProcessing() (doc *purchaseorder.PurchaseOrder) {
	identifier := testingutils.Rand32Bytes()
	doc = purchaseorder.NewEmptyPurchaseOrder()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
	}
	return
}

func TestPurchaseOrderDocumentService_HandleAnchorPurchaseOrderDocument_Integration(t *testing.T) {
	s := purchaseorderservice.PurchaseOrderDocumentService{
		Repository:            purchaseorderrepository.GetRepository(),
		CoreDocumentProcessor: coredocument.GetDefaultCoreDocumentProcessor(),
	}
	doc := generateEmptyPurchaseOrderForProcessing()
	doc.Document.Data.OrderCountry = "DE"

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	//Call overall worked well and receive roughly sensical data back
	assert.Nil(t, err)
	assert.Equal(t, anchoredDoc.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument.DocumentIdentifier,
		"DocumentIdentifier doesn't match")

	//PurchaseOrder document got stored in the DB
	loadedDoc := new(purchaseorderpb.PurchaseOrderDocument)
	err = purchaseorderrepository.GetRepository().GetByID(doc.Document.CoreDocument.DocumentIdentifier, loadedDoc)
	assert.Nil(t, err)
	assert.Equal(t, "DE", loadedDoc.Data.OrderCountry,
		"Didn't save the purchaseorder data correctly")

	//PO Service should error out if trying to anchor the same document ID again
	doc.Document.Data.OrderCountry = "ES"
	anchoredDoc2, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})
	assert.Nil(t, anchoredDoc2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document already exists")

	loadedDoc2 := new(purchaseorderpb.PurchaseOrderDocument)
	err = purchaseorderrepository.GetRepository().GetByID(doc.Document.CoreDocument.DocumentIdentifier, loadedDoc2)
	assert.Nil(t, err)
	assert.Equal(t, "DE", loadedDoc2.Data.OrderCountry,
		"Document on DB should have not not gotten overwritten after rejected anchor call")
}
