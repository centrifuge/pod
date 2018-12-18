// +build unit

package documents_test

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testRepoGlobal documents.Repository

var centIDBytes = utils.RandomSlice(identity.CentIDLength)


func getServiceWithMockedLayers() documents.Service {

	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(centIDBytes, nil)
	repo := testRepo()
	return documents.DefaultService(c,repo)
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
	
	assert.Equal(t, documentIdentifier,cd.DocumentIdentifier, "should be same document Identifier")
	assert.Equal(t, currentVersion,cd.CurrentVersion, "should be same version")
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


func createMockDocument() (*invoice.Invoice, error) {
	documentIdentifier := utils.RandomSlice(32)
	nextIdentifier := utils.RandomSlice(32)
	inv1 := &invoice.Invoice{
		InvoiceNumber: "test_invoice",
		GrossAmount:   60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     documentIdentifier,
			NextVersion:        nextIdentifier,
		},
	}
	err := testRepo().Create(centIDBytes, documentIdentifier, inv1)
	return inv1, err
}



