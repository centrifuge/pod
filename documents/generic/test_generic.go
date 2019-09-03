// +build integration unit testworld

package generic

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/stretchr/testify/assert"
)

func InitGeneric(t *testing.T, did identity.DID, payload documents.CreatePayload) *Generic {
	gen := new(Generic)
	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	assert.NoError(t, gen.DeriveFromCreatePayload(context.Background(), payload))
	return gen
}

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}
