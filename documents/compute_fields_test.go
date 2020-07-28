// +build unit

package documents

import (
	"io/ioutil"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func wasmLoader(t *testing.T, wasm string) []byte {
	if wasm == "" {
		return utils.RandomSlice(32)
	}

	d, err := ioutil.ReadFile(wasm)
	assert.NoError(t, err)
	return d
}

func Test_validateWASM(t *testing.T) {
	tests := []struct {
		wasm string
		err  error
	}{
		{
			err: errors.AppendError(nil, ErrComputeFieldsInvalidWASM),
		},

		{
			wasm: "../testingutils/compute_fields/without_allocate.wasm",
			err:  errors.AppendError(nil, ErrComputeFieldsAllocateNotFound),
		},

		{
			wasm: "../testingutils/compute_fields/without_compute.wasm",
			err:  errors.AppendError(nil, ErrComputeFieldsComputeNotFound),
		},

		{
			wasm: "../testingutils/compute_fields/simple_average.wasm",
		},
	}

	for _, test := range tests {
		wasm := wasmLoader(t, test.wasm)
		_, _, _, err := fetchComputeFunctions(wasm)
		assert.Equal(t, err, test.err)
	}
}

func getInvalidComputeFieldAttrs(t *testing.T) []Attribute {
	dec, err := NewDecimal("1000")
	assert.NoError(t, err)
	attr1, err := NewMonetaryAttribute("test", dec, nil, "USD")
	assert.NoError(t, err)
	return []Attribute{attr1}
}

func getValidComputeFieldAttrs(t *testing.T) []Attribute {
	attr1, err := NewStringAttribute("test", AttrInt256, "1000")
	assert.NoError(t, err)
	attr2, err := NewStringAttribute("test2", AttrInt256, "2000")
	assert.NoError(t, err)
	attrKey, err := AttrKeyFromLabel("test3")
	assert.NoError(t, err)
	i, err := NewInt256("3000")
	assert.NoError(t, err)
	ib := i.Bytes()
	attr3 := Attribute{
		KeyLabel: "test3",
		Key:      attrKey,
		Value: AttrVal{
			Type: AttrSigned,
			Signed: Signed{
				Identity:        testingidentity.GenerateRandomDID(),
				Type:            AttrInt256,
				DocumentVersion: utils.RandomSlice(32),
				Value:           ib[:],
				Signature:       utils.RandomSlice(32),
				PublicKey:       utils.RandomSlice(32),
			},
		},
	}
	return []Attribute{attr1, attr2, attr3}
}

func Test_toComputeFieldsAttribute(t *testing.T) {
	// invalid attributes
	attrs := getInvalidComputeFieldAttrs(t)
	cattrs, err := toComputeFieldsAttributes(attrs)
	assert.EqualError(t, err, "'monetary' attribute type not supported by compute fields")
	assert.Nil(t, cattrs)

	// valid attributes
	attrs = getValidComputeFieldAttrs(t)
	cattrs, err = toComputeFieldsAttributes(attrs)
	assert.NoError(t, err)
	assert.Len(t, cattrs, len(attrs))
}

func Test_executeWASM(t *testing.T) {
	tests := []struct {
		wasm   string
		attrs  []Attribute
		result [32]byte
	}{
		// invalid WASM
		{
			wasm: "../testingutils/compute_fields/without_allocate.wasm",
		},

		// invalid Attributes
		{
			wasm:  "../testingutils/compute_fields/simple_average.wasm",
			attrs: getInvalidComputeFieldAttrs(t),
		},

		// success
		{
			wasm:  "../testingutils/compute_fields/simple_average.wasm",
			attrs: getValidComputeFieldAttrs(t),
			// result = risk(1) + value((1000+2000+3000)/3) = 2000
			result: [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x7, 0xd0},
		},
	}

	for _, test := range tests {
		wasm := wasmLoader(t, test.wasm)
		result := executeWASM(wasm, test.attrs)
		assert.Equal(t, test.result, result)
	}
}
