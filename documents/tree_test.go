// +build unit

package documents

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToProofAndProtoSalts(t *testing.T) {
	cd, err := newCoreDocument()
	assert.Error(t, err)
	salts, err := GenerateNewSalts(&cd.Document, "", nil)
	assert.NoError(t, err)
	assert.NotNil(t, salts)

	nilProto := ConvertToProtoSalts(nil)
	assert.Nil(t, nilProto)

	nilProof := ConvertToProofSalts(nil)
	assert.Nil(t, nilProof)

	protoSalts := ConvertToProtoSalts(salts)
	assert.NotNil(t, protoSalts)
	assert.Len(t, protoSalts, len(*salts))
	assert.Equal(t, protoSalts[0].Value, (*salts)[0].Value)

	cSalts := ConvertToProofSalts(protoSalts)
	assert.NotNil(t, cSalts)
	assert.Len(t, *cSalts, len(*salts))
	assert.Equal(t, (*cSalts)[0].Value, (*salts)[0].Value)
}
