// +build unit

package documents

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

type model struct {
	shouldError bool
	Data        string `json:"data"`
}

func (m *model) ID() ([]byte, error)                                      { return []byte{}, nil }
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
	id := utils.RandomSlice(32)

	// missing ID
	err := testLevelDB.LoadByID(id, new(model))
	assert.Error(t, err, "error must be non nil")

	// nil document
	err = testLevelDB.LoadByID(id, nil)
	assert.Error(t, err, "error must be non nil")

	// Failed unmarshal
	m := &model{shouldError: true}
	err = testLevelDB.LoadByID(id, m)
	assert.Error(t, err, "error must be non nil")

	// success
	m = &model{Data: "hello, world"}
	err = testLevelDB.Create(id, m)
	assert.Nil(t, err, "error should be nil")
	nm := new(model)
	err = testLevelDB.LoadByID(id, nm)
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, m, nm, "models must match")
}

func TestDefaultLevelDB_Create(t *testing.T) {
	id := utils.RandomSlice(32)
	d := &model{Data: "Create it"}
	err := testLevelDB.Create(id, d)
	assert.Nil(t, err, "create must pass")

	// same id
	err = testLevelDB.Create(id, new(model))
	assert.Error(t, err, "create must fail")

	// nil model
	err = testLevelDB.Create(id, nil)
	assert.Error(t, err, "create must fail")
}

func TestDefaultLevelDB_UpdateModel(t *testing.T) {
	id := utils.RandomSlice(32)

	// missing Id
	err := testLevelDB.Update(id, new(model))
	assert.Error(t, err, "update must fail")

	// nil model
	err = testLevelDB.Update(id, nil)
	assert.Error(t, err, "update must fail")

	m := &model{Data: "create it"}
	err = testLevelDB.Create(id, m)
	assert.Nil(t, err, "create must pass")

	// successful one
	m.Data = "update it"
	err = testLevelDB.Update(id, m)
	assert.Nil(t, err, "update must pass")
	nm := new(model)
	err = testLevelDB.LoadByID(id, nm)
	assert.Nil(t, err, "get mode must pass")
	assert.Equal(t, m, nm)
}
