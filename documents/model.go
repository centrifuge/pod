package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs/proto"
)

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
type Model interface {
	storage.Model
	CoreDocumentInterface

	// Get the ID of the document represented by this model
	ID() ([]byte, error)

	// PackCoreDocument packs the implementing document into a core document
	// should create the identifiers for the core document if not present
	PackCoreDocument() (*coredocumentpb.CoreDocument, error)

	// UnpackCoreDocument must return the document.Model
	// assumes that core document has valid identifiers set
	UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error

	// CreateProofs creates precise-proofs for given fields
	CreateProofs(fields []string) (coreDoc *coredocumentpb.CoreDocument, proofs []*proofspb.Proof, err error)
}

type CoreDocumentInterface interface {

	New() *CoreDocModel


}

type CoreDocModel struct {
	CD *coredocumentpb.CoreDocument
}

// New returns a new core document
// Note: collaborators and salts are to be filled by the caller
func (m *CoreDocModel) New() *CoreDocModel {
	id := utils.RandomSlice(32)
	cd := &coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(32),
	}
	return &CoreDocModel{
		cd,
	}
}

type Packer interface {
	Pack()
	Unpack()
}

