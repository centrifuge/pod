// +build unit

package p2phandler

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	handler = Handler{Notifier: &MockWebhookSender{}}
)

// MockWebhookSender implements notification.Sender
type MockWebhookSender struct{}

func (wh *MockWebhookSender) Send(notification *notificationpb.NotificationMessage) (status notification.NotificationStatus, err error) {
	return
}

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&coredocumentrepository.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
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

func TestHandler_SendAnchoredDocument_getServiceAndModel_fail(t *testing.T) {
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

func Test_getServiceAndModel(t *testing.T) {
	// document nil fail
	s, m, err := getServiceAndModel(nil)
	assert.Error(t, err)

	// docType fetch fail
	cd := coredocument.New()
	s, m, err = getServiceAndModel(cd)
	assert.Error(t, err)
	assert.Nil(t, s)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "failed to get type of the document")

	// missing service
	cd.EmbeddedData = &any.Any{
		TypeUrl: "model_type_fail",
		Value:   []byte("some data"),
	}
	s, m, err = getServiceAndModel(cd)
	assert.Error(t, err)
	assert.Nil(t, s)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "failed to locate the service")

	// derive fails
	reg := documents.GetRegistryInstance()
	srv := mockService{}
	srv.On("DeriveFromCoreDocument", cd).Return(nil, fmt.Errorf("error")).Once()
	err = reg.Register(cd.EmbeddedData.TypeUrl, srv)
	assert.Nil(t, err)
	s, m, err = getServiceAndModel(cd)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, s)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "failed to derive model from core document")

	// success
	model := &mockModel{}
	cd.EmbeddedData.TypeUrl = "get_model_type"
	srv = mockService{}
	srv.On("DeriveFromCoreDocument", cd).Return(model, nil).Once()
	err = reg.Register(cd.EmbeddedData.TypeUrl, srv)
	assert.Nil(t, err)
	s, m, err = getServiceAndModel(cd)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.NotNil(t, m)
	assert.Equal(t, model, m)
}
