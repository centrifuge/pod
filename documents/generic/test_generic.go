// +build integration unit testworld

package generic

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/stretchr/testify/assert"
)

func InitGeneric(t *testing.T, did identity.DID, payload documents.CreatePayload) *Generic {
	gen := new(Generic)
	assert.NoError(t, gen.unpackFromCreatePayload(did, payload))
	return gen
}
