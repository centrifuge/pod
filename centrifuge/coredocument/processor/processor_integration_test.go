// +build ethereum

package coredocumentprocessor

import (
	"context"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
)

func TestDefaultProcessor_Anchor(t *testing.T) {
	ctx := context.Background()
	p2pClient := &testingcommons.MockP2PWrapperClient{}
	dp := DefaultProcessor(identity.NewEthereumIdentityService(), p2pClient)
	doc := createDummyCD()
	collaborators := []identity.CentID{
		identity.NewRandomCentID(),
		identity.NewRandomCentID(),
	}
	p2pClient.On("GetSignaturesForDocument", ctx, doc, collaborators)
	err := dp.Anchor(ctx, doc, collaborators)
	assert.Nil(t, err, "Document should be anchored correctly")
	p2pClient.AssertExpectations(t)
}

func createDummyCD() *coredocumentpb.CoreDocument {
	cd := coredocumentpb.CoreDocument{DocumentIdentifier: tools.RandomSlice(32)}
	cd, _ = coredocument.FillIdentifiers(cd)
	randomRoot := anchors.NewRandomDocRoot()
	cd.DataRoot = randomRoot[:]
	cds := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(&cd, cds)
	cd.CoredocumentSalts = cds
	return &cd
}
