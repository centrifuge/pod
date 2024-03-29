//go:build integration

package documents_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	genericpb "github.com/centrifuge/centrifuge-protobufs/gen/go/generic"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/config/configstore"
	"github.com/centrifuge/pod/contextutil"
	protocolIDDispatcher "github.com/centrifuge/pod/dispatcher"
	. "github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/ipfs"
	"github.com/centrifuge/pod/jobs"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	"github.com/centrifuge/pod/notification"
	"github.com/centrifuge/pod/p2p"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pallets/anchors"
	"github.com/centrifuge/pod/pending"
	"github.com/centrifuge/pod/storage"
	"github.com/centrifuge/pod/storage/leveldb"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	jobs2 "github.com/centrifuge/pod/testingutils/jobs"
	"github.com/centrifuge/pod/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&integration_test.Bootstrapper{},
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&configstore.Bootstrapper{},
	&jobs.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&protocolIDDispatcher.Bootstrapper{},
	&v2.AccountTestBootstrapper{},
	Bootstrapper{},
	pending.Bootstrapper{},
	&ipfs.TestBootstrapper{},
	&nftv3.Bootstrapper{},
	&p2p.Bootstrapper{},
	PostBootstrapper{},
}

var (
	storageRepo storage.Repository
	anchorSrv   anchors.API
	registry    *ServiceRegistry
	dispatcher  jobs.Dispatcher
	configSrv   config.Service
	docSrv      Service
	repo        Repository
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	storageRepo = ctx[storage.BootstrappedDB].(storage.Repository)
	anchorSrv = ctx[pallets.BootstrappedAnchorService].(anchors.API)
	dispatcher = ctx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)
	configSrv = ctx[config.BootstrappedConfigStorage].(config.Service)
	docSrv = ctx[BootstrappedDocumentService].(Service)
	repo = ctx[BootstrappedDocumentRepository].(Repository)
	registry = ctx[BootstrappedRegistry].(*ServiceRegistry)

	if err := registry.Register(testDocScheme, &testService{}); err != nil {
		panic(err)
	}

	storageRepo.Register(&testDoc{})
	storageRepo.Register(&doc{})

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_Service_GetCurrentVersion(t *testing.T) {
	documentID := utils.RandomSlice(32)
	documentCurrentVersion := utils.RandomSlice(32)
	documentNextVersion := utils.RandomSlice(32)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	document := &doc{
		SomeString: "Hello, World!",
		DocID:      documentID,
		Current:    documentCurrentVersion,
		Next:       documentNextVersion,
		status:     Committed,
	}

	err = repo.Create(acc.GetIdentity().ToBytes(), documentCurrentVersion, document)
	assert.NoError(t, err)

	res, err := docSrv.GetCurrentVersion(ctx, documentID)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Service_GetCurrentVersion_RepoError(t *testing.T) {
	documentID := utils.RandomSlice(32)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	res, err := docSrv.GetCurrentVersion(ctx, documentID)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_GetVersion(t *testing.T) {
	documentID := utils.RandomSlice(32)
	documentCurrentVersion := utils.RandomSlice(32)
	documentNextVersion := utils.RandomSlice(32)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	document := &doc{
		SomeString: "Hello, World!",
		DocID:      documentID,
		Current:    documentCurrentVersion,
		Next:       documentNextVersion,
		status:     Committed,
	}

	err = repo.Create(acc.GetIdentity().ToBytes(), documentCurrentVersion, document)
	assert.NoError(t, err)

	res, err := docSrv.GetVersion(ctx, documentID, documentCurrentVersion)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Service_GetVersion_RepoError(t *testing.T) {
	documentID := utils.RandomSlice(32)
	documentCurrentVersion := utils.RandomSlice(32)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	res, err := docSrv.GetVersion(ctx, documentID, documentCurrentVersion)
	assert.True(t, errors.IsOfType(ErrDocumentVersionNotFound, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_GetVersion_InvalidVersion(t *testing.T) {
	documentID := utils.RandomSlice(32)
	documentCurrentVersion := utils.RandomSlice(32)
	documentNextVersion := utils.RandomSlice(32)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	document := &doc{
		SomeString: "Hello, World!",
		// This document ID is different from the requested one.
		DocID:   utils.RandomSlice(32),
		Current: documentCurrentVersion,
		Next:    documentNextVersion,
		status:  Committed,
	}

	err = repo.Create(acc.GetIdentity().ToBytes(), documentCurrentVersion, document)
	assert.NoError(t, err)

	res, err := docSrv.GetVersion(ctx, documentID, documentCurrentVersion)
	assert.True(t, errors.IsOfType(ErrDocumentVersionNotFound, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_CreateProofs(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
	}

	jobID, err := docSrv.Commit(ctx, testDoc)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	fields := []string{"cd_tree.document_type"}

	res, err := docSrv.CreateProofs(ctx, testDoc.ID(), fields)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Service_CreateProofs_GetCurrentVersionError(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)
	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	fields := []string{"cd_tree.document_type"}

	res, err := docSrv.CreateProofs(ctx, testDoc.ID(), fields)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_CreateProofs_ValidationError(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docTimestamp := time.Now()

	cd.Document.Timestamp = timestamppb.New(docTimestamp)
	cd.Document.Author = acc.GetIdentity().ToBytes()
	cd.Status = Committed

	signature, err := acc.SignMsg(ConsensusSignaturePayload(signingRoot, false))
	assert.NoError(t, err)

	cd.AppendSignatures(signature)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	// Repository

	repo.Register(testDoc)

	err = repo.Create(acc.GetIdentity().ToBytes(), testDoc.CurrentVersion(), testDoc)
	assert.NoError(t, err)

	fields := []string{"cd_tree.document_type"}

	res, err := docSrv.CreateProofs(ctx, testDoc.ID(), fields)
	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_CreateProofsForVersion(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
	}

	jobID, err := docSrv.Commit(ctx, testDoc)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	fields := []string{"cd_tree.document_type"}

	res, err := docSrv.CreateProofsForVersion(ctx, testDoc.ID(), testDoc.CurrentVersion(), fields)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Service_CreateProofsForVersion_GetVersionError(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docTimestamp := time.Now()

	cd.Document.Timestamp = timestamppb.New(docTimestamp)
	cd.Document.Author = acc.GetIdentity().ToBytes()
	cd.Status = Committed

	signature, err := acc.SignMsg(ConsensusSignaturePayload(signingRoot, false))
	assert.NoError(t, err)

	cd.AppendSignatures(signature)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	// Anchors

	anchorID, err := anchors.ToAnchorID(testDoc.CurrentVersionPreimage())
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	err = anchorSrv.CommitAnchor(ctx, anchorID, docRoot, utils.RandomByte32())
	assert.NoError(t, err)

	fields := []string{"cd_tree.document_type"}

	res, err := docSrv.CreateProofsForVersion(ctx, testDoc.ID(), testDoc.CurrentVersion(), fields)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_CreateProofsForVersion_ValidationError(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docTimestamp := time.Now()

	cd.Document.Timestamp = timestamppb.New(docTimestamp)
	cd.Document.Author = acc.GetIdentity().ToBytes()
	cd.Status = Committed

	signature, err := acc.SignMsg(ConsensusSignaturePayload(signingRoot, false))
	assert.NoError(t, err)

	cd.AppendSignatures(signature)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	// Repository

	repo.Register(testDoc)

	err = repo.Create(acc.GetIdentity().ToBytes(), testDoc.CurrentVersion(), testDoc)
	assert.NoError(t, err)

	fields := []string{"cd_tree.document_type"}

	res, err := docSrv.CreateProofsForVersion(ctx, testDoc.ID(), testDoc.CurrentVersion(), fields)
	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_RequestDocumentSignature(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
	}

	testDoc.AddUpdateLog(acc.GetIdentity())

	sr, err := testDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, false))
	assert.NoError(t, err)

	testDoc.AppendSignatures(sig)

	res, err := docSrv.RequestDocumentSignature(ctx, testDoc, acc.GetIdentity())
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 1)
	assert.False(t, res[0].TransitionValidated)
}

func TestIntegration_Service_RequestDocumentSignature_DocIDAndCurrentVersionMismatch(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	// Current version of the document will be different from the ID, in this case, we will store
	// the document twice - once by using the ID and once by using the current version.

	cd, err = cd.PrepareNewVersion(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
	}

	testDoc.AddUpdateLog(acc.GetIdentity())

	sr, err := testDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, false))
	assert.NoError(t, err)

	testDoc.AppendSignatures(sig)

	res, err := docSrv.RequestDocumentSignature(ctx, testDoc, acc.GetIdentity())
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 1)
	assert.False(t, res[0].TransitionValidated)
}

func TestIntegration_Service_RequestDocumentSignature_OldDocumentPresent(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	oldCd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	oldDoc := &testDoc{
		CoreDocument: oldCd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	oldDoc.AddUpdateLog(acc.GetIdentity())

	sr, err := oldDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, false))
	assert.NoError(t, err)

	oldDoc.AppendSignatures(sig)

	err = repo.Create(acc.GetIdentity().ToBytes(), oldDoc.CurrentVersion(), oldDoc)
	assert.NoError(t, err)

	newCd, err := oldCd.PrepareNewVersion(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	newDoc := &testDoc{
		CoreDocument: newCd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	newDoc.AddUpdateLog(acc.GetIdentity())

	sr, err = newDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	sig, err = acc.SignMsg(ConsensusSignaturePayload(sr, false))
	assert.NoError(t, err)

	newDoc.AppendSignatures(sig)

	res, err := docSrv.RequestDocumentSignature(ctx, newDoc, acc.GetIdentity())
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 1)
	assert.True(t, res[0].TransitionValidated)
}

func TestIntegration_Service_RequestDocumentSignature_ValidationError(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docTimestamp := time.Now()

	cd.Document.Timestamp = timestamppb.New(docTimestamp)
	cd.Document.Author = acc.GetIdentity().ToBytes()
	cd.Status = Pending

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	res, err := docSrv.RequestDocumentSignature(ctx, testDoc, acc.GetIdentity())
	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_RequestDocumentSignature_DocumentCreateError(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	testDoc.AddUpdateLog(acc.GetIdentity())

	sr, err := testDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, false))
	assert.NoError(t, err)

	testDoc.AppendSignatures(sig)

	// Store the document using the current version to ensure that an error is thrown.
	err = repo.Create(acc.GetIdentity().ToBytes(), testDoc.CurrentVersion(), testDoc)
	assert.NoError(t, err)

	res, err := docSrv.RequestDocumentSignature(ctx, testDoc, acc.GetIdentity())
	assert.True(t, errors.IsOfType(ErrDocumentPersistence, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_ReceiveAnchoredDocument(t *testing.T) {
	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
	}

	testDoc.AddUpdateLog(acc.GetIdentity())

	sr, err := testDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, false))
	assert.NoError(t, err)

	testDoc.AppendSignatures(sig)

	// HTTP test server
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		body, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		var notificationMsg notification.Message

		err = json.Unmarshal(body, &notificationMsg)
		assert.NoError(t, err)

		assert.Equal(t, notificationMsg.EventType, notification.EventTypeDocument)
		assert.Equal(t, notificationMsg.Document.ID.Bytes(), testDoc.ID())
		assert.Equal(t, notificationMsg.Document.VersionID.Bytes(), testDoc.CurrentVersion())
		assert.Equal(t, notificationMsg.Document.From.Bytes(), collaborator.ToBytes())
		assert.Equal(t, notificationMsg.Document.To.Bytes(), acc.GetIdentity().ToBytes())
	}))

	defer testServer.Close()

	account, ok := acc.(*configstore.Account)
	assert.True(t, ok)

	account.WebhookURL = testServer.URL

	// Repo
	err = repo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
	assert.NoError(t, err)

	// Anchors
	anchorID, err := anchors.ToAnchorID(testDoc.CurrentVersionPreimage())
	assert.NoError(t, err)

	documentRoot, err := testDoc.CalculateDocumentRoot()
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	err = anchorSrv.CommitAnchor(ctx, anchorID, docRoot, utils.RandomByte32())
	assert.NoError(t, err)

	err = docSrv.ReceiveAnchoredDocument(ctx, testDoc, collaborator)
	assert.NoError(t, err)

	// Sleep to ensure that the notifier is called.
	time.Sleep(1 * time.Second)
}

func TestIntegration_Service_ReceiveAnchoredDocument_OldDocumentPresent(t *testing.T) {
	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	oldCd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{
				acc.GetIdentity(),
				collaborator,
			},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	oldDoc := &testDoc{
		CoreDocument: oldCd,
		Data:         docData,
	}

	err = repo.Create(acc.GetIdentity().ToBytes(), oldDoc.CurrentVersion(), oldDoc)
	assert.NoError(t, err)

	newCd, err := oldCd.PrepareNewVersion(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{
				acc.GetIdentity(),
				collaborator,
			},
		},
		nil,
	)
	assert.NoError(t, err)

	newDoc := &testDoc{
		CoreDocument: newCd,
		Data:         docData,
	}

	newDoc.AddUpdateLog(acc.GetIdentity())

	sr, err := newDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, false))
	assert.NoError(t, err)

	newDoc.AppendSignatures(sig)

	// HTTP test server
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		body, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		var notificationMsg notification.Message

		err = json.Unmarshal(body, &notificationMsg)
		assert.NoError(t, err)

		assert.Equal(t, notificationMsg.EventType, notification.EventTypeDocument)
		assert.Equal(t, notificationMsg.Document.ID.Bytes(), newDoc.ID())
		assert.Equal(t, notificationMsg.Document.VersionID.Bytes(), newDoc.CurrentVersion())
		assert.Equal(t, notificationMsg.Document.From.Bytes(), collaborator.ToBytes())
		assert.Equal(t, notificationMsg.Document.To.Bytes(), acc.GetIdentity().ToBytes())
	}))

	defer testServer.Close()

	account, ok := acc.(*configstore.Account)
	assert.True(t, ok)

	account.WebhookURL = testServer.URL

	// Repo
	err = repo.Create(acc.GetIdentity().ToBytes(), newDoc.CurrentVersion(), newDoc)
	assert.NoError(t, err)

	// Anchors
	anchorID, err := anchors.ToAnchorID(newDoc.CurrentVersionPreimage())
	assert.NoError(t, err)

	documentRoot, err := newDoc.CalculateDocumentRoot()
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	err = anchorSrv.CommitAnchor(ctx, anchorID, docRoot, utils.RandomByte32())
	assert.NoError(t, err)

	err = docSrv.ReceiveAnchoredDocument(ctx, newDoc, collaborator)
	assert.NoError(t, err)

	// Sleep to ensure that the notifier is called.
	time.Sleep(1 * time.Second)
}

