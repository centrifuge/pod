package documents

import (
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("document-api")

// grpcHandler handles all the common document related actions: proof generation
type grpcHandler struct {
	config   config.Service
	registry *ServiceRegistry
}

// GRPCHandler returns an implementation of documentpb.DocumentServiceServer
func GRPCHandler(config config.Service, registry *ServiceRegistry) documentpb.DocumentServiceServer {
	return grpcHandler{config: config, registry: registry}
}

// CreateDocumentProof creates precise proofs for the given fields
func (h grpcHandler) CreateDocumentProof(ctx context.Context, createDocumentProofEnvelope *documentpb.CreateDocumentProofRequest) (*documentpb.DocumentProof, error) {
	apiLog.Infof("Document proof request %v", createDocumentProofEnvelope)
	cctx, err := contextutil.Context(ctx, h.config)
	if err != nil {
		return &documentpb.DocumentProof{}, err
	}

	service, err := h.registry.LocateService(createDocumentProofEnvelope.Type)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.Wrap(err, "could not locate service for document type")
	}

	identifier, err := hexutil.Decode(createDocumentProofEnvelope.Identifier)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}

	proof, err := service.CreateProofs(cctx, identifier, createDocumentProofEnvelope.Fields)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}
	return ConvertDocProofToClientFormat(proof)
}

// CreateDocumentProofForVersion creates precise proofs for the given fields for the given version of the document
func (h grpcHandler) CreateDocumentProofForVersion(ctx context.Context, createDocumentProofForVersionEnvelope *documentpb.CreateDocumentProofForVersionRequest) (*documentpb.DocumentProof, error) {
	apiLog.Infof("Document proof request %v", createDocumentProofForVersionEnvelope)
	cctx, err := contextutil.Context(ctx, h.config)
	if err != nil {
		return &documentpb.DocumentProof{}, err
	}

	service, err := h.registry.LocateService(createDocumentProofForVersionEnvelope.Type)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.Wrap(err, "could not locate service for document type")
	}

	identifier, err := hexutil.Decode(createDocumentProofForVersionEnvelope.Identifier)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}

	version, err := hexutil.Decode(createDocumentProofForVersionEnvelope.Version)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}

	proof, err := service.CreateProofsForVersion(cctx, identifier, version, createDocumentProofForVersionEnvelope.Fields)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}
	return ConvertDocProofToClientFormat(proof)
}

// ConvertDocProofToClientFormat converts a DocumentProof to client api format
func ConvertDocProofToClientFormat(proof *DocumentProof) (*documentpb.DocumentProof, error) {
	return &documentpb.DocumentProof{
		Header: &documentpb.ResponseHeader{
			DocumentId: hexutil.Encode(proof.DocumentID),
			VersionId:  hexutil.Encode(proof.VersionID),
			State:      proof.State,
		},
		FieldProofs: ConvertProofsToClientFormat(proof.FieldProofs)}, nil
}

// ConvertProofsToClientFormat converts a proof protobuf from precise proofs into a client protobuf proof format
func ConvertProofsToClientFormat(proofs []*proofspb.Proof) []*documentpb.Proof {
	converted := make([]*documentpb.Proof, len(proofs))
	for i, proof := range proofs {
		converted[i] = ConvertProofToClientFormat(proof)
	}
	return converted
}

// ConvertProofToClientFormat converts a proof in precise proof format in to a client protobuf proof
func ConvertProofToClientFormat(proof *proofspb.Proof) *documentpb.Proof {
	return &documentpb.Proof{
		Property:     proof.GetReadableName(),
		Value:        proof.Value,
		Salt:         hexutil.Encode(proof.Salt),
		Hash:         hexutil.Encode(proof.Hash),
		SortedHashes: utils.SliceOfByteSlicesToHexStringSlice(proof.SortedHashes),
	}
}
