// +build integration

package invoicerepository_test

import (
	"testing"
	"os"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{cc.GetLevelDBStorage()})

	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func TestStorageService(t *testing.T) {
	identifier := []byte("1")
	invalidIdentifier := []byte("2")

	invoice := invoicepb.InvoiceDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: identifier}}
	repo := invoicerepository.GetInvoiceRepository()
	err := repo.Store(&invoice)
	assert.Nil(t, err, "Store should not return error")

	inv, err := repo.FindById(identifier)
	assert.Nil(t, err, "FindById should not return error")
	assert.Equal(t, invoice.CoreDocument.DocumentIdentifier, inv.CoreDocument.DocumentIdentifier, "Invoice DocumentIdentifier should be equal")

	inv, err = repo.FindById(invalidIdentifier)
	assert.NotNil(t, err, "FindById should return error")
	assert.Nil(t, inv, "Invoice should be NIL")
}

func TestStoreOnce(t *testing.T) {
	identifier := []byte("1234")

	repo := invoicerepository.GetInvoiceRepository()

	invoice := invoicepb.InvoiceDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: identifier, CurrentIdentifier: []byte("333")}}
	err := repo.StoreOnce(&invoice)
	assert.Nil(t, err)

	loadedInvoice, err := repo.FindById(identifier)
	assert.Nil(t, err)
	assert.Equal(t, identifier, loadedInvoice.CoreDocument.DocumentIdentifier)
	assert.Equal(t, []byte("333"), loadedInvoice.CoreDocument.CurrentIdentifier)

	invoice2 := invoicepb.InvoiceDocument{CoreDocument: &coredocumentpb.CoreDocument{DocumentIdentifier: identifier, CurrentIdentifier: []byte("666")}}
	err = repo.StoreOnce(&invoice2)
	assert.Error(t, err)
	assert.Equal(t, "Document already exists. StoreOnce will not overwrite.", err.Error())

	loadedInvoice, err = repo.FindById(identifier)
	assert.Nil(t, err)
	assert.Equal(t, identifier, loadedInvoice.CoreDocument.DocumentIdentifier)
	assert.Equal(t, []byte("333"), loadedInvoice.CoreDocument.CurrentIdentifier, "Loaded invoice should still have the old values as the overwrite should have failed")
	//TODO make into a generic error from the errors package after Miguel's merge

}

func TestLevelDBInvoiceRepository_StoreNilDocument(t *testing.T) {
	repo := invoicerepository.GetInvoiceRepository()
	err := repo.Store(nil)

	assert.Error(t, err, "should have thrown an error")
}

func TestLevelDBInvoiceRepository_StoreNilCoreDocument(t *testing.T) {
	repo := invoicerepository.GetInvoiceRepository()
	err := repo.Store(&invoicepb.InvoiceDocument{})

	assert.Error(t, err, "should have thrown an error")
}