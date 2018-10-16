// +build integration

package invoice

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestLegacyRepository(t *testing.T) {
	repo := GetLegacyRepository()

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
	err = repo.GetByID(doc.CoreDocument.NextVersion, getDoc)
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

func TestRepository(t *testing.T) {
	repo := getRepository()
	invRepo := repo.(*repository)
	assert.Equal(t, invRepo.KeyPrefix, "invoice")
	assert.NotNil(t, invRepo.LevelDB, "missing leveldb instance")

	id := tools.RandomSlice(32)
	doc := &InvoiceModel{
		InvoiceNumber:    "inv1234",
		SenderName:       "Jack",
		SenderZipcode:    "921007",
		SenderCountry:    "AUS",
		RecipientName:    "John",
		RecipientZipcode: "12345",
		RecipientCountry: "Germany",
		Currency:         "EUR",
		GrossAmount:      800,
	}

	err := repo.Create(id, doc)
	assert.Nil(t, err, "create must pass")

	// successful get
	getDoc := new(InvoiceModel)
	err = repo.LoadByID(id, getDoc)
	assert.Nil(t, err, "get must pass")
	assert.Equal(t, getDoc, doc, "documents mismatch")

	// successful update
	doc.GrossAmount = 200
	err = repo.Update(id, doc)
	assert.Nil(t, err, "update must pass")
	assert.Nil(t, repo.LoadByID(id, getDoc), "get must pass")
	assert.Equal(t, getDoc.GrossAmount, doc.GrossAmount, "amount must match")
}
