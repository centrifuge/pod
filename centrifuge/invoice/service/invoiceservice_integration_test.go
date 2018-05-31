// +build ethereum

package invoiceservice

import (
	"testing"
	"github.com/spf13/viper"
	"os"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/stretchr/testify/assert"
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
)

var dbFileName = "/tmp/centrifuge_testing_inv_service.leveldb"

func TestMain(m *testing.M) {
	//for now set up the env vars manually in integration test
	//TODO move to generalized config once it is available
	viper.BindEnv("ethereum.gethSocket", "CENT_ETHEREUM_GETH_SOCKET")
	viper.BindEnv("ethereum.gasLimit", "CENT_ETHEREUM_GASLIMIT")
	viper.BindEnv("ethereum.gasPrice", "CENT_ETHEREUM_GASPRICE")
	viper.BindEnv("ethereum.contextWaitTimeout", "CENT_ETHEREUM_CONTEXTWAITTIMEOUT")
	viper.BindEnv("ethereum.accounts.main.password", "CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD")
	viper.BindEnv("ethereum.accounts.main.key", "CENT_ETHEREUM_ACCOUNTS_MAIN_KEY")
	viper.BindEnv("anchor.ethereum.anchorRegistryAddress", "CENT_ANCHOR_ETHEREUM_ANCHORREGISTRYADDRESS")

	viper.Set("storage.Path", dbFileName)
	defer Bootstrap().Close()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func Bootstrap() (*leveldb.DB) {
	levelDB := storage.NewLeveldbStorage(dbFileName)

	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})

	return levelDB
}

func generateEmptyInvoiceForProcessing() (doc *invoice.Invoice){
	identifier := testingutils.Rand32Bytes()
	doc = invoice.NewEmptyInvoice()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		DataMerkleRoot:     testingutils.Rand32Bytes(),
	}
	return
}

func TestInvoiceDocumentService_HandleAnchorInvoiceDocument_Integration(t *testing.T) {
	s := InvoiceDocumentService{}
	doc := generateEmptyInvoiceForProcessing()
	doc.Document.Data.Country = "DE"

	anchoredDoc, err := s.HandleAnchorInvoiceDocument(context.Background(), &invoicepb.AnchorInvoiceEnvelope{Document: doc.Document})

	//Call overall worked well and receive roughly sensical data back
	assert.Nil(t, err)
	assert.Equal(t, anchoredDoc.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument.DocumentIdentifier,
		"DocumentIdentifier doesn't match")

	//Invoice document got stored in the DB
	loadedInvoice, _ := invoicerepository.GetInvoiceRepository().FindById(doc.Document.CoreDocument.DocumentIdentifier)
	assert.Equal(t, "DE", loadedInvoice.Data.Country,
		"Didn't save the invoice data correctly")
}