// +build unit

package coredocumentprocessor

import (
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
)

var cdp defaultProcessor

func TestMain(m *testing.M) {
	cdp = defaultProcessor{}
	result := m.Run()
	os.Exit(result)
}

func TestCoreDocumentProcessor_SendNilDocument(t *testing.T) {
	err := cdp.Send(nil, nil, []byte{})
	assert.Error(t, err, "should have thrown an error")
}

func TestCoreDocumentProcessor_AnchorNilDocument(t *testing.T) {
	err := cdp.Anchor(nil)
	assert.Error(t, err, "should have thrown an error")
}

func TestCoreDocumentProcessor_getDocumentTree(t *testing.T) {
	cd := &coredocumentpb.CoreDocument{DocumentIdentifier: tools.RandomSlice32()}
	cds := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(cd, cds)
	cd.CoredocumentSalts = cds
	tree, err := cdp.getDocumentTree(cd)
	assert.Nil(t, err)
	assert.NotNil(t, tree)
}