func TestIntegration_Service_ReceiveAnchoredDocument_OldDocumentRetrievalError(t *testing.T) {
	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	// Set a previous version to ensure that we will attempt to retrieve it.
	cd.Document.PreviousVersion = utils.RandomSlice(32)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
	}

	testDoc.AddUpdateLog(acc.GetIdentity())

	sr, err := testDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, false))
	assert.NoError(t, err)

	testDoc.AppendSignatures(sig)

	// HTTP test server
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		body, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		var notificationMsg notification.Message

		err = json.Unmarshal(body, &notificationMsg)
		assert.NoError(t, err)

		assert.Equal(t, notificationMsg.EventType, notification.EventTypeDocument)
		assert.Equal(t, notificationMsg.Document.ID.Bytes(), testDoc.ID())
		assert.Equal(t, notificationMsg.Document.VersionID.Bytes(), testDoc.CurrentVersion())
		assert.Equal(t, notificationMsg.Document.From.Bytes(), collaborator.ToBytes())
		assert.Equal(t, notificationMsg.Document.To.Bytes(), acc.GetIdentity().ToBytes())
	}))

	defer testServer.Close()

	account, ok := acc.(*configstore.Account)
	assert.True(t, ok)

	account.WebhookURL = testServer.URL

	// Repo
	err = repo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
	assert.NoError(t, err)

	// Anchors
	anchorID, err := anchors.ToAnchorID(testDoc.CurrentVersionPreimage())
	assert.NoError(t, err)

	documentRoot, err := testDoc.CalculateDocumentRoot()
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	err = anchorSrv.CommitAnchor(ctx, anchorID, docRoot, utils.RandomByte32())
	assert.NoError(t, err)

	err = docSrv.ReceiveAnchoredDocument(ctx, testDoc, collaborator)
	assert.NoError(t, err)

	// Sleep to ensure that the notifier is called.
	time.Sleep(1 * time.Second)
}

