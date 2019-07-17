// +build integration unit testworld

package generic

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/stretchr/testify/assert"
	"testing"
)

func InitGeneric(t *testing.T, did identity.DID, payload documents.CreatePayload) *Generic{
	gen := new(Generic)
	assert.NoError(t, gen.unpackFromCreatePayload(did, payload))
	return gen
}