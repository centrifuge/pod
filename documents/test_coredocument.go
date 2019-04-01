// +build integration unit

package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
)

// GetTestCoreDocWithReset must only be used by tests for manipulations. It gets the embedded coredoc protobuf.
// All calls to this function will cause a regeneration of salts next time for precise-proof trees.
func (cd *CoreDocument) GetTestCoreDocWithReset() *coredocumentpb.CoreDocument {
	cd.Modified = true
	return &cd.Document
}