func TestIntegration_Service_ReceiveAnchoredDocument_ValidationError(t *testing.T) {
	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docTimestamp := time.Now()

	cd.Document.Timestamp = timestamppb.New(docTimestamp)
	cd.Document.Author = acc.GetIdentity().ToBytes()
	cd.Status = Pending

	signature, err := acc.SignMsg(ConsensusSignaturePayload(signingRoot, false))
	assert.NoError(t, err)

	cd.AppendSignatures(signature)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	err = docSrv.ReceiveAnchoredDocument(ctx, testDoc, collaborator)
	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))
}

func TestIntegration_Service_ReceiveAnchoredDocument_UpdateError(t *testing.T) {
	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	cd, err := NewCoreDocument(
		compactTestDocPrefix(),
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{acc.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
	}

	testDoc.AddUpdateLog(acc.GetIdentity())

	sr, err := testDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, false))
	assert.NoError(t, err)

	testDoc.AppendSignatures(sig)

	// Anchors
	anchorID, err := anchors.ToAnchorID(testDoc.CurrentVersionPreimage())
	assert.NoError(t, err)

	documentRoot, err := testDoc.CalculateDocumentRoot()
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	err = anchorSrv.CommitAnchor(ctx, anchorID, docRoot, utils.RandomByte32())
	assert.NoError(t, err)

	err = docSrv.ReceiveAnchoredDocument(ctx, testDoc, collaborator)
	assert.True(t, errors.IsOfType(ErrDocumentPersistence, err))
}

