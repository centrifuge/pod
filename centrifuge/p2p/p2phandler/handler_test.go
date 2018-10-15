// +build unit

package p2phandler

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	cented25519 "github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/centrifuge/go-centrifuge/centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ed25519"
)

var (
	key1Pub = [...]byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1    = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	handler = Handler{Notifier: &MockWebhookSender{}}
)

// MockWebhookSender implements notification.Sender
type MockWebhookSender struct{}

func (wh *MockWebhookSender) Send(notification *notificationpb.NotificationMessage) (status notification.NotificationStatus, err error) {
	return
}

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	coredocumentrepository.InitLevelDBRepository(cc.GetLevelDBStorage())
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

var coreDoc = testingutils.GenerateCoreDocument()

func TestP2PService(t *testing.T) {
	req := p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID()}
	res, err := handler.Post(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.Equal(t, res.Document.DocumentIdentifier, coreDoc.DocumentIdentifier, "Incorrect identifier")

	doc := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(coreDoc.DocumentIdentifier, doc)
	assert.Equal(t, doc.DocumentIdentifier, coreDoc.DocumentIdentifier, "Document Identifier doesn't match")
}

func TestP2PService_IncompatibleRequest(t *testing.T) {
	// Test invalid version
	req := p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: "1000.0.0-invalid", NetworkIdentifier: config.Config.GetNetworkID()}
	res, err := handler.Post(context.Background(), &req)

	assert.Error(t, err)
	p2perr, _ := centerrors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.VersionMismatch)
	assert.Nil(t, res)

	// Test invalid network
	req = p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID() + 1}
	res, err = handler.Post(context.Background(), &req)

	assert.Error(t, err)
	p2perr, _ = centerrors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.NetworkMismatch)
	assert.Nil(t, res)
}

func TestP2PService_HandleP2PPostNilDocument(t *testing.T) {
	req := p2ppb.P2PMessage{CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID()}
	res, err := handler.Post(context.Background(), &req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestHandler_RequestDocumentSignature_nilDocument(t *testing.T) {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID(),
	}}

	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Error(t, err, "must return error")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_RequestDocumentSignature_version_fail(t *testing.T) {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: "1000.0.1-invalid", NetworkIdentifier: config.Config.GetNetworkID(),
	}}

	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")
}

func getSignatureRequest() *p2ppb.SignatureRequest {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID(),
	}, Document: testingutils.GenerateCoreDocument()}

	return req
}

