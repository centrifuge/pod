package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
)

// ModelDeriver can be implemented by any document that can represent itself as a CoreDocument
type ModelDeriver interface {
	// DeriveFromCoreDocument must return the document.Model
	// assumes that core document has valid identifiers set
	DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (Model, error)
}
