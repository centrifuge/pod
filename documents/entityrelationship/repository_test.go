// +build unit

package entityrelationship

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestRepo_FindEntityRelationshipIdentifier(t *testing.T) {
	// setup repo
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	repo := testEntityRepo()
	assert.NotNil(t, repo)

	// no relationships in repo
	rp := testingdocuments.CreateRelationshipPayload()
	id, err := hexutil.Decode(rp.DocumentId)
	assert.NoError(t, err)

	tID, err := identity.StringsToDIDs(rp.TargetIdentity)
	assert.NoError(t, err)

	_, err = repo.FindEntityRelationshipIdentifier(id, did, *tID[0])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found in the system database")

	// create relationships
	m, err := service{}.DeriveFromCreatePayload(ctxh, rp)
	assert.NoError(t, err)

	err = repo.Create(did[:], m.ID(), m)
	assert.NoError(t, err)

	rp.TargetIdentity = testingidentity.GenerateRandomDID().String()
	m2, err := service{}.DeriveFromCreatePayload(ctxh, rp)
	assert.NoError(t, err)

	err = repo.Create(did[:], m2.ID(), m2)
	assert.NoError(t, err)

	// attempt to get relationships
	r, err := repo.FindEntityRelationshipIdentifier(id, did, *tID[0])
	assert.NoError(t, err)
	assert.Equal(t, r, m.CurrentVersion())

	// throws err if relationship not found in the repo
	r, err = repo.FindEntityRelationshipIdentifier(id, testingidentity.GenerateRandomDID(), *tID[0])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found in the system database")
}

func TestRepo_ListAllRelationships(t *testing.T) {
	// setup repo
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	repo := testEntityRepo()
	assert.NotNil(t, repo)

	rp := testingdocuments.CreateRelationshipPayload()
	id, err := hexutil.Decode(rp.DocumentId)
	assert.NoError(t, err)

	// no relationships in repo returns a nil map
	r, err := repo.ListAllRelationships(id, did)
	assert.Equal(t, r, map[string][]byte{})

	// create relationships
	m, err := service{}.DeriveFromCreatePayload(ctxh, rp)
	assert.NoError(t, err)

	err = repo.Create(did[:], m.ID(), m)
	assert.NoError(t, err)

	rp.TargetIdentity = testingidentity.GenerateRandomDID().String()
	m2, err := service{}.DeriveFromCreatePayload(ctxh, rp)
	assert.NoError(t, err)

	err = repo.Create(did[:], m2.ID(), m2)
	assert.NoError(t, err)

	rp.TargetIdentity = testingidentity.GenerateRandomDID().String()
	m3, err := service{}.DeriveFromCreatePayload(ctxh, rp)
	assert.NoError(t, err)

	err = repo.Create(did[:], m3.ID(), m3)
	assert.NoError(t, err)

	// attempt to get relationships
	id, err = hexutil.Decode(rp.DocumentId)
	assert.NoError(t, err)

	r, err = repo.ListAllRelationships(id, did)
	assert.Len(t, r, 3)
}