func TestHandler_RequestDocumentSignature_verification_fail(t *testing.T) {
	req := getSignatureRequest()
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.NotNil(t, err, "must be non nil")
	assert.Nil(t, resp, "must be nil")
	assert.Contains(t, err.Error(), "signing root missing")
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	idConfig, err := cented25519.GetIDConfig()
	assert.Nil(t, err)
	sig := &coredocumentpb.Signature{
		EntityId:  idConfig.ID,
		PublicKey: key1Pub[:],
	}
	centID, _ := identity.ToCentID(sig.EntityId)
	idkey := &identity.EthereumIdentityKey{
		Key:       key1Pub,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	id := &testingcommons.MockID{}
	srv := &testingcommons.MockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", key1Pub[:]).Return(idkey, nil).Once()
	identity.IDService = srv
	doc := testingutils.GenerateCoreDocument()
	tree, _ := coredocument.GetDocumentSigningTree(doc)
	doc.SigningRoot = tree.RootHash()
	sig = signatures.Sign(&config.IdentityConfig{
		ID:         sig.EntityId,
		PublicKey:  key1Pub[:],
		PrivateKey: key1,
	}, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	req := getSignatureRequest()
	req.Document = doc
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig = resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestSendAnchoredDocument_IncompatibleRequest(t *testing.T) {
	// Test invalid version
	header := &p2ppb.CentrifugeHeader{
		CentNodeVersion:   "1000.0.0-invalid",
		NetworkIdentifier: config.Config.GetNetworkID(),
	}
	req := p2ppb.AnchDocumentRequest{Document: coreDoc, Header: header}
	res, err := handler.SendAnchoredDocument(context.Background(), &req)
	assert.Error(t, err)
	p2perr, _ := centerrors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.VersionMismatch)
	assert.Nil(t, res)

	// Test invalid network
	header.NetworkIdentifier = config.Config.GetNetworkID() + 1
	header.CentNodeVersion = version.GetVersion().String()
	res, err = handler.SendAnchoredDocument(context.Background(), &req)
	assert.Error(t, err)
	p2perr, _ = centerrors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.NetworkMismatch)
	assert.Nil(t, res)
}

func TestSendAnchoredDocument_NilDocument(t *testing.T) {
	header := &p2ppb.CentrifugeHeader{
		CentNodeVersion:   version.GetVersion().String(),
		NetworkIdentifier: config.Config.GetNetworkID(),
	}
	req := p2ppb.AnchDocumentRequest{Header: header}
	res, err := handler.SendAnchoredDocument(context.Background(), &req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestHandler_SendAnchoredDocument_getModelAndRepo_fail(t *testing.T) {
	req := &p2ppb.AnchDocumentRequest{
		Header: &p2ppb.CentrifugeHeader{
			CentNodeVersion:   version.GetVersion().String(),
			NetworkIdentifier: config.Config.GetNetworkID(),
		},
		Document: coredocument.New(),
	}

	res, err := handler.SendAnchoredDocument(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get type of the document")
	assert.Nil(t, res)
}

func TestHandler_SendAnchoredDocument_update_fail(t *testing.T) {
	req := &p2ppb.AnchDocumentRequest{
		Header: &p2ppb.CentrifugeHeader{
			CentNodeVersion:   version.GetVersion().String(),
			NetworkIdentifier: config.Config.GetNetworkID(),
		},
	}

	cd := coredocument.New()
	cd.EmbeddedData = &any.Any{
		TypeUrl: "update_fail_type",
		Value:   []byte("some data"),
	}
	req.Document = cd
	model := mockModel{}
	srv := mockService{}
	repo := mockRepo{}
	repo.On("Update", cd.CurrentVersion, model).Return(fmt.Errorf("update failed")).Once()
	srv.On("DeriveFromCoreDocument", cd).Return(model, nil).Once()
	srv.On("Repository").Return(repo).Once()
	err := documents.GetRegistryInstance().Register(cd.EmbeddedData.TypeUrl, srv)
	assert.Nil(t, err)
	res, err := handler.SendAnchoredDocument(context.Background(), req)
	repo.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
	assert.Nil(t, res)
}

func TestHandler_SendAnchoredDocument(t *testing.T) {
	req := &p2ppb.AnchDocumentRequest{
		Header: &p2ppb.CentrifugeHeader{
			CentNodeVersion:   version.GetVersion().String(),
			NetworkIdentifier: config.Config.GetNetworkID(),
		},
	}

	cd := coredocument.New()
	cd.EmbeddedData = &any.Any{
		TypeUrl: "send_doc_type",
		Value:   []byte("some data"),
	}
	req.Document = cd

	model := mockModel{}
	srv := mockService{}
	repo := mockRepo{}
	repo.On("Update", cd.CurrentVersion, model).Return(nil).Once()
	srv.On("DeriveFromCoreDocument", cd).Return(model, nil).Once()
	srv.On("Repository").Return(repo).Once()
	err := documents.GetRegistryInstance().Register(cd.EmbeddedData.TypeUrl, srv)
	assert.Nil(t, err)
	res, err := handler.SendAnchoredDocument(context.Background(), req)
	repo.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.Accepted)
}

func TestP2PService_basicChecks(t *testing.T) {
	tests := []struct {
		version   string
		networkID uint32
		err       error
	}{
		{
			version:   "someversion",
			networkID: 12,
			err:       version.IncompatibleVersionError("someversion"),
		},

		{
			version:   "0.0.1",
			networkID: 12,
			err:       incompatibleNetworkError(12),
		},

		{
			version:   version.GetVersion().String(),
			networkID: config.Config.GetNetworkID(),
		},
	}

	for _, c := range tests {
		err := basicChecks(c.version, c.networkID)
		if err != nil {
			if c.err == nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			assert.EqualError(t, err, c.err.Error(), "error mismatch")
		}
	}

}

type mockRepo struct {
	mock.Mock
	documents.Repository
}

func (r mockRepo) Update(id []byte, m documents.Model) error {
	args := r.Called(id, m)
	return args.Error(0)
}

type mockModel struct {
	mock.Mock
	documents.Model
}

type mockService struct {
	mock.Mock
	documents.Service
}

func (s mockService) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	args := s.Called(cd)
	m, _ := args.Get(0).(documents.Model)
	return m, args.Error(1)
}

func (s mockService) Repository() documents.Repository {
	args := s.Called()
	return args.Get(0).(documents.Repository)
}

func Test_getModelAndRepo(t *testing.T) {
	// docType fetch fail
	cd := coredocument.New()
	m, r, err := getModelAndRepo(cd)
	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "failed to get type of the document")

	// missing service
	cd.EmbeddedData = &any.Any{
		TypeUrl: "model_type_fail",
		Value:   []byte("some data"),
	}
	m, r, err = getModelAndRepo(cd)
	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "failed to locate the service")

	// derive fails
	reg := documents.GetRegistryInstance()
	srv := mockService{}
	srv.On("DeriveFromCoreDocument", cd).Return(nil, fmt.Errorf("error")).Once()
	err = reg.Register(cd.EmbeddedData.TypeUrl, srv)
	assert.Nil(t, err)
	m, r, err = getModelAndRepo(cd)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "failed to derive model from core document")

	// success
	model := &mockModel{}
	cd.EmbeddedData.TypeUrl = "get_model_type"
	srv = mockService{}
	repo := mockRepo{}
	srv.On("DeriveFromCoreDocument", cd).Return(model, nil).Once()
	srv.On("Repository").Return(repo).Once()
	err = reg.Register(cd.EmbeddedData.TypeUrl, srv)
	assert.Nil(t, err)
	m, r, err = getModelAndRepo(cd)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.NotNil(t, m)
	assert.Equal(t, model, m)
}
