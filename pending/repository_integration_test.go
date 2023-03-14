//go:build integration

package pending

import (
	"os"
	"testing"

	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/documents/generic"
	storage "github.com/centrifuge/pod/storage/leveldb"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
)

const (
	repoTestDirPattern = "pending-repo-integration-test*"
)

func Test_Integration_Repository_CRUD(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath(repoTestDirPattern)
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	storageRepo := storage.NewLevelDBRepository(db)

	repository := NewRepository(storageRepo)

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	res, err := repository.Get(accountID, documentID)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	doc, err := getTestDoc()
	assert.NoError(t, err)

	err = repository.Create(accountID, documentID, doc)
	assert.NoError(t, err)

	res, err = repository.Get(accountID, documentID)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	storageRepo.Register(doc)

	res, err = repository.Get(accountID, documentID)
	assert.NoError(t, err)
	assert.Equal(t, doc, res)

	newDoc, err := getTestDoc()
	assert.NoError(t, err)
	assert.NotEqual(t, newDoc, doc)

	err = repository.Update(accountID, documentID, newDoc)
	assert.NoError(t, err)

	res, err = repository.Get(accountID, documentID)
	assert.NoError(t, err)
	assert.Equal(t, newDoc, res)

	err = repository.Delete(accountID, documentID)
	assert.NoError(t, err)

	res, err = repository.Get(accountID, documentID)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func getTestDoc() (documents.Document, error) {
	cd, err := documents.NewCoreDocument(utils.RandomSlice(32), documents.CollaboratorsAccess{}, nil)

	if err != nil {
		return nil, err
	}

	return &generic.Generic{CoreDocument: cd}, nil
}
