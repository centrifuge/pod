// +build unit

package coredocumentprocessor

import (
	"crypto/sha256"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
)

// TODO(ved): more tests required for processor
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
	cd := coredocumentpb.CoreDocument{DocumentIdentifier: tools.RandomSlice(32)}
	cd, _ = coredocument.FillIdentifiers(cd)
	cds := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(&cd, cds)
	cd.CoredocumentSalts = cds
	tree, err := cdp.getDocumentSigningTree(&cd)
	assert.Nil(t, err)
	assert.NotNil(t, tree)
}

func TestCoreDocumentProcessor_GetDataProofHashes(t *testing.T) {
	cd := coredocumentpb.CoreDocument{
		DataRoot: tools.RandomSlice(32),
	}
	cd , err := coredocument.FillIdentifiers(cd)
	assert.Nil(t, err)
	cds := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(&cd, cds)

	cd.CoredocumentSalts = cds

	err = cdp.calculateSigningRoot(&cd)
	assert.Nil(t, err)

	hashes, err := cdp.GetDataProofHashes(&cd)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(hashes))

	valid, err := proofs.ValidateProofSortedHashes(cd.DataRoot, hashes, cd.SigningRoot, sha256.New())
	assert.True(t, valid)
	assert.Nil(t, err)
}
