// +build ethereum

package purchaseorderservice_test

import (
	"context"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/processor"
	clientpurchaseorderpb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/service"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
)

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	db := cc.GetLevelDBStorage()
	purchaseorderrepository.InitLevelDBRepository(db)
	coredocumentrepository.InitLevelDBRepository(db)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func generateEmptyPurchaseOrderForProcessing() (doc *purchaseorder.PurchaseOrder) {
	identifier := testingutils.Rand32Bytes()
	doc = purchaseorder.Empty()
	salts := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(doc.Document.Data, salts)
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		CoredocumentSalts:  salts,
	}
	return
}

func TestPurchaseOrderDocumentService_HandleAnchorPurchaseOrderDocument_Integration(t *testing.T) {
	s := purchaseorderservice.PurchaseOrderDocumentService{
		Repository:            purchaseorderrepository.GetRepository(),
		CoreDocumentProcessor: coredocumentprocessor.DefaultProcessor(),
	}
	doc := generateEmptyPurchaseOrderForProcessing()
	doc.Document.Data = &purchaseorderpb.PurchaseOrderData{
		PoNumber:         "po1234",
		OrderName:        "Jack",
		OrderZipcode:     "921007",
		OrderCountry:     "AUS",
		RecipientName:    "John",
		RecipientZipcode: "12345",
		RecipientCountry: "DE",
		Currency:         "EUR",
		OrderAmount:      800,
	}

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	//Call overall worked well and receive roughly sensical data back
	assert.Nil(t, err)
	assert.Equal(t, anchoredDoc.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument.DocumentIdentifier,
		"DocumentIdentifier doesn't match")

	//PurchaseOrder document got stored in the DB
	loadedDoc := new(purchaseorderpb.PurchaseOrderDocument)
	err = purchaseorderrepository.GetRepository().GetByID(doc.Document.CoreDocument.DocumentIdentifier, loadedDoc)
	assert.Nil(t, err)
	assert.Equal(t, "AUS", loadedDoc.Data.OrderCountry,
		"Didn't save the purchaseorder data correctly")

	// Invoice stored after anchoring has Salts populated
	assert.NotNil(t, loadedDoc.Salts.OrderCountry)

	//PO Service should error out if trying to anchor the same document ID again
	doc.Document.Data.OrderCountry = "ES"
	anchoredDoc2, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})
	assert.Nil(t, anchoredDoc2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document already exists")

	loadedDoc2 := new(purchaseorderpb.PurchaseOrderDocument)
	err = purchaseorderrepository.GetRepository().GetByID(doc.Document.CoreDocument.DocumentIdentifier, loadedDoc2)
	assert.Nil(t, err)
	assert.Equal(t, "AUS", loadedDoc2.Data.OrderCountry,
		"Document on DB should have not not gotten overwritten after rejected anchor call")
}
