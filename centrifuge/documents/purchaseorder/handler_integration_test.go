// +build integration

package purchaseorder_test

import (
	"context"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	legacy "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/legacy/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	db := cc.GetLevelDBStorage()
	purchaseorder.InitLevelDBRepository(db)
	coredocumentrepository.InitLevelDBRepository(db)
	// TODO Once we move these tests to new model locations we can get rid of these configs
	config.Config.V.Set("keys.signing.publicKey", "../../../../example/resources/signature1.pub.pem")
	config.Config.V.Set("keys.signing.privateKey", "../../../../example/resources/signature1.key.pem")
	config.Config.V.Set("keys.ethauth.publicKey", "../../../../example/resources/ethauth.pub.pem")
	config.Config.V.Set("keys.ethauth.privateKey", "../../../../example/resources/ethauth.key.pem")
	testingutils.CreateIdentityWithKeys()
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func generateEmptyPurchaseOrderForProcessing() (doc *purchaseorder.PurchaseOrder) {
	identifier := testingutils.Rand32Bytes()
	doc = purchaseorder.Empty()
	orderAny := &any.Any{
		TypeUrl: documenttypes.PurchaseOrderDataTypeUrl,
		Value:   []byte{},
	}
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        testingutils.Rand32Bytes(),
		EmbeddedData:       orderAny,
	}
	salts := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(doc.Document.CoreDocument, salts)
	doc.Document.CoreDocument.CoredocumentSalts = salts
	return
}

func TestPurchaseOrderDocumentService_HandleAnchorPurchaseOrderDocument_Integration(t *testing.T) {
	p2pClient := testingcommons.NewMockP2PWrapperClient()
	s := purchaseorder.LegacyGRPCHandler(purchaseorder.GetRepository(), coredocumentprocessor.DefaultProcessor(identity.IDService, p2pClient))
	p2pClient.On("GetSignaturesForDocument", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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

	anchoredDoc, err := s.AnchorPurchaseOrderDocument(context.Background(), &legacy.AnchorPurchaseOrderEnvelope{Document: doc.Document})
	assertDocument(t, err, anchoredDoc, doc, s)
}

// TODO enable this after properly mocking p2p package eg: server.go
//func TestPurchaseOrderDocumentService_HandleSendPurchaseOrderDocument_Integration(t *testing.T) {
//	p2pClient := testingcommons.NewMockP2PWrapperClient()
//	s := purchaseorderservice.PurchaseOrderDocumentService{
//		Repository:            purchaseorderrepository.GetRepository(),
//		CoreDocumentProcessor: coredocumentprocessor.DefaultProcessor(identity.NewEthereumIdentityService(), p2pClient),
//	}
//	p2pClient.On("GetSignaturesForDocument", mock.Anything, mock.Anything, mock.Anything).Return(nil)
//	doc := generateEmptyPurchaseOrderForProcessing()
//	doc.Document.Data = &purchaseorderpb.PurchaseOrderData{
//		PoNumber:         "po1234",
//		OrderName:        "Jack",
//		OrderZipcode:     "921007",
//		OrderCountry:     "AUS",
//		RecipientName:    "John",
//		RecipientZipcode: "12345",
//		RecipientCountry: "DE",
//		Currency:         "EUR",
//		OrderAmount:      800,
//	}
//
//	anchoredDoc, err := s.HandleSendPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.SendPurchaseOrderEnvelope{
//		Document:   doc.Document,
//		Recipients: testingutils.GenerateP2PRecipientsOnEthereum(2),
//	})
//	assertDocument(t, err, anchoredDoc, doc, s)
//}

func assertDocument(t *testing.T, err error, anchoredDoc *purchaseorderpb.PurchaseOrderDocument, doc *purchaseorder.PurchaseOrder, s legacy.PurchaseOrderDocumentServiceServer) {
	//Call overall worked well and receive roughly sensical data back
	assert.Nil(t, err)
	assert.Equal(t, anchoredDoc.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument.DocumentIdentifier,
		"DocumentIdentifier doesn't match")
	//PurchaseOrder document got stored in the DB
	loadedDoc := new(purchaseorderpb.PurchaseOrderDocument)
	err = purchaseorder.GetRepository().GetByID(doc.Document.CoreDocument.DocumentIdentifier, loadedDoc)
	assert.Nil(t, err)
	assert.Equal(t, "AUS", loadedDoc.Data.OrderCountry,
		"Didn't save the purchaseorder data correctly")
	// Invoice stored after anchoring has Salts populated
	assert.NotNil(t, loadedDoc.Salts.OrderCountry)
	//PO Service should error out if trying to anchor the same document ID again
	doc.Document.Data.OrderCountry = "ES"
	anchoredDoc2, err := s.AnchorPurchaseOrderDocument(context.Background(), &legacy.AnchorPurchaseOrderEnvelope{Document: doc.Document})
	assert.Nil(t, anchoredDoc2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document already exists")
	loadedDoc2 := new(purchaseorderpb.PurchaseOrderDocument)
	err = purchaseorder.GetRepository().GetByID(doc.Document.CoreDocument.DocumentIdentifier, loadedDoc2)
	assert.Nil(t, err)
	assert.Equal(t, "AUS", loadedDoc2.Data.OrderCountry,
		"Document on DB should have not not gotten overwritten after rejected anchor call")
}
