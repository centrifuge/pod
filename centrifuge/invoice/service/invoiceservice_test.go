// +build unit

package invoiceservice

import (
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
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

func TestInvoiceService(t *testing.T) {
	// Set default key to use for signing
	config.Config.V.Set("keys.signing.publicKey", "../../example/resources/signingKey.pub")
	config.Config.V.Set("keys.signing.privateKey", "../../example/resources/signingKey.key")

	identifier := testingutils.Rand32Bytes()
	identifierIncorrect := testingutils.Rand32Bytes()
	s := InvoiceDocumentService{}
	doc := invoice.NewEmptyInvoice()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		DataMerkleRoot:     testingutils.Rand32Bytes(),
	}

	sentDoc, err := s.HandleSendInvoiceDocument(context.Background(), &invoicepb.SendInvoiceEnvelope{Recipients: [][]byte{}, Document: doc.Document})
	assert.Nil(t, err, "Error in RPC Call")

	assert.Equal(t, sentDoc.CoreDocument.DocumentIdentifier, identifier,
		"DocumentIdentifier doesn't match")

	receivedDoc, err := s.HandleGetInvoiceDocument(context.Background(),
		&invoicepb.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifier})
	assert.Nil(t, err, "Error in RPC Call")
	assert.Equal(t, receivedDoc.CoreDocument.DocumentIdentifier, identifier,
		"DocumentIdentifier doesn't match")

	_, err = s.HandleGetInvoiceDocument(context.Background(),
		&invoicepb.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifierIncorrect})
	assert.NotNil(t, err,
		"RPC call should have raised exception")

}
