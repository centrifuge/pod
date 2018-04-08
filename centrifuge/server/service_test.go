// +build unit

package server

import (
	"testing"
	"context"
	"bytes"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/documentservice"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	invoicepb "github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"os"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var dbFileName = "/tmp/centrifuge_testing_server_service.leveldb"

func TestMain(m *testing.M) {
	viper.Set("storage.Path", dbFileName)
	mockBootstrap()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestInvoiceService(t *testing.T) {
	// Set default key to use for signing
	viper.Set("keys.signing.publicKey", "../../resources/signingKey.pub")
	viper.Set("keys.signing.privateKey", "../../resources/signingKey.key")
	viper.Set("identityId", "1")
	cc.Node.GetSigningService().LoadIdentityKeyFromConfig()

	identifier := testingutils.Rand32Bytes()
	identifierIncorrect := testingutils.Rand32Bytes()
	s := documentservice.InvoiceDocumentService{}
	doc := invoicepb.InvoiceDocument{
		CoreDocument: &coredocument.CoreDocument{
			DocumentIdentifier:identifier,
			CurrentIdentifier:identifier,
			NextIdentifier:testingutils.Rand32Bytes(),
			DataMerkleRoot: testingutils.Rand32Bytes(),
			},
		Data: &invoicepb.InvoiceData{},
	}

	sentDoc, err := s.SendInvoiceDocument(context.Background(), &invoice.SendInvoiceEnvelope{[][]byte{}, &doc})
	assert.Nil(t, err, "Error in RPC Call")

	assert.Equal(t, sentDoc.CoreDocument.DocumentIdentifier, identifier,
		"DocumentIdentifier doesn't match")

	receivedDoc, err := s.GetInvoiceDocument(context.Background(),
		&invoice.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifier})
	assert.Nil(t, err, "Error in RPC Call")
	assert.Equal(t, receivedDoc.CoreDocument.DocumentIdentifier, identifier,
		"DocumentIdentifier doesn't match")

	_, err = s.GetInvoiceDocument(context.Background(),
		&invoice.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifierIncorrect})
	assert.Nil(t, err,
		"RPC call should have raised exception")

}

func mockBootstrap() {
	(&cc.MockCentNode{}).BootstrapDependencies()
}