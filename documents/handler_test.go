// +build unit

package documents_test

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestGrpcHandler_CreateDocumentProof(t *testing.T) {
	registry := documents.NewServiceRegistry()
	serviceName := "CreateDocumentProof"
	service := &testingdocuments.MockService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofRequest{
		Identifier: "0xc32b1400b8c66e54448bec863233682d19c770b94ea8d90e1cf02f3bb8ca7da4",
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	id, _ := hexutil.Decode(req.Identifier)
	doc := &documents.DocumentProof{}
	service.On("CreateProofs", id, req.Fields).Return(doc, nil)
	grpcHandler := documents.GRPCHandler(registry)
	retDoc, _ := grpcHandler.CreateDocumentProof(context.TODO(), req)
	service.AssertExpectations(t)
	conv, _ := documents.ConvertDocProofToClientFormat(doc)
	assert.Equal(t, conv, retDoc)
}

func TestGrpcHandler_CreateDocumentProofUnableToLocateService(t *testing.T) {
	registry := documents.NewServiceRegistry()
	serviceName := "CreateDocumentProofUnableToLocateService"
	service := &testingdocuments.MockService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofRequest{
		Identifier: "0x11111111111111", // invalid
		Type:       "wrongService",
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler(registry)
	_, err := grpcHandler.CreateDocumentProof(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofs")
}

func TestGrpcHandler_CreateDocumentProofInvalidHex(t *testing.T) {
	registry := documents.NewServiceRegistry()
	serviceName := "CreateDocumentProofInvalidHex"
	service := &testingdocuments.MockService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofRequest{
		Identifier: "0x1111111111111", // invalid
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler(registry)
	_, err := grpcHandler.CreateDocumentProof(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofs")
}

func TestGrpcHandler_CreateDocumentProofForVersion(t *testing.T) {
	registry := documents.NewServiceRegistry()
	serviceName := "CreateDocumentProofForVersion"
	service := &testingdocuments.MockService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofForVersionRequest{
		Identifier: "0x11111111111111",
		Version:    "0x1212121212",
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	id, _ := hexutil.Decode(req.Identifier)
	version, _ := hexutil.Decode(req.Version)
	doc := &documents.DocumentProof{DocumentID: utils.RandomSlice(32)}
	service.On("CreateProofsForVersion", id, version, req.Fields).Return(doc, nil)
	grpcHandler := documents.GRPCHandler(registry)
	retDoc, _ := grpcHandler.CreateDocumentProofForVersion(context.TODO(), req)
	service.AssertExpectations(t)
	conv, _ := documents.ConvertDocProofToClientFormat(doc)
	assert.Equal(t, conv, retDoc)
}

func TestGrpcHandler_CreateDocumentProofForVersionUnableToLocateService(t *testing.T) {
	registry := documents.NewServiceRegistry()
	serviceName := "CreateDocumentProofForVersionUnableToLocateService"
	service := &testingdocuments.MockService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofForVersionRequest{
		Identifier: "0x11111111111111",
		Version:    "0x1212121212",
		Type:       "wrongService",
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler(registry)
	_, err := grpcHandler.CreateDocumentProofForVersion(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofsForVersion")
}

func TestGrpcHandler_CreateDocumentProofForVersionInvalidHexForId(t *testing.T) {
	registry := documents.NewServiceRegistry()
	serviceName := "CreateDocumentProofForVersionInvalidHexForId"
	service := &testingdocuments.MockService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofForVersionRequest{
		Identifier: "0x111111111111111",
		Version:    "0x1212121212",
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler(registry)
	_, err := grpcHandler.CreateDocumentProofForVersion(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofsForVersion")
}

func TestGrpcHandler_CreateDocumentProofForVersionInvalidHexForVersion(t *testing.T) {
	registry := documents.NewServiceRegistry()
	serviceName := "CreateDocumentProofForVersionInvalidHexForVersion"
	service := &testingdocuments.MockService{}
	registry.Register(serviceName, service)
	req := &documentpb.CreateDocumentProofForVersionRequest{
		Identifier: "0x11111111111111",
		Version:    "0x12121212121",
		Type:       serviceName,
		Fields:     []string{"field1"},
	}
	grpcHandler := documents.GRPCHandler(registry)
	_, err := grpcHandler.CreateDocumentProofForVersion(context.TODO(), req)
	assert.NotNil(t, err)
	service.AssertNotCalled(t, "CreateProofsForVersion")
}

func TestConvertDocProofToClientFormat(t *testing.T) {
	tests := []struct {
		name   string
		input  *documents.DocumentProof
		output documentpb.DocumentProof
	}{
		{
			name: "happy",
			input: &documents.DocumentProof{
				DocumentID: []byte{1, 2, 1},
				VersionID:  []byte{1, 2, 2},
				State:      "state",
				FieldProofs: []*proofspb.Proof{
					{
						Property: "prop1",
						Value:    "val1",
						Salt:     []byte{1, 2, 3},
						Hash:     []byte{1, 2, 4},
						SortedHashes: [][]byte{
							{1, 2, 5},
							{1, 2, 6},
							{1, 2, 7},
						},
					},
				},
			},
			output: documentpb.DocumentProof{
				Header: &documentpb.ResponseHeader{
					DocumentId: "0x010201",
					VersionId:  "0x010202",
					State:      "state",
				},
				FieldProofs: []*documentpb.Proof{
					{
						Property: "prop1",
						Value:    "val1",
						Salt:     "0x010203",
						Hash:     "0x010204",
						SortedHashes: []string{
							"0x010205",
							"0x010206",
							"0x010207",
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := documents.ConvertDocProofToClientFormat(test.input)
			assert.Nil(t, err)
			assert.Equal(t, test.output.Header.DocumentId, out.Header.DocumentId)
			assert.Equal(t, test.output.Header.VersionId, out.Header.VersionId)
			assert.Equal(t, test.output.Header.State, out.Header.State)
			assert.Equal(t, len(test.output.FieldProofs), len(out.FieldProofs))
			for i, converted := range test.output.FieldProofs {
				assert.Equal(t, converted.Hash, out.FieldProofs[i].Hash)
				assert.Equal(t, converted.Salt, out.FieldProofs[i].Salt)
				assert.Equal(t, converted.Property, out.FieldProofs[i].Property)
				assert.Equal(t, converted.Value, out.FieldProofs[i].Value)
				for j, h := range converted.SortedHashes {
					assert.Equal(t, h, out.FieldProofs[i].SortedHashes[j])
				}
			}
		})
	}
}

func TestConvertProofsToClientFormat(t *testing.T) {
	clientFormat := documents.ConvertProofsToClientFormat([]*proofspb.Proof{
		{
			Property: "prop1",
			Value:    "val1",
			Salt:     utils.RandomSlice(32),
			Hash:     utils.RandomSlice(32),
			SortedHashes: [][]byte{
				utils.RandomSlice(32),
				utils.RandomSlice(32),
				utils.RandomSlice(32),
			},
		},
		{
			Property: "prop2",
			Value:    "val2",
			Salt:     utils.RandomSlice(32),
			Hash:     utils.RandomSlice(32),
			SortedHashes: [][]byte{
				utils.RandomSlice(32),
				utils.RandomSlice(32),
			},
		},
	})
	for _, converted := range clientFormat {
		verifyConverted(t, converted)
	}
}

func verifyConverted(t *testing.T, proof *documentpb.Proof) {
	verifyHex(t, proof.Salt)
	verifyHex(t, proof.Hash)
	for _, h := range proof.SortedHashes {
		verifyHex(t, h)
	}
}

func verifyHex(t *testing.T, val string) {
	_, err := hexutil.Decode(val)
	assert.Nil(t, err)
}
