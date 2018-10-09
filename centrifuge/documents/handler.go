package documents

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("document-api")

// grpcHandler handles all the invoice document related actions: proof generation
type grpcHandler struct {
}

// GRPCHandler returns an implementation of documentpb.DocumentServiceServer
func GRPCHandler() documentpb.DocumentServiceServer {
	return grpcHandler{}
}

// CreateDocumentProof creates precise proofs for the given fields
func (grpcHandler) CreateDocumentProof(ctx context.Context, createDocumentProofEnvelope *documentpb.CreateDocumentProofRequest) (*documentpb.DocumentProof, error) {
	apiLog.Error(createDocumentProofEnvelope)
	return nil, nil
}

func (grpcHandler) CreateDocumentProofForVersion(context.Context, *documentpb.CreateDocumentProofForVersionRequest) (*documentpb.DocumentProof, error) {
	panic("implement me")
}
