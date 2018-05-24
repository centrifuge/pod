// +build unit

package invoiceservice

import (
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"testing"
)

var dbFileName = "/tmp/centrifuge_testing_inv_service.leveldb"

func TestMain(m *testing.M) {
	viper.Set("storage.Path", dbFileName)
	defer Bootstrap().Close()

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

func Bootstrap() *leveldb.DB {
	levelDB := storage.NewLeveldbStorage(dbFileName)

	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})

	return levelDB
}
