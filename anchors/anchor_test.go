// +build unit

package anchors

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}
var cfg Config

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[config.BootstrappedConfig].(Config)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestNewAnchorId(t *testing.T) {
	tests := []struct {
		name  string
		slice []byte
		err   string
	}{
		{
			"smallerSlice",
			utils.RandomSlice(AnchorIDLength - 1),
			"invalid length byte slice provided for anchorID",
		},
		{
			"largerSlice",
			utils.RandomSlice(AnchorIDLength + 1),
			"invalid length byte slice provided for anchorID",
		},
		{
			"nilSlice",
			nil,
			"invalid length byte slice provided for anchorID",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ToAnchorID(test.slice)
			assert.Equal(t, test.err, err.Error())
		})
	}
}

func TestNewDocRoot(t *testing.T) {
	tests := []struct {
		name  string
		slice []byte
		err   string
	}{
		{
			"smallerSlice",
			utils.RandomSlice(DocumentRootLength - 1),
			"invalid length byte slice provided for docRoot",
		},
		{
			"largerSlice",
			utils.RandomSlice(DocumentRootLength + 1),
			"invalid length byte slice provided for docRoot",
		},
		{
			"nilSlice",
			nil,
			"invalid length byte slice provided for docRoot",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ToDocumentRoot(test.slice)
			assert.Equal(t, test.err, err.Error())
		})
	}
}
