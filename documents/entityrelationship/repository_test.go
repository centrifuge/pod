// +build unit

package entityrelationship

import (
	"testing"

	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestRepo_FindEntityRelationshipIdentifier(t *testing.T) {
	// setup repo
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	repo := testEntityRepo()
	assert.NotNil(t, repo)

	// no relationships in repo
	er := CreateRelationship(t, ctxh)
	_, err := repo.FindEntityRelationshipIdentifier(er.Data.EntityIdentifier, did, *er.Data.TargetIdentity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found in the system database")

	err = repo.Create(did[:], er.ID(), er)
	assert.NoError(t, err)

	tid := testingidentity.GenerateRandomDID()
	m2 := CreateRelationship(t, ctxh)
	m2.Data.TargetIdentity = &tid
	err = repo.Create(did[:], m2.ID(), m2)
	assert.NoError(t, err)

	// attempt to get relationships
	r, err := repo.FindEntityRelationshipIdentifier(m2.Data.EntityIdentifier, did, tid)
	assert.NoError(t, err)
	assert.Equal(t, r, m2.CurrentVersion())

	// throws err if relationship not found in the repo
	r, err = repo.FindEntityRelationshipIdentifier(er.ID(), testingidentity.GenerateRandomDID(), tid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found in the system database")
}

func TestRepo_ListAllRelationships(t *testing.T) {
	// setup repo
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	repo := testEntityRepo()
	assert.NotNil(t, repo)

	// no relationships in repo returns a nil map
	id := utils.RandomSlice(32)
	r, err := repo.ListAllRelationships(id, did)
	assert.Equal(t, r, map[string][]byte{})

	// create relationships
	m := CreateRelationship(t, ctxh)
	m.Data.EntityIdentifier = id
	err = repo.Create(did[:], m.ID(), m)
	assert.NoError(t, err)

	r, err = repo.ListAllRelationships(id, did)
	assert.Len(t, r, 1)
}
