package documents

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/precise-proofs/proofs/proto"
)

// DocumentProof is a value to represent a document and its field proofs
type DocumentProof struct {
	DocumentID  []byte
	VersionID   []byte
	State       string
	FieldProofs []*proofspb.Proof
}

// Service provides an interface for functions common to all document types
type Service interface {

	// GetCurrentVersion reads a document from the database
	GetCurrentVersion(ctx context.Context, documentID []byte) (Model, error)

	// Exists checks if a document exists
	Exists(ctx context.Context, documentID []byte) bool

	// GetVersion reads a document from the database
	GetVersion(ctx context.Context, documentID []byte, version []byte) (Model, error)

	// DeriveFromCoreDocument derives a model given the core document
	DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (Model, error)

	// CreateProofs creates proofs for the latest version document given the fields
	CreateProofs(ctx context.Context, documentID []byte, fields []string) (*DocumentProof, error)

	// CreateProofsForVersion creates proofs for a particular version of the document given the fields
	CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*DocumentProof, error)

	// RequestDocumentSignature Validates and Signs document received over the p2p layer
	RequestDocumentSignature(ctx context.Context, model Model) (*coredocumentpb.Signature, error)

	// ReceiveAnchoredDocument receives a new anchored document over the p2p layer, validates and updates the document in DB
	ReceiveAnchoredDocument(ctx context.Context, model Model, senderID []byte) error
}
