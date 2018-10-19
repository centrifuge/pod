package documents

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs/proto"
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
	apiLog.Infof("Document proof request %v", createDocumentProofEnvelope)

	service, err := GetRegistryInstance().LocateService(createDocumentProofEnvelope.Type)
	if err != nil {
		return &documentpb.DocumentProof{}, err
	}

	identifier, err := hexutil.Decode(createDocumentProofEnvelope.Identifier)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}

	proof, err := service.CreateProofs(identifier, createDocumentProofEnvelope.Fields)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}
	return ConvertDocProofToClientFormat(proof)
}

// CreateDocumentProofForVersion creates precise proofs for the given fields for the given version of the document
func (grpcHandler) CreateDocumentProofForVersion(ctx context.Context, createDocumentProofForVersionEnvelope *documentpb.CreateDocumentProofForVersionRequest) (*documentpb.DocumentProof, error) {
	apiLog.Infof("Document proof request %v", createDocumentProofForVersionEnvelope)

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

	proof, err := service.CreateProofsForVersion(identifier, version, createDocumentProofForVersionEnvelope.Fields)
	if err != nil {
		return &documentpb.DocumentProof{}, centerrors.New(code.Unknown, err.Error())
	}
	return ConvertDocProofToClientFormat(proof)
}

// ConvertDocProofToClientFormat converts a DocumentProof to client api format
func ConvertDocProofToClientFormat(proof *DocumentProof) (*documentpb.DocumentProof, error) {
	return &documentpb.DocumentProof{
		Header: &documentpb.ResponseHeader{
			DocumentId: hexutil.Encode(proof.DocumentId),
			VersionId:  hexutil.Encode(proof.VersionId),
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
		Property:     proof.Property,
		Value:        proof.Value,
		Salt:         hexutil.Encode(proof.Salt),
		Hash:         hexutil.Encode(proof.Hash),
		SortedHashes: tools.SliceOfByteSlicesToHexStringSlice(proof.SortedHashes),
	}
}
