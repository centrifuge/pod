// +build integration

package invoice

import (
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/stretchr/testify/assert"
)

var dbFileName = "/tmp/centrifuge_testing_invoicedoc.leveldb"

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	db := storage.NewLevelDBStorage(dbFileName)
	coredocumentrepository.InitLevelDBRepository(db)
	InitLevelDBRepository(db)
	result := m.Run()
	cc.TestIntegrationTearDown()
	db.Close()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestRepository(t *testing.T) {
	repo := GetRepository()

	// failed validation for create
	doc := &invoicepb.InvoiceDocument{
		CoreDocument: testingutils.GenerateCoreDocument(),
		Data: &invoicepb.InvoiceData{
			InvoiceNumber:    "inv1234",
			SenderName:       "Jack",
			SenderZipcode:    "921007",
			SenderCountry:    "AUS",
			RecipientName:    "John",
			RecipientZipcode: "12345",
			RecipientCountry: "Germany",
			Currency:         "EUR",
			GrossAmount:      800,
		},
	}

	docID := doc.CoreDocument.DocumentIdentifier
	err := repo.Create(docID, doc)
	assert.Error(t, err, "create must fail")

	// successful creation
	doc.Salts = &invoicepb.InvoiceDataSalts{}
	err = repo.Create(docID, doc)
	assert.Nil(t, err, "create must pass")

	// failed get
	getDoc := new(invoicepb.InvoiceDocument)
	err = repo.GetByID(doc.CoreDocument.NextIdentifier, getDoc)
	assert.Error(t, err, "get must fail")

	// successful get
	err = repo.GetByID(docID, getDoc)
	assert.Nil(t, err, "get must pass")
	assert.Equal(t, getDoc.CoreDocument.DocumentIdentifier, docID, "identifiers mismatch")

	// successful update
	doc.Data.GrossAmount = 200
	err = repo.Update(docID, doc)
	assert.Nil(t, err, "update must pass")
	assert.Nil(t, repo.GetByID(docID, getDoc), "get must pass")
	assert.Equal(t, getDoc.Data.GrossAmount, doc.Data.GrossAmount, "amount must match")
}
