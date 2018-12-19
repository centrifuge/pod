// +build unit

package documents_test

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var testRepoGlobal documents.Repository

var centIDBytes = utils.RandomSlice(identity.CentIDLength)

func getServiceWithMockedLayers() documents.Service {

	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	repo := testRepo()
	return documents.DefaultService(c, repo)
}

func TestService_GetCurrentVersion_successful(t *testing.T) {

	service := getServiceWithMockedLayers()
	documentIdentifier := utils.RandomSlice(32)
	const amountVersions = 10

	version := documentIdentifier
	var currentVersion []byte

	nonExistingVersion := utils.RandomSlice(32)

	for i := 0; i < amountVersions; i++ {

		var next []byte
		if i != amountVersions-1 {
			next = utils.RandomSlice(32)
		} else {
			next = nonExistingVersion
		}

		inv := &invoice.Invoice{
			GrossAmount: int64(i + 1),
			CoreDocument: &coredocumentpb.CoreDocument{
				DocumentIdentifier: documentIdentifier,
				CurrentVersion:     version,
				NextVersion:        next,
			},
		}

		err := testRepo().Create(centIDBytes, version, inv)
		currentVersion = version
		version = next
		assert.Nil(t, err)

	}

	model, err := service.GetCurrentVersion(documentIdentifier)
	assert.Nil(t, err)

	cd, err := model.PackCoreDocument()
	assert.Nil(t, err)

	assert.Equal(t, cd.CurrentVersion, currentVersion, "should return latest version")
	assert.Equal(t, cd.NextVersion, nonExistingVersion, "latest version should have a non existing id as nextVersion ")

}

func TestService_GetVersion_successful(t *testing.T) {

	service := getServiceWithMockedLayers()
	documentIdentifier := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	inv := &invoice.Invoice{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     currentVersion,
		},
	}

	err := testRepo().Create(centIDBytes, currentVersion, inv)
	assert.Nil(t, err)

	mod, err := service.GetVersion(documentIdentifier, currentVersion)
	assert.Nil(t, err)

	cd, err := mod.PackCoreDocument()
	assert.Nil(t, err)

	assert.Equal(t, documentIdentifier, cd.DocumentIdentifier, "should be same document Identifier")
	assert.Equal(t, currentVersion, cd.CurrentVersion, "should be same version")
}

func TestService_GetCurrentVersion_error(t *testing.T) {
	service := getServiceWithMockedLayers()
	documentIdentifier := utils.RandomSlice(32)

	//document is not existing
	_, err := service.GetCurrentVersion(documentIdentifier)
	assert.True(t, errors.IsOfType(documents.ErrDocumentVersionNotFound, err))

	inv := &invoice.Invoice{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     documentIdentifier,
		},
	}

	err = testRepo().Create(centIDBytes, documentIdentifier, inv)
	assert.Nil(t, err)

	_, err = service.GetCurrentVersion(documentIdentifier)
	assert.Nil(t, err)

}

func TestService_GetVersion_error(t *testing.T) {

	service := getServiceWithMockedLayers()

	documentIdentifier := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)

	//document is not existing
	_, err := service.GetVersion(documentIdentifier, currentVersion)
	assert.True(t, errors.IsOfType(documents.ErrDocumentVersionNotFound, err))

	inv := &invoice.Invoice{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     currentVersion,
		},
	}
	err = testRepo().Create(centIDBytes, currentVersion, inv)
	assert.Nil(t, err)

	//random version
	_, err = service.GetVersion(documentIdentifier, utils.RandomSlice(32))
	assert.True(t, errors.IsOfType(documents.ErrDocumentVersionNotFound, err))

	//random document id
	_, err = service.GetVersion(utils.RandomSlice(32), documentIdentifier)
	assert.True(t, errors.IsOfType(documents.ErrDocumentVersionNotFound, err))
}

func testRepo() documents.Repository {
	if testRepoGlobal == nil {
		ldb, err := storage.NewLevelDBStorage(storage.GetRandomTestStoragePath())
		if err != nil {
			panic(err)
		}
		testRepoGlobal = documents.NewDBRepository(storage.NewLevelDBRepository(ldb))
		testRepoGlobal.Register(&invoice.Invoice{})
	}
	return testRepoGlobal
}
