package model

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"reflect"
)

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
type Model interface {

	//Returns the underlying type of the Model
	Type() reflect.Type

	// Convert the model into a core document to be transported. It embeds the business object specific fields into the `EmbeddedData` field.
	CoreDocument() *coredocumentpb.CoreDocument

	// SetCoreDocument sets fields from given CoreDocument into the model
	SetCoreDocument(cd *coredocumentpb.CoreDocument) error

	// return the json representation of the model
	JSON() ([]byte, error)

}

