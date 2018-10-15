// +build integration

package coredocumentprocessor_test

import (
	"context"
	"testing"

	"os"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
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
	dp := coredocumentprocessor.DefaultProcessor(identity.IDService, p2pClient)
	doc := createDummyCD()

	p2pClient.On("GetSignaturesForDocument", ctx, doc).Return(nil)
	err := dp.Anchor(ctx, doc, nil)
	assert.Nil(t, err, "Document should be anchored correctly")
	p2pClient.AssertExpectations(t)
}

func createDummyCD() *coredocumentpb.CoreDocument {
	cd := coredocument.New()
	randomRoot := anchors.NewRandomDocRoot()
	cd.DataRoot = randomRoot[:]
	cd.Collaborators = [][]byte{
		tools.RandomSlice(identity.CentIDLength),
		tools.RandomSlice(identity.CentIDLength),
	}
	docAny := &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	cd.EmbeddedData = docAny
	coredocument.FillSalts(cd)
	return cd
}
