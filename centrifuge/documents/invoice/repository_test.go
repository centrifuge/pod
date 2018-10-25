// +build integration

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	repo := getRepository()
	invRepo := repo.(*repository)
	assert.Equal(t, invRepo.KeyPrefix, "invoice")
	assert.NotNil(t, invRepo.LevelDB, "missing leveldb instance")

	id := utils.RandomSlice(32)
	doc := &Invoice{
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
	getDoc := new(Invoice)
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

func TestRepository_getRepository(t *testing.T) {
	r := getRepository()
	assert.NotNil(t, r)
	assert.Equal(t, "invoice", r.(*repository).KeyPrefix)
}
