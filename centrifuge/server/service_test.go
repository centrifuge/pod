// +build unit

package server

import (
	"testing"
	"context"
	"bytes"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/documentservice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"os"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/spf13/viper"
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
	doc := invoice.InvoiceDocument{
		CoreDocument: &coredocument.CoreDocument{
			DocumentIdentifier:identifier,
			CurrentIdentifier:identifier,
			NextIdentifier:testingutils.Rand32Bytes(),
			DataMerkleRoot: testingutils.Rand32Bytes(),
			},
		Data: &invoice.InvoiceData{},
	}

	sentDoc, err := s.SendInvoiceDocument(context.Background(), &invoice.SendInvoiceEnvelope{[][]byte{}, &doc})
	if err != nil {
		t.Fatal("Error in RPC Call", err)
	}
	if !bytes.Equal(sentDoc.CoreDocument.DocumentIdentifier, identifier) {
		t.Fatal("DocumentIdentifier doesn't match")
	}
	receivedDoc, err := s.GetInvoiceDocument(context.Background(),
		&invoice.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifier})
	if err != nil {
		t.Fatal("Error in RPC Call", err)
	}
	if !bytes.Equal(receivedDoc.CoreDocument.DocumentIdentifier, identifier) {
		t.Fatal("DocumentIdentifier doesn't match")
	}
	_, err = s.GetInvoiceDocument(context.Background(),
		&invoice.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifierIncorrect})
	if err == nil {
		t.Fatal("RPC call should have raised exception")
	}
}

func mockBootstrap() {
	(&cc.MockCentNode{}).BootstrapDependencies()
}