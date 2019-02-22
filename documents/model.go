package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
)

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
// It should only handle protocol-level Document actions
type Model interface {
	storage.Model

	// ID returns the document identifier
	ID() []byte

	// CurrentVersion returns the current version identifier of the document
	CurrentVersion() []byte

	// PreviousVersion returns the previous version identifier of the document
	PreviousVersion() []byte

	// NextVersion returns the next version identifier of the document.
	NextVersion() []byte

	// PackCoreDocument packs the implementing document into a core document
	// Should only be called when the document is about to be put on wire.
	PackCoreDocument() (coredocumentpb.CoreDocument, error)

	// UnpackCoreDocument takes a core document protobuf and loads the data into the model.
	UnpackCoreDocument(cd coredocumentpb.CoreDocument) error

	// DocumentType returns the type of the document
	DocumentType() string

	// DataRoot returns the data root of the model.
	DataRoot() ([]byte, error)

	// SigningRoot returns the signing root of the model.
	SigningRoot() ([]byte, error)

	// DocumentRoot returns the document root of the model.
	DocumentRoot() ([]byte, error)

	// PreviousDocumentRoot returns the document root of the previous version.
	PreviousDocumentRoot() []byte

	// AppendSignatures appends the signatures to the model.
	AppendSignatures(signatures ...*coredocumentpb.Signature)

	// Signatures returns a copy of the signatures on the document
	Signatures() []coredocumentpb.Signature

	// CreateProofs creates precise-proofs for given fields
	CreateProofs(fields []string) (proofs []*proofspb.Proof, err error)

	// CreateNFTProofs creates NFT proofs for minting.
	CreateNFTProofs(
		account identity.CentID,
		registry common.Address,
		tokenID []byte,
		nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error)

	// IsNFTMinted checks if there is any NFT minted for the registry given
	IsNFTMinted(tr TokenRegistry, registry common.Address) bool

	// AddNFT adds an NFT to the document.
	// Note: The document should be anchored after successfully adding the NFT.
	AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte) error

	// GetCollaborators returns the collaborators of this document.
	// filter ids should not be returned
	GetCollaborators(filterIDs ...identity.CentID) ([]identity.CentID, error)
}

// TokenRegistry defines NFT related functions.
type TokenRegistry interface {
	// OwnerOf to retrieve owner of the tokenID
	OwnerOf(registry common.Address, tokenID []byte) (common.Address, error)
}
