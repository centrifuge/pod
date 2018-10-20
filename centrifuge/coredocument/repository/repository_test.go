// +build unit

package coredocumentrepository

import (
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, nil)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

var (
	id1 = utils.RandomSlice(32)
	id2 = utils.RandomSlice(32)
	id3 = utils.RandomSlice(32)
	id4 = utils.RandomSlice(32)
	id5 = utils.RandomSlice(32)
)

func TestRepository(t *testing.T) {
	repo := GetRepository()

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
	id6 := utils.RandomSlice(32)
	doc.NextVersion = id6
	err = repo.Update(doc.DocumentIdentifier, doc)
	assert.Nil(t, err, "update must pass")
	err = repo.GetByID(doc.DocumentIdentifier, getDoc)
	assert.Nil(t, err, "get  must pass")
	assert.Equal(t, doc.DocumentIdentifier, getDoc.DocumentIdentifier, "identifier mismatch")
	assert.Equal(t, id6, getDoc.NextVersion, "identifier mismatch")
}
