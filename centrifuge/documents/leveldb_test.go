// +build unit

package documents_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

var levelDB documents.Repository

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	levelDB = documents.LevelDBRepository{LevelDB: cc.GetLevelDBStorage()}
	flag.Parse()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

type model struct {
	shouldError bool
	Data        string `json:"data"`
}

func (m *model) GetDocumentID() ([]byte, error)                           { panic("implement me") }
func (m *model) Type() reflect.Type                                       { return reflect.TypeOf(m) }
func (m *model) PackCoreDocument() (*coredocumentpb.CoreDocument, error)  { return nil, nil }
func (m *model) UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error { return nil }

func (m *model) JSON() ([]byte, error) {
	if m.shouldError {
		return nil, fmt.Errorf("failed to marshal")
	}

	return json.Marshal(m)
}

func (m *model) FromJSON(data []byte) error {
	if m.shouldError {
		return fmt.Errorf("failed to unmarshal")
	}

	return json.Unmarshal(data, m)
}

func TestDefaultLevelDB_LoadByID(t *testing.T) {
	id := tools.RandomSlice(32)

	// missing ID
	err := levelDB.LoadByID(id, new(model))
	assert.Error(t, err, "error must be non nil")

	// nil document
	err = levelDB.LoadByID(id, nil)
	assert.Error(t, err, "error must be non nil")

	// Failed unmarshal
	m := &model{shouldError: true}
	err = levelDB.LoadByID(id, m)
	assert.Error(t, err, "error must be non nil")

	// success
	m = &model{Data: "hello, world"}
	err = levelDB.Create(id, m)
	assert.Nil(t, err, "error should be nil")
	nm := new(model)
	err = levelDB.LoadByID(id, nm)
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, m, nm, "models must match")
}

func TestDefaultLevelDB_Create(t *testing.T) {
	id := tools.RandomSlice(32)
	d := &model{Data: "Create it"}
	err := levelDB.Create(id, d)
	assert.Nil(t, err, "create must pass")

	// same id
	err = levelDB.Create(id, new(model))
	assert.Error(t, err, "create must fail")

	// nil model
	err = levelDB.Create(id, nil)
	assert.Error(t, err, "create must fail")
}

func TestDefaultLevelDB_UpdateModel(t *testing.T) {
	id := tools.RandomSlice(32)

	// missing Id
	err := levelDB.Update(id, new(model))
	assert.Error(t, err, "update must fail")

	// nil model
	err = levelDB.Update(id, nil)
	assert.Error(t, err, "update must fail")

	m := &model{Data: "create it"}
	err = levelDB.Create(id, m)
	assert.Nil(t, err, "create must pass")

	// successful one
	m.Data = "update it"
	err = levelDB.Update(id, m)
	assert.Nil(t, err, "update must pass")
	nm := new(model)
	err = levelDB.LoadByID(id, nm)
	assert.Nil(t, err, "get mode must pass")
	assert.Equal(t, m, nm)
}
