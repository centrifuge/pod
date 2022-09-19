//go:build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"

	"github.com/centrifuge/precise-proofs/proofs"

	"github.com/stretchr/testify/assert"
)

func TestCoreDocument_DefaultTreeWithPrefix(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	prefix := "prefix"
	compactPrefix := []byte("compact-prefix")

	res, err := cd.DefaultTreeWithPrefix(prefix, compactPrefix)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	res, err = cd.DefaultTreeWithPrefix("", compactPrefix)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	res, err = cd.DefaultTreeWithPrefix("", nil)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestCoreDocument_DefaultOrderedTreeWithPrefix(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	prefix := "prefix"
	compactPrefix := []byte("compact-prefix")

	res, err := cd.DefaultOrderedTreeWithPrefix(prefix, compactPrefix)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	res, err = cd.DefaultOrderedTreeWithPrefix("", compactPrefix)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	res, err = cd.DefaultOrderedTreeWithPrefix("", nil)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestCoreDocument_DocumentSaltsFunc(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	saltCompact1 := utils.RandomSlice(32)
	saltValue1 := utils.RandomSlice(32)
	saltCompact2 := utils.RandomSlice(32)
	saltValue2 := utils.RandomSlice(32)

	salts := []*proofspb.Salt{
		{
			Compact: saltCompact1,
			Value:   saltValue1,
		},
		{
			Compact: saltCompact2,
			Value:   saltValue2,
		},
	}

	cd.Document.Salts = salts
	cd.Modified = true

	saltsFn := cd.DocumentSaltsFunc()

	res, err := saltsFn(saltCompact1)
	assert.NoError(t, err)
	assert.Equal(t, res, saltValue1)

	res, err = saltsFn(saltCompact2)
	assert.NoError(t, err)
	assert.Equal(t, res, saltValue2)

	res, err = saltsFn(utils.RandomSlice(32))
	assert.NoError(t, err)
	assert.NotEqual(t, res, saltValue1)
	assert.NotEqual(t, res, saltValue2)
}

func TestCoreDocument_DocumentSaltsFunc_DocumentNotModified(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	saltCompact1 := utils.RandomSlice(32)
	saltValue1 := utils.RandomSlice(32)
	saltCompact2 := utils.RandomSlice(32)
	saltValue2 := utils.RandomSlice(32)

	salts := []*proofspb.Salt{
		{
			Compact: saltCompact1,
			Value:   saltValue1,
		},
		{
			Compact: saltCompact2,
			Value:   saltValue2,
		},
	}

	cd.Document.Salts = salts
	cd.Modified = false

	saltsFn := cd.DocumentSaltsFunc()

	res, err := saltsFn(saltCompact1)
	assert.NoError(t, err)
	assert.Equal(t, res, saltValue1)

	res, err = saltsFn(saltCompact2)
	assert.NoError(t, err)
	assert.Equal(t, res, saltValue2)

	res, err = saltsFn(utils.RandomSlice(32))
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNewLeafProperty(t *testing.T) {
	prefix := "prefix"
	compactPrefix := []byte("compact-prefix")

	assert.Equal(
		t,
		proofs.NewProperty(prefix, compactPrefix...),
		NewLeafProperty(prefix, compactPrefix),
	)
}
