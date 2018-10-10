package documents

import "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"

// ModelDeriver can be implemented by any type that can unpack the core document into a Model
type ModelDeriver interface {
	DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (Model, error)
}
