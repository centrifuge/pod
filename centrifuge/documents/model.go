package documents

import (
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
)

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
type Model interface {
	Packer
	Unpacker

	//Returns the underlying type of the Model
	Type() reflect.Type

	// JSON return the json representation of the model
	JSON() ([]byte, error)

	// FromJSON initialize the model with a json
	FromJSON(json []byte) error
}

// Packer can be implemented by any type can pack itself as a core document
type Packer interface {
	// PackCoreDocument packs the implementing document into a core document
	// should create the identifiers for the core document if not present
	PackCoreDocument() (*coredocumentpb.CoreDocument, error)
}

// Unpacker can be implemented by any type that return a model from core document
type Unpacker interface {
	// UnpackCoreDocument must return the document.Model
	// assumes that core document has valid identifiers set
	UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error
}
