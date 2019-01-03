// +build unit

package genericdoc

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/common"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testRepoGlobal documents.Repository
var (
	centIDBytes = utils.RandomSlice(identity.CentIDLength)
	tenantID    = common.DummyIdentity.Bytes()
	key1Pub     = [...]byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1        = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

func TestMain(m *testing.M) {
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestService_ReceiveAnchoredDocument(t *testing.T) {
	poSrv := service{}
	err := poSrv.ReceiveAnchoredDocument(nil, nil)
	assert.Error(t, err)
}

func getServiceWithMockedLayers() (documents.Service, testingcommons.MockIDService) {
	repo := testRepo()
	idService := testingcommons.MockIDService{}
	idService.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil)
	return DefaultService(nil, repo, &mockAnchorRepo{}, &idService), idService
}

type mockAnchorRepo struct {
	mock.Mock
	anchors.AnchorRepository
}

func (r *mockAnchorRepo) GetDocumentRootOf(anchorID anchors.AnchorID) (anchors.DocumentRoot, error) {
	args := r.Called(anchorID)
	docRoot, _ := args.Get(0).(anchors.DocumentRoot)
	return docRoot, args.Error(1)
}

func createAnchoredMockDocument(t *testing.T, skipSave bool) (*invoice.Invoice, error) {
	i := &invoice.Invoice{
		InvoiceNumber: "test_invoice",
		GrossAmount:   60,
		CoreDocument:  coredocument.New(),
	}
	err := i.CalculateDataRoot()
	if err != nil {
		return nil, err
	}
	// get the coreDoc for the invoice
	corDoc, err := i.PackCoreDocument()
	if err != nil {
		return nil, err
	}
	assert.Nil(t, coredocument.FillSalts(corDoc))
	err = coredocument.CalculateSigningRoot(corDoc)
	if err != nil {
		return nil, err
	}

	centID, err := identity.ToCentID(centIDBytes)
	assert.Nil(t, err)
	signKey := identity.IDKey{
		PublicKey:  key1Pub[:],
		PrivateKey: key1,
	}
	idConfig := &identity.IDConfig{
		ID: centID,
		Keys: map[int]identity.IDKey{
			identity.KeyPurposeSigning: signKey,
		},
	}

	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, corDoc.SigningRoot)

	corDoc.Signatures = append(corDoc.Signatures, sig)

	err = coredocument.CalculateDocumentRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = i.UnpackCoreDocument(corDoc)
	if err != nil {
		return nil, err
	}

	if !skipSave {
		err = testRepo().Create(tenantID, i.CoreDocument.CurrentVersion, i)
		if err != nil {
			return nil, err
		}
	}

	return i, nil
}

func updatedAnchoredMockDocument(t *testing.T, i *invoice.Invoice) (*invoice.Invoice, error) {
	i.GrossAmount = 50
	err := i.CalculateDataRoot()
	if err != nil {
		return nil, err
	}
	// get the coreDoc for the invoice
	corDoc, err := i.PackCoreDocument()
	if err != nil {
		return nil, err
	}
	// hacky update to version
	corDoc.CurrentVersion = corDoc.NextVersion
	corDoc.NextVersion = utils.RandomSlice(32)
	if err != nil {
		return nil, err
	}
	err = coredocument.CalculateSigningRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = coredocument.CalculateDocumentRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = i.UnpackCoreDocument(corDoc)
	if err != nil {
		return nil, err
	}
	err = testRepo().Create(tenantID, i.CoreDocument.CurrentVersion, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Functions returns service mocks
func mockSignatureCheck(i *invoice.Invoice, idService testingcommons.MockIDService, s documents.Service) testingcommons.MockIDService {
	idkey := &identity.EthereumIdentityKey{
		Key:       key1Pub,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	anchorID, _ := anchors.ToAnchorID(i.CoreDocument.DocumentIdentifier)
	docRoot, _ := anchors.ToDocumentRoot(i.CoreDocument.DocumentRoot)
	mockRepo := s.(service).anchorRepository.(*mockAnchorRepo)
	mockRepo.On("GetDocumentRootOf", anchorID).Return(docRoot, nil).Once()
	id := &testingcommons.MockID{}
	centID, _ := identity.ToCentID(centIDBytes)
	idService.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", key1Pub[:]).Return(idkey, nil).Once()
	return idService
}

func TestService_CreateProofs(t *testing.T) {
	service, idService := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	idService = mockSignatureCheck(i, idService, service)
	proof, err := service.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"invoice.invoice_number"})
	assert.Nil(t, err)
	assert.Equal(t, i.CoreDocument.DocumentIdentifier, proof.DocumentID)
	assert.Equal(t, i.CoreDocument.DocumentIdentifier, proof.VersionID)
	assert.Equal(t, len(proof.FieldProofs), 1)
	assert.Equal(t, proof.FieldProofs[0].GetReadableName(), "invoice.invoice_number")
}
func TestService_CreateProofsValidationFails(t *testing.T) {
	service, idService := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	i.CoreDocument.SigningRoot = nil
	err = testRepo().Update(tenantID, i.CoreDocument.CurrentVersion, i)
	assert.Nil(t, err)
	idService = mockSignatureCheck(i, idService, service)
	_, err = service.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"invoice.invoice_number"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "signing root missing")
}

