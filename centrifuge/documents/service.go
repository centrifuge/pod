package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/common"
)

// Service provides an interface for functions common to all document types
type Service interface {

	// GetLastVersion reads a document from the database
	GetLastVersion(documentID []byte) (Model, error)

	// GetVersion reads a document from the database
	GetVersion(documentID []byte, version []byte) (Model, error)

	// DeriveFromCoreDocument derives a model given the core document
	DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (Model, error)

	// CreateProofs creates proofs for the latest version document given the fields
	CreateProofs(documentID []byte, fields []string) (common.DocumentProof, error)

	// CreateProofsForVersion creates proofs for a particular version of the document given the fields
	CreateProofsForVersion(documentID, version []byte, fields []string) (common.DocumentProof, error)

	// RequestDocumentSignature Validates and Signs document received over the p2p layer
	RequestDocumentSignature(model Model) (*coredocumentpb.Signature, error)

	// ReceiveAnchoredDocument receives a new anchored document over the p2p layer, validates and updates the document in DB
	ReceiveAnchoredDocument(model Model, headers *p2ppb.CentrifugeHeader) error
}
