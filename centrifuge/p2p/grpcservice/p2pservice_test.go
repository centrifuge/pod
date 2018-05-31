package grpcservice

import (
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var dbFileName = "/tmp/centrifuge_testing_inv_service.leveldb"

func TestMain(m *testing.M) {
	config.Config.InitializeViper()
	config.Config.V.Set("storage.Path", dbFileName)
	cc.TestBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestP2PService(t *testing.T) {
	identifier := []byte("1")
	coredoc := &coredocumentpb.CoreDocument{DocumentIdentifier: identifier}

	req := p2ppb.P2PMessage{Document: coredoc, CentNodeVersion: version.CENTRIFUGE_NODE_VERSION, NetworkIdentifier: config.Config.GetNetworkID()}
	rpc := P2PService{}
	res, err := rpc.HandleP2PPost(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.Equal(t, res.Document.DocumentIdentifier, identifier, "Incorrect identifier")

	doc, err := coredocumentrepository.GetCoreDocumentRepository().FindById(identifier)
	assert.Equal(t, doc.DocumentIdentifier, identifier, "Document Identifier doesn't match")

}
