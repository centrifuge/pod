package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
)

type Service interface {

	DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (Model, error)

	CreateProofs(documentID string, fields []string) (*documentpb.DocumentProof, error)
}
