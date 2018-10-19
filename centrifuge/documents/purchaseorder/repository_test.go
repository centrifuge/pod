// +build unit

package purchaseorder

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	repo := GetLegacyRepository()

	// failed validation for create
	doc := &purchaseorderpb.PurchaseOrderDocument{
		CoreDocument: testingutils.GenerateCoreDocument(),
		Data: &purchaseorderpb.PurchaseOrderData{
			PoNumber:         "po1234",
			OrderName:        "Jack",
			OrderZipcode:     "921007",
			OrderCountry:     "Australia",
			RecipientName:    "John",
			RecipientZipcode: "12345",
			RecipientCountry: "Germany",
			Currency:         "EUR",
			OrderAmount:      800,
		},
	}

	docID := doc.CoreDocument.DocumentIdentifier
	err := repo.Create(docID, doc)
	assert.Error(t, err, "create must fail")

	// successful creation
	doc.Salts = &purchaseorderpb.PurchaseOrderDataSalts{}
	err = repo.Create(docID, doc)
	assert.Nil(t, err, "create must pass")

	// failed get
	getDoc := new(purchaseorderpb.PurchaseOrderDocument)
	err = repo.GetByID(doc.CoreDocument.NextVersion, getDoc)
	assert.Error(t, err, "get must fail")

	// successful get
	err = repo.GetByID(docID, getDoc)
	assert.Nil(t, err, "get must pass")
	assert.Equal(t, getDoc.CoreDocument.DocumentIdentifier, docID, "identifiers mismatch")

	// successful update
	doc.Data.OrderAmount = 200
	err = repo.Update(docID, doc)
	assert.Nil(t, err, "update must pass")
	assert.Nil(t, repo.GetByID(docID, getDoc), "get must pass")
	assert.Equal(t, getDoc.Data.OrderAmount, doc.Data.OrderAmount, "amount must match")
}

func TestRepository_getRepository(t *testing.T) {
	r := getRepository()
	assert.NotNil(t, r)
	assert.Equal(t, "purchaseorder", r.(*repository).KeyPrefix)
}
