// +build unit

package p2p

import (
	"testing"
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"os"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var dbFileName = "/tmp/centrifuge_testing_p2p_server.leveldb"

func TestMain(m *testing.M) {
	viper.Set("storage.Path", dbFileName)
	mockBootstrap()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestP2PService(t *testing.T) {

	identifier := []byte("1")
	inv := invoice.NewInvoiceDocument()
	inv.CoreDocument = &coredocumentpb.CoreDocument{DocumentIdentifier: identifier}

	coredoc := invoice.ConvertToCoreDocument(inv)
	req := P2PMessage{Document: &coredoc}
	rpc := P2PService{}
	res, err := rpc.Transmit(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.Equal(t, res.Document.DocumentIdentifier, identifier, "Incorrect identifier")

	doc, err := cc.Node.GetCoreDocumentStorageService().GetDocument(identifier)
	unmarshalledInv := invoice.ConvertToInvoiceDocument(doc)
	assert.Equal(t, unmarshalledInv.Data.Amount, inv.Data.Amount,
		"Invoice Amount doesn't match")
	assert.Equal(t, doc.DocumentIdentifier, identifier,
		"Document Identifier doesn't match")

}

func mockBootstrap() {
	(&cc.MockCentNode{}).BootstrapDependencies()
}