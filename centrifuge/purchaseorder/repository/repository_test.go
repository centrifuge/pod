// +build integration

package purchaseorderrepository

import (
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/stretchr/testify/assert"
)

var dbFileName = "/tmp/centrifuge_testing_podoc.leveldb"

func TestMain(m *testing.M) {
	levelDB := storage.NewLevelDBStorage(dbFileName)
	coredocumentrepository.InitLevelDBRepository(levelDB)
	InitLevelDBRepository(levelDB)
	result := m.Run()
	levelDB.Close()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestRepository(t *testing.T) {
	repo := GetRepository()

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
	err = repo.GetByID(doc.CoreDocument.NextIdentifier, getDoc)
	assert.Error(t, err, "get must fail")

	// successful get
	err = repo.GetByID(docID, getDoc)
	assert.Nil(t, err, "get must pass")
	assert.Equal(t, getDoc.CoreDocument.DocumentIdentifier, docID, "identifiers mismatch")

	// failed update
	doc.Data.OrderAmount = 0
	err = repo.Update(docID, doc)
	assert.Error(t, err, "update must fail")

	// successful update
	doc.Data.OrderAmount = 200
	err = repo.Update(docID, doc)
	assert.Nil(t, err, "update must pass")
	assert.Nil(t, repo.GetByID(docID, getDoc), "get must pass")
	assert.Equal(t, getDoc.Data.OrderAmount, doc.Data.OrderAmount, "amount must match")
}
