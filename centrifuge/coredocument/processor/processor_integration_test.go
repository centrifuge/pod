// +build integration

package coredocumentprocessor

import (
	"context"
	"testing"

	"os"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	db := cc.GetLevelDBStorage()
	coredocumentrepository.InitLevelDBRepository(db)
	testingutils.CreateIdentityWithKeys()
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestDefaultProcessor_Anchor(t *testing.T) {
	ctx := context.Background()
	p2pClient := &testingcommons.MockP2PWrapperClient{}
	dp := DefaultProcessor(identity.NewEthereumIdentityService(), p2pClient)
	doc := createDummyCD()
	collaborators := []identity.CentID{
		identity.NewRandomCentID(),
		identity.NewRandomCentID(),
	}
	p2pClient.On("GetSignaturesForDocument", ctx, doc, collaborators).Return(nil)
	err := dp.Anchor(ctx, doc, collaborators, nil)
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
