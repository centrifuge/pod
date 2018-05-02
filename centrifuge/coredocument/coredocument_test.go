// +build unit

package coredocument

import (
	"testing"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"os"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/stretchr/testify/assert"
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/grpc"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/documentservice"
)

var dbFileName = "/tmp/centrifuge_testing_p2p_server.leveldb"

func TestMain(m *testing.M) {
	viper.Set("storage.Path", dbFileName)
	cc.Bootstrap()
	defer cc.LevelDB.Close()

	result := m.Run()
	os.RemoveAll(dbFileName)
	os.Exit(result)
}

func TestP2PService(t *testing.T) {

	identifier := []byte("1")
	coredoc := NewCoreDocument(&coredocumentpb.CoreDocument{DocumentIdentifier: identifier})

	req := grpc.P2PMessage{Document: coredoc.Document}
	rpc := documentservice.P2PService{}
	res, err := rpc.Post(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.Equal(t, res.Document.DocumentIdentifier, identifier, "Incorrect identifier")

	doc, err := repository.NewLevelDBCoreDocumentRepository(cc.LevelDB).FindById(identifier)
	assert.Equal(t, doc.DocumentIdentifier, identifier, "Document Identifier doesn't match")

}
