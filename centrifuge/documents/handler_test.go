package documents_test

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
)

func TestGrpcHandler_CreateDocumentProof(t *testing.T) {
	registry := documents.GetRegistryInstance()
	serviceName := "CreateDocumentProof"
	service := &testingcommons.MockDocService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofRequest{
		Identifier: "0xc32b1400b8c66e54448bec863233682d19c770b94ea8d90e1cf02f3bb8ca7da4",
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	id, _ := hexutil.Decode(req.Identifier)
	doc := &documentpb.DocumentProof{}
	service.On("CreateProofs", id, req.Fields).Return(doc, errors.New("dummy"))
	grpcHandler := documents.GRPCHandler()
	retDoc, _ := grpcHandler.CreateDocumentProof(context.TODO(), req)
	service.AssertExpectations(t)
	assert.Equal(t, doc, retDoc)
}

func TestGrpcHandler_CreateDocumentProofUnableToLocateService(t *testing.T) {
	registry := documents.GetRegistryInstance()
	serviceName := "CreateDocumentProofUnableToLocateService"
	service := &testingcommons.MockDocService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofRequest{
		Identifier: "0x11111111111111", // invalid
		Type:       "wrongService",
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler()
	_, err := grpcHandler.CreateDocumentProof(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofs")
}

func TestGrpcHandler_CreateDocumentProofInvalidHex(t *testing.T) {
	registry := documents.GetRegistryInstance()
	serviceName := "CreateDocumentProofInvalidHex"
	service := &testingcommons.MockDocService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofRequest{
		Identifier: "0x1111111111111", // invalid
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler()
	_, err := grpcHandler.CreateDocumentProof(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofs")
}

func TestGrpcHandler_CreateDocumentProofForVersion(t *testing.T) {
	registry := documents.GetRegistryInstance()
	serviceName := "CreateDocumentProofForVersion"
	service := &testingcommons.MockDocService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofForVersionRequest{
		Identifier: "0x11111111111111",
		Version:    "0x1212121212",
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	id, _ := hexutil.Decode(req.Identifier)
	version, _ := hexutil.Decode(req.Version)
	doc := &documentpb.DocumentProof{}
	service.On("CreateProofsForVersion", id, version, req.Fields).Return(doc, errors.New("dummy"))
	grpcHandler := documents.GRPCHandler()
	retDoc, _ := grpcHandler.CreateDocumentProofForVersion(context.TODO(), req)
	service.AssertExpectations(t)
	assert.Equal(t, doc, retDoc)
}

func TestGrpcHandler_CreateDocumentProofForVersionUnableToLocateService(t *testing.T) {
	registry := documents.GetRegistryInstance()
	serviceName := "CreateDocumentProofForVersionUnableToLocateService"
	service := &testingcommons.MockDocService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofForVersionRequest{
		Identifier: "0x11111111111111",
		Version:    "0x1212121212",
		Type:       "wrongService",
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler()
	_, err := grpcHandler.CreateDocumentProofForVersion(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofsForVersion")
}

func TestGrpcHandler_CreateDocumentProofForVersionInvalidHexForId(t *testing.T) {
	registry := documents.GetRegistryInstance()
	serviceName := "CreateDocumentProofForVersionInvalidHexForId"
	service := &testingcommons.MockDocService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofForVersionRequest{
		Identifier: "0x111111111111111",
		Version:    "0x1212121212",
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler()
	_, err := grpcHandler.CreateDocumentProofForVersion(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofsForVersion")
}

func TestGrpcHandler_CreateDocumentProofForVersionInvalidHexForVersion(t *testing.T) {
	registry := documents.GetRegistryInstance()
	serviceName := "CreateDocumentProofForVersionInvalidHexForVersion"
	service := &testingcommons.MockDocService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofForVersionRequest{
		Identifier: "0x11111111111111",
		Version:    "0x12121212121",
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler()
	_, err := grpcHandler.CreateDocumentProofForVersion(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofsForVersion")
}
