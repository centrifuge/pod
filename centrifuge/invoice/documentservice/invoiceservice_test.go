// +build unit

package documentservice

import (
	"testing"
	"github.com/spf13/viper"
	"os"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"github.com/stretchr/testify/assert"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"context"
)

var dbFileName = "/tmp/centrifuge_testing_inv_service.leveldb"

func TestMain(m *testing.M) {
	viper.Set("storage.Path", dbFileName)
	cc.Bootstrap()
	defer cc.LevelDB.Close()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestInvoiceService(t *testing.T) {
	// Set default key to use for signing
	viper.Set("keys.signing.publicKey", "../../resources/signingKey.pub")
	viper.Set("keys.signing.privateKey", "../../resources/signingKey.key")
	viper.Set("identityId", "1")

	identifier := testingutils.Rand32Bytes()
	identifierIncorrect := testingutils.Rand32Bytes()
	s := InvoiceDocumentService{}
	doc := invoice.NewEmptyInvoice()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier:identifier,
		CurrentIdentifier:identifier,
		NextIdentifier:testingutils.Rand32Bytes(),
		DataMerkleRoot: testingutils.Rand32Bytes(),
	}

	sentDoc, err := s.SendInvoiceDocument(context.Background(), &invoice.SendInvoiceEnvelope{[][]byte{}, doc.Document})
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
	assert.NotNil(t, err,
		"RPC call should have raised exception")

}
