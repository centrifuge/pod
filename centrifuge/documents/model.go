package documents

import (
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
)

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
type Model interface {

	// Get the ID of the document represented by this model
	ID() ([]byte, error)

	// PackCoreDocument packs the implementing document into a core document
	// should create the identifiers for the core document if not present
	PackCoreDocument() (*coredocumentpb.CoreDocument, error)

	// UnpackCoreDocument must return the document.Model
	// assumes that core document has valid identifiers set
	UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error

	//Returns the underlying type of the Model
	Type() reflect.Type

	// JSON return the json representation of the model
	JSON() ([]byte, error)

	// FromJSON initialize the model with a json
	FromJSON(json []byte) error
}
