//go:build integration || unit || testworld

package entityrelationship

import (
	"context"
	"encoding/json"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}

func InitEntityRelationship(t *testing.T, ctx context.Context, data Data) *EntityRelationship {
	er := new(EntityRelationship)
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	err = er.DeriveFromCreatePayload(ctx, documents.CreatePayload{Data: d})
	assert.NoError(t, err)
	return er
}

func CreateRelationship(t *testing.T, ctx context.Context) *EntityRelationship {
	did, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	target, _ := identity.StringsToDIDs("0x5F9132e0F92952abCb154A9b34563891ffe1AAcb")
	d := Data{
		EntityIdentifier: utils.RandomSlice(32),
		OwnerIdentity:    &did,
		TargetIdentity:   target[0],
	}
	return InitEntityRelationship(t, ctx, d)
}

func CreateCDWithEmbeddedEntityRelationship(t *testing.T, ctx context.Context) (documents.Document, coredocumentpb.CoreDocument) {
	e := CreateRelationship(t, ctx)
	_, err := e.CalculateSigningRoot()
	assert.NoError(t, err)
	_, err = e.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err := e.PackCoreDocument()
	assert.NoError(t, err)
	return e, cd
}
