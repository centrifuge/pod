package documents

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("document-api")

// grpcHandler handles all the invoice document related actions
// anchoring, sending, proof generation, finding stored invoice document
type grpcHandler struct {
}

// GRPCHandler returns an implementation of invoice.DocumentServiceServer
func GRPCHandler() documentpb.DocumentServiceServer {
	return &grpcHandler{}
}

// CreateDocumentProof creates precise proofs for the given fields
func (grpcHandler) CreateDocumentProof(ctx context.Context, createDocumentProofEnvelope *documentpb.CreateDocumentProofEnvelope) (*documentpb.DocumentProof, error) {
	apiLog.Error(createDocumentProofEnvelope)
	return nil, nil
}
