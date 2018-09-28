package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
)

type ModelDeriver interface {
	DeriveWithCoreDocument(cd *coredocumentpb.CoreDocument) (Model, error)
}