func TestIntegration_Service_Derive_FromUpdatePayload(t *testing.T) {
	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(compactTestDocPrefix(), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	cd.Status = Committed

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	//repo.Register(testDoc)

	err = repo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
	assert.NoError(t, err)

	payload := UpdatePayload{
		CreatePayload: CreatePayload{
			Scheme: testDocScheme,
		},
		DocumentID: testDoc.ID(),
	}

	expectedResult, err := testDoc.DeriveFromUpdatePayload(ctx, payload)
	assert.NoError(t, err)

	res, err := docSrv.Derive(ctx, payload)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult.ID(), res.ID())
	assert.Equal(t, expectedResult.CurrentVersion(), res.CurrentVersion())
	assert.Equal(t, expectedResult.CurrentVersionPreimage(), res.CurrentVersionPreimage())
	assert.Equal(t, expectedResult.PreviousVersion(), res.PreviousVersion())
	assert.Equal(t, expectedResult.Scheme(), res.Scheme())
	assert.Equal(t, expectedResult.Type(), res.Type())
	assert.NotEqual(t, expectedResult.NextVersion(), res.NextVersion())
	assert.NotEqual(t, expectedResult.NextPreimage(), res.NextPreimage())
}

func TestIntegration_Service_Derive_FromUpdatePayload_CurrentVersionNotFound(t *testing.T) {
	// Document

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(compactTestDocPrefix(), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	cd.Status = Committed

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	payload := UpdatePayload{
		CreatePayload: CreatePayload{
			Scheme: testDocScheme,
		},
		DocumentID: testDoc.ID(),
	}

	res, err := docSrv.Derive(ctx, payload)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestIntegration_Service_Derive_FromCreatePayload(t *testing.T) {
	docData := "test-data"

	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	payload := UpdatePayload{
		CreatePayload: CreatePayload{
			Data:   []byte(docData),
			Scheme: testDocScheme,
		},
	}

	res, err := docSrv.Derive(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, testDocScheme, res.Scheme())
	assert.Equal(t, []byte(docData), res.GetData())
}

func TestIntegration_Service_DeriveClone(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(compactTestDocPrefix(), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	cd.Status = Committed

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	//repo.Register(testDoc)

	err = repo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
	assert.NoError(t, err)

	payload := ClonePayload{
		Scheme:     testDocScheme,
		TemplateID: testDoc.ID(),
	}

	res, err := docSrv.DeriveClone(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Service_DeriveClone_DocumentNotFound(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(compactTestDocPrefix(), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	cd.Status = Committed

	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	payload := ClonePayload{
		Scheme:     testDocScheme,
		TemplateID: testDoc.ID(),
	}

	res, err := docSrv.DeriveClone(ctx, payload)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_Commit_UpdateExistingDoc(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(compactTestDocPrefix(), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	cd.Status = Pending
	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	// Store the current doc to ensure that an update is triggered.
	err = repo.Create(acc.GetIdentity().ToBytes(), testDoc.ID(), testDoc)
	assert.NoError(t, err)

	res, err := docSrv.Commit(ctx, testDoc)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), res)
	assert.NoError(t, err)

	testDoc.Document.PreviousVersion = testDoc.Document.CurrentVersion
	testDoc.Document.CurrentVersion = testDoc.Document.NextVersion
	testDoc.Document.CurrentPreimage = testDoc.Document.NextPreimage
	testDoc.Document.NextVersion = utils.RandomSlice(32)
	testDoc.Document.NextPreimage = utils.RandomSlice(32)

	res, err = docSrv.Commit(ctx, testDoc)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), res)
	assert.NoError(t, err)
}

func TestIntegration_Service_Commit_CreateNewDoc(t *testing.T) {
	accs, err := configSrv.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(compactTestDocPrefix(), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	cd.Status = Pending
	docData := "test-data"

	testDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	// Since this document is not stored, we will store it when committing.
	jobID, err := docSrv.Commit(ctx, testDoc)
	assert.NoError(t, err)
	assert.NotNil(t, jobID)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	testDoc.Document.PreviousVersion = testDoc.Document.CurrentVersion
	testDoc.Document.CurrentVersion = testDoc.Document.NextVersion
	testDoc.Document.CurrentPreimage = testDoc.Document.NextPreimage
	testDoc.Document.NextVersion = utils.RandomSlice(32)
	testDoc.Document.NextPreimage = utils.RandomSlice(32)

	jobID, err = docSrv.Commit(ctx, testDoc)
	assert.NoError(t, err)
	assert.NotNil(t, jobID)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)
}

func TestIntegration_Service_DeriveFromCoreDocument(t *testing.T) {
	pbcd := &coredocumentpb.CoreDocument{}
	pbcd.EmbeddedData = &anypb.Any{TypeUrl: testDocScheme}

	res, err := docSrv.DeriveFromCoreDocument(pbcd)
	assert.NoError(t, err)

	cd, err := NewCoreDocumentFromProtobuf(pbcd)
	assert.NoError(t, err)

	testDoc, ok := res.(*testDoc)
	assert.True(t, ok)
	assert.Equal(t, cd, testDoc.CoreDocument)
}

func TestIntegration_Service_New(t *testing.T) {
	res, err := docSrv.New(testDocScheme)
	assert.NoError(t, err)
	assert.IsType(t, &testDoc{}, res)
}

func TestIntegration_Service_Validate(t *testing.T) {
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	cd, err := NewCoreDocument(compactTestDocPrefix(), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	docData := "test-data"

	newDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	ctx := context.Background()

	err = docSrv.Validate(ctx, newDoc, nil)
	assert.NoError(t, err)
	assert.True(t, newDoc.Validated)
}

func TestIntegration_Service_Validate_WithOldDocument(t *testing.T) {
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)
	docData := "test-data"

	cd, err := NewCoreDocument(compactTestDocPrefix(), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	oldDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	cd, err = NewCoreDocument(compactTestDocPrefix(), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	newDoc := &testDoc{
		CoreDocument: cd,
		Data:         docData,
		SigningRoot:  signingRoot,
		DocumentRoot: documentRoot,
	}

	newDoc.Document.DocumentIdentifier = oldDoc.Document.DocumentIdentifier
	newDoc.Document.PreviousVersion = oldDoc.Document.CurrentVersion
	newDoc.Document.CurrentVersion = oldDoc.Document.NextVersion

	ctx := context.Background()

	err = docSrv.Validate(ctx, newDoc, oldDoc)
	assert.NoError(t, err)
	assert.True(t, newDoc.Validated)
}

type testService struct {
	Service
}

func (s testService) DeriveFromCoreDocument(pbcd *coredocumentpb.CoreDocument) (Document, error) {
	cd, err := NewCoreDocumentFromProtobuf(pbcd)

	if err != nil {
		return nil, err
	}

	return &testDoc{CoreDocument: cd}, nil
}

func (s testService) New(_ string) (Document, error) {
	return &testDoc{}, nil
}

func (s testService) Validate(_ context.Context, new Document, old Document) error {
	testDocNew, ok := new.(*testDoc)

	if !ok {
		return fmt.Errorf("expected *testDoc, got %T", new)
	}

	testDocNew.Validated = true

	if old == nil {
		return nil
	}

	testDocOld, ok := old.(*testDoc)

	if !ok {
		return fmt.Errorf("expected *testDoc, got %T", old)
	}

	testDocOld.Validated = true

	return nil
}

const (
	testDocPrefix = "test-doc-prefix"
	testDocScheme = "test-doc"
	testDocType   = "test-doc-type"
)

func compactTestDocPrefix() []byte {
	return []byte{0, 5, 0, 0}
}

type testDoc struct {
	*CoreDocument
	Data         any
	SigningRoot  []byte
	DocumentRoot []byte
	Validated    bool
}

func (t testDoc) Type() reflect.Type {
	return reflect.TypeOf(t)
}

func (t *testDoc) JSON() ([]byte, error) {
	return json.Marshal(t)
}

func (t *testDoc) FromJSON(j []byte) error {
	return json.Unmarshal(j, t)
}

func (t *testDoc) DocumentType() string {
	return testDocType
}

func (t *testDoc) Scheme() string {
	return testDocScheme
}

func (t *testDoc) GetData() any {
	return t.Data
}

func (t *testDoc) UnpackCoreDocument(cd *coredocumentpb.CoreDocument) (err error) {
	t.CoreDocument, err = NewCoreDocumentFromProtobuf(cd)
	return err
}

func (t *testDoc) PackCoreDocument() (cd *coredocumentpb.CoreDocument, err error) {
	data, err := proto.Marshal(getProtoGenericData())
	if err != nil {
		return cd, errors.New("couldn't serialise GenericData: %v", err)
	}

	embedData := &anypb.Any{
		TypeUrl: t.DocumentType(),
		Value:   data,
	}
	return t.CoreDocument.PackCoreDocument(embedData), nil
}

func (t *testDoc) CalculateSigningRoot() ([]byte, error) {
	dataLeaves, err := t.getDataLeaves()
	if err != nil {
		return nil, err
	}

	return t.CoreDocument.CalculateSigningRoot(t.DocumentType(), dataLeaves)
}

func (t *testDoc) CalculateDocumentRoot() ([]byte, error) {
	dataLeaves, err := t.getDataLeaves()
	if err != nil {
		return nil, err
	}

	return t.CoreDocument.CalculateDocumentRoot(t.DocumentType(), dataLeaves)
}

func (t *testDoc) Patch(payload UpdatePayload) error {
	ncd, err := t.CoreDocument.Patch([]byte(testDocPrefix), payload.Collaborators, payload.Attributes)
	if err != nil {
		return err
	}

	t.CoreDocument = ncd
	return nil
}

func (t *testDoc) AddAttributes(ca CollaboratorsAccess, prepareNewVersion bool, attrs ...Attribute) error {
	ncd, err := t.CoreDocument.AddAttributes(ca, prepareNewVersion, []byte(testDocPrefix), attrs...)
	if err != nil {
		return err
	}

	t.CoreDocument = ncd
	return nil
}

func (t *testDoc) AddNFT(grantReadAccess bool, collectionID types.U64, itemID types.U128) error {
	cd, err := t.CoreDocument.AddNFT(grantReadAccess, collectionID, itemID)
	if err != nil {
		return err
	}

	t.CoreDocument = cd
	return nil
}

func (t *testDoc) DeleteAttribute(key AttrKey, prepareNewVersion bool) error {
	ncd, err := t.CoreDocument.DeleteAttribute(key, prepareNewVersion, []byte(testDocPrefix))
	if err != nil {
		return err
	}

	t.CoreDocument = ncd
	return nil
}

func (t *testDoc) CreateProofs(fields []string) (prf *DocumentProof, err error) {
	dataLeaves, err := t.getDataLeaves()
	if err != nil {
		return nil, err
	}

	return t.CoreDocument.CreateProofs(t.DocumentType(), dataLeaves, fields)
}

func (t *testDoc) CollaboratorCanUpdate(updated Document, collaborator *types.AccountID) error {
	newDoc, ok := updated.(*testDoc)
	if !ok {
		return errors.New("expecting a test doc but got %T", updated)
	}

	// check the core document changes
	err := t.CoreDocument.CollaboratorCanUpdate(newDoc.CoreDocument, collaborator, t.DocumentType())
	if err != nil {
		return err
	}

	// check generic doc specific changes
	oldTree, err := t.getDocumentDataTree()
	if err != nil {
		return err
	}

	newTree, err := newDoc.getDocumentDataTree()
	if err != nil {
		return err
	}

	rules := t.CoreDocument.TransitionRulesFor(collaborator)
	cf := GetChangedFields(oldTree, newTree)
	return ValidateTransitions(rules, cf)
}

func (t *testDoc) getDataLeaves() ([]proofs.LeafNode, error) {
	tree, err := t.getRawDataTree()
	if err != nil {
		return nil, err
	}
	return tree.GetLeaves(), nil
}

func (t *testDoc) getRawDataTree() (*proofs.DocumentTree, error) {
	if t.CoreDocument == nil {
		return nil, errors.New("getDataTree error CoreDocument not set")
	}

	tree, err := t.CoreDocument.DefaultTreeWithPrefix(testDocPrefix, compactTestDocPrefix())
	if err != nil {
		return nil, err
	}

	err = tree.AddLeavesFromDocument(getProtoGenericData())
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func (t *testDoc) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	if t.CoreDocument == nil {
		return nil, errors.New("getDocumentDataTree error CoreDocument not set")
	}

	tree, err = t.CoreDocument.DefaultTreeWithPrefix(testDocPrefix, compactTestDocPrefix())
	if err != nil {
		return nil, err
	}

	err = tree.AddLeavesFromDocument(getProtoGenericData())
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}

	err = tree.Generate()
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}

	return tree, nil
}

func getProtoGenericData() *genericpb.GenericData {
	return &genericpb.GenericData{
		Scheme: []byte(testDocScheme),
	}
}

func (t *testDoc) DeriveFromCreatePayload(_ context.Context, payload CreatePayload) error {
	cd, err := NewCoreDocument(compactTestDocPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return err
	}

	t.Data = payload.Data

	t.CoreDocument = cd
	return nil
}

func (t *testDoc) DeriveFromClonePayload(_ context.Context, m Document) error {
	d, err := m.PackCoreDocument()
	if err != nil {
		return err
	}

	cd, err := NewClonedDocument(d)
	if err != nil {
		return err
	}

	t.CoreDocument = cd
	return nil
}

func (t *testDoc) DeriveFromUpdatePayload(_ context.Context, payload UpdatePayload) (Document, error) {
	ncd, err := t.CoreDocument.PrepareNewVersion([]byte(testDocPrefix), payload.Collaborators, payload.Attributes)
	if err != nil {
		return nil, err
	}

	return &testDoc{
		CoreDocument: ncd,
	}, nil
}

func TestIntegration_Repo_Exists(t *testing.T) {
	repo := NewDBRepository(storageRepo)

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	res := repo.Exists(accountID, documentID)
	assert.False(t, res)

	key := GetKey(accountID, documentID)

	doc := &doc{}

	storageRepo.Register(doc)

	err := storageRepo.Create(key, doc)
	assert.NoError(t, err)

	res = repo.Exists(accountID, documentID)
	assert.True(t, res)
}

func TestIntegration_Repo_Get(t *testing.T) {
	repo := NewDBRepository(storageRepo)

	accountID := utils.RandomSlice(32)
	documentID := utils.RandomSlice(32)

	res, err := repo.Get(accountID, documentID)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	key := GetKey(accountID, documentID)

	doc := &doc{}

	storageRepo.Register(doc)

	err = storageRepo.Create(key, doc)
	assert.NoError(t, err)

	res, err = repo.Get(accountID, documentID)
	assert.NoError(t, err)
	assert.Equal(t, doc, res)
}

type doc struct {
	Document
	DocID, Current, Next []byte
	SomeString           string `json:"some_string"`
	Time                 time.Time
	status               Status
}

func (m *doc) ID() []byte {
	return m.DocID
}

func (m *doc) CurrentVersion() []byte {
	return m.Current
}

func (m *doc) NextVersion() []byte {
	return m.Next
}

func (m *doc) JSON() ([]byte, error) {
	return json.Marshal(m)
}

func (m *doc) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *doc) Type() reflect.Type {
	return reflect.TypeOf(m)
}

func (m *doc) Timestamp() (time.Time, error) {
	return m.Time, nil
}

func (m *doc) GetStatus() Status {
	return m.status
}

type unknownDoc struct {
	SomeString string `json:"some_string"`
}

func (unknownDoc) Type() reflect.Type {
	return reflect.TypeOf(unknownDoc{})
}

func (u *unknownDoc) JSON() ([]byte, error) {
	return json.Marshal(u)
}

func (u *unknownDoc) FromJSON(j []byte) error {
	return json.Unmarshal(j, u)
}
