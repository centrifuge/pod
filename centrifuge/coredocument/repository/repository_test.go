// +build unit

package coredocumentrepository_test

import (
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	coredocumentrepository.InitLevelDBRepository(cc.GetLevelDBStorage())
	result := m.Run()
	os.Exit(result)
}

var (
	id1 = tools.RandomSlice(32)
	id2 = tools.RandomSlice(32)
	id3 = tools.RandomSlice(32)
	id4 = tools.RandomSlice(32)
	id5 = tools.RandomSlice(32)
)

func TestRepository(t *testing.T) {
	repo := coredocumentrepository.GetRepository()

	// failed validation for create
	doc := &coredocumentpb.CoreDocument{
		DocumentRoot:       id1,
		DocumentIdentifier: id2,
		CurrentVersion:     id3,
		DataRoot:           id5,
	}

	err := repo.Create(doc.DocumentIdentifier, doc)
	assert.Error(t, err, "create must fail")

	// successful creation
	doc.NextVersion = id4
	doc.CoredocumentSalts = &coredocumentpb.CoreDocumentSalts{
		DocumentIdentifier: id1,
		CurrentVersion:     id2,
		NextVersion:        id3,
		DataRoot:           id4,
		PreviousRoot:       id5,
	}
	err = repo.Create(doc.DocumentIdentifier, doc)
	assert.Nil(t, err, "create must pass")

	// failed get
	getDoc := new(coredocumentpb.CoreDocument)
	err = repo.GetByID(doc.NextVersion, getDoc)
	assert.Error(t, err, "get must fail")

	// successful get
	err = repo.GetByID(doc.DocumentIdentifier, getDoc)
	assert.Nil(t, err, "get must pass")
	assert.Equal(t, doc.DocumentIdentifier, getDoc.DocumentIdentifier, "identifiers mismatch")

	// failed update
	doc.NextVersion = doc.CurrentVersion
	err = repo.Update(doc.DocumentIdentifier, doc)
	assert.Error(t, err, "update must fail")

	// successful update
	id6 := tools.RandomSlice(32)
	doc.NextVersion = id6
	err = repo.Update(doc.DocumentIdentifier, doc)
	assert.Nil(t, err, "update must pass")
	err = repo.GetByID(doc.DocumentIdentifier, getDoc)
	assert.Nil(t, err, "get  must pass")
	assert.Equal(t, doc.DocumentIdentifier, getDoc.DocumentIdentifier, "identifier mismatch")
	assert.Equal(t, id6, getDoc.NextVersion, "identifier mismatch")
}