func TestService_CreateProofsInvalidField(t *testing.T) {
	service, idService := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	idService = mockSignatureCheck(i, idService, service)
	_, err = service.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"invalid_field"})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentProof, err))
}

func TestService_CreateProofsDocumentDoesntExist(t *testing.T) {
	service, _ := getServiceWithMockedLayers()
	_, err := service.CreateProofs(utils.RandomSlice(32), []string{"invoice.invoice_number"})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))
}

func TestService_CreateProofsForVersion(t *testing.T) {
	service, idService := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	idService = mockSignatureCheck(i, idService, service)
	olderVersion := i.CoreDocument.CurrentVersion
	i, err = updatedAnchoredMockDocument(t, i)
	assert.Nil(t, err)
	proof, err := service.CreateProofsForVersion(i.CoreDocument.DocumentIdentifier, olderVersion, []string{"invoice.invoice_number"})
	assert.Nil(t, err)
	assert.Equal(t, i.CoreDocument.DocumentIdentifier, proof.DocumentID)
	assert.Equal(t, olderVersion, proof.VersionID)
	assert.Equal(t, len(proof.FieldProofs), 1)
	assert.Equal(t, proof.FieldProofs[0].GetReadableName(), "invoice.invoice_number")
}

func TestService_RequestDocumentSignature_SigningRootNil(t *testing.T) {
	service, idService := getServiceWithMockedLayers()
	i, err := createAnchoredMockDocument(t, true)
	assert.Nil(t, err)
	idService = mockSignatureCheck(i, idService, service)
	i.CoreDocument.SigningRoot = nil
	ctxh, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	signature, err := service.RequestDocumentSignature(ctxh, i)
	assert.NotNil(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))
	assert.Nil(t, signature)
}

func TestService_CreateProofsForVersionDocumentDoesntExist(t *testing.T) {
	i, err := createAnchoredMockDocument(t, false)
	s, _ := getServiceWithMockedLayers()
	assert.Nil(t, err)
	_, err = s.CreateProofsForVersion(i.CoreDocument.DocumentIdentifier, utils.RandomSlice(32), []string{"invoice.invoice_number"})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentVersionNotFound, err))
}

func TestService_GetCurrentVersion_successful(t *testing.T) {
	service, _ := getServiceWithMockedLayers()
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

		err := testRepo().Create(tenantID, version, inv)
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
	service, _ := getServiceWithMockedLayers()
	documentIdentifier := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	inv := &invoice.Invoice{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     currentVersion,
		},
	}

	err := testRepo().Create(tenantID, currentVersion, inv)
	assert.Nil(t, err)

	mod, err := service.GetVersion(documentIdentifier, currentVersion)
	assert.Nil(t, err)

	cd, err := mod.PackCoreDocument()
	assert.Nil(t, err)

	assert.Equal(t, documentIdentifier, cd.DocumentIdentifier, "should be same document Identifier")
	assert.Equal(t, currentVersion, cd.CurrentVersion, "should be same version")
}

func TestService_GetCurrentVersion_error(t *testing.T) {
	service, _ := getServiceWithMockedLayers()
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

	err = testRepo().Create(tenantID, documentIdentifier, inv)
	assert.Nil(t, err)

	_, err = service.GetCurrentVersion(documentIdentifier)
	assert.Nil(t, err)

}

func TestService_GetVersion_error(t *testing.T) {
	service, _ := getServiceWithMockedLayers()

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
	err = testRepo().Create(tenantID, currentVersion, inv)
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

func TestService_Exists(t *testing.T) {
	service, _ := getServiceWithMockedLayers()
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

	err = testRepo().Create(tenantID, documentIdentifier, inv)

	exists := service.Exists(documentIdentifier)
	assert.True(t, exists, "document should exist")

	exists = service.Exists(utils.RandomSlice(32))
	assert.False(t, exists, "document should not exist")

}
