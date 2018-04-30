// +build unit

package p2p

import (
	"testing"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"os"
	"github.com/spf13/viper"
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

//func TestP2PService(t *testing.T) {
//
//	identifier := []byte("1")
//	coredoc := coredocument.NewCoreDocument(&coredocumentpb.CoreDocument{DocumentIdentifier: identifier})
//
//	req := P2PMessage{Document: coredoc.Document}
//	rpc := P2PService{}
//	res, err := rpc.Transmit(context.Background(), &req)
//	assert.Nil(t, err, "Received error")
//	assert.Equal(t, res.Document.DocumentIdentifier, identifier, "Incorrect identifier")
//
//	doc, err := repository.NewLevelDBCoreDocumentRepository(cc.LevelDB).FindById(identifier)
//	assert.Equal(t, doc.DocumentIdentifier, identifier, "Document Identifier doesn't match")
//
//}