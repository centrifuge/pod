package documents

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	service, err := GetRegistryInstance().LocateService(createDocumentProofEnvelope.Type)
	if err != nil {
		return &documentpb.DocumentProof{}, err
	}
	identifier, err := hexutil.Decode(createDocumentProofEnvelope.Identifier)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}
	return service.CreateProofs(identifier, createDocumentProofEnvelope.Fields)
}

// CreateDocumentProofForVersion creates precise proofs for the given fields for the given version of the document
func (grpcHandler) CreateDocumentProofForVersion(ctx context.Context, createDocumentProofForVersionEnvelope *documentpb.CreateDocumentProofForVersionRequest) (*documentpb.DocumentProof, error) {
	service, err := GetRegistryInstance().LocateService(createDocumentProofForVersionEnvelope.Type)
	if err != nil {
		return &documentpb.DocumentProof{}, err
	}
	identifier, err := hexutil.Decode(createDocumentProofForVersionEnvelope.Identifier)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}
	version, err := hexutil.Decode(createDocumentProofForVersionEnvelope.Version)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}
	return service.CreateProofsForVersion(identifier, version, createDocumentProofForVersionEnvelope.Fields)
}
