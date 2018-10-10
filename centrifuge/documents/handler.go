package documents

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("document-api")

// grpcHandler handles all the common document related actions: proof generation
type grpcHandler struct {
}

// GRPCHandler returns an implementation of documentpb.DocumentServiceServer
func GRPCHandler() documentpb.DocumentServiceServer {
	return grpcHandler{}
}

// CreateDocumentProof creates precise proofs for the given fields
func (grpcHandler) CreateDocumentProof(ctx context.Context, createDocumentProofEnvelope *documentpb.CreateDocumentProofRequest) (*documentpb.DocumentProof, error) {
	apiLog.Error("implement me")
	return nil, centerrors.New(code.Unknown, "implement me")
}

// CreateDocumentProof creates precise proofs for the given fields for the given version of the document
func (grpcHandler) CreateDocumentProofForVersion(context.Context, *documentpb.CreateDocumentProofForVersionRequest) (*documentpb.DocumentProof, error) {
	apiLog.Error("implement me")
	return nil, centerrors.New(code.Unknown, "implement me")
}
