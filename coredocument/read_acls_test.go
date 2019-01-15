// +build unit

package coredocument

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/stretchr/testify/assert"
)

func TestReadACLs_initReadRules(t *testing.T) {
	cd := New()
	err := initReadRules(cd, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrZeroCollaborators, err))

	cs := []identity.CentID{identity.RandomCentID()}
	err = initReadRules(cd, cs)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Len(t, cd.Roles, 1)

	err = initReadRules(cd, cs)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Len(t, cd.Roles, 1)
}

func TestReadAccessValidator_PeerCanRead(t *testing.T) {
	pv := peerValidator()
	peer, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)

	cd, err := NewWithCollaborators([]string{peer.String()})
	assert.NoError(t, err)
	assert.NotNil(t, cd.ReadRules)
	assert.NotNil(t, cd.Roles)

	// peer who cant access
	rcid := identity.RandomCentID()
	err = pv.PeerCanRead(cd, rcid)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrPeerNotFound, err))

	// peer can access
	assert.NoError(t, pv.PeerCanRead(cd, peer))
}
