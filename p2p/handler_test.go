// +build unit

package p2p

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/coredocument"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	handler = Handler{}
)

var coreDoc = testingcoredocument.GenerateCoreDocument()

func TestHandler_RequestDocumentSignature_nilDocument(t *testing.T) {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config().GetNetworkID(),
	}}

	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Error(t, err, "must return error")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_RequestDocumentSignature_version_fail(t *testing.T) {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: "1000.0.1-invalid", NetworkIdentifier: config.Config().GetNetworkID(),
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
		NetworkIdentifier: config.Config().GetNetworkID(),
	}
	req := p2ppb.AnchorDocumentRequest{Document: coreDoc, Header: header}
	res, err := handler.SendAnchoredDocument(context.Background(), &req)
	assert.Error(t, err)
	p2perr, _ := centerrors.FromError(err)
	assert.Contains(t, p2perr.Message(), strconv.Itoa(int(code.VersionMismatch)))
	assert.Nil(t, res)

	// Test invalid network
	header.NetworkIdentifier = config.Config().GetNetworkID() + 1
	header.CentNodeVersion = version.GetVersion().String()
	res, err = handler.SendAnchoredDocument(context.Background(), &req)
	assert.Error(t, err)
	p2perr, _ = centerrors.FromError(err)
	assert.Contains(t, p2perr.Message(), strconv.Itoa(int(code.NetworkMismatch)))
	assert.Nil(t, res)
}

func TestSendAnchoredDocument_NilDocument(t *testing.T) {
	header := &p2ppb.CentrifugeHeader{
		CentNodeVersion:   version.GetVersion().String(),
		NetworkIdentifier: config.Config().GetNetworkID(),
	}
	req := p2ppb.AnchorDocumentRequest{Header: header}
	res, err := handler.SendAnchoredDocument(context.Background(), &req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestHandler_SendAnchoredDocument_getServiceAndModel_fail(t *testing.T) {
	req := &p2ppb.AnchorDocumentRequest{
		Header: &p2ppb.CentrifugeHeader{
			CentNodeVersion:   version.GetVersion().String(),
			NetworkIdentifier: config.Config().GetNetworkID(),
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
		header *p2ppb.CentrifugeHeader
		err    error
	}{
		{
			header: &p2ppb.CentrifugeHeader{CentNodeVersion: "someversion", NetworkIdentifier: 12},
			err:    documents.AppendError(version.IncompatibleVersionError("someversion"), incompatibleNetworkError(config.Config().GetNetworkID(), 12)),
		},

		{
			header: &p2ppb.CentrifugeHeader{CentNodeVersion: "0.0.1", NetworkIdentifier: 12},
			err:    documents.AppendError(incompatibleNetworkError(config.Config().GetNetworkID(), 12), nil),
		},

		{
			header: &p2ppb.CentrifugeHeader{CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config().GetNetworkID()},
		},
	}

	for _, c := range tests {
		err := handshakeValidator().Validate(c.header)
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
