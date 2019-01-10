package receiver

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/libp2p/go-libp2p-protocol"
	"github.com/stretchr/testify/assert"
)

func TestExtractCID(t *testing.T) {
	p := protocol.ID("/centrifuge/0.0.1/0xd9f72e705074")
	cid, err := ExtractCID(p)
	assert.NoError(t, err)
	assert.Equal(t, "0xd9f72e705074", cid.String())
}

func TestProtocolForCID(t *testing.T) {
	cid := identity.RandomCentID()
	p := ProtocolForCID(cid)
	assert.Contains(t, p, cid.String())
	cidE, err := ExtractCID(p)
	assert.NoError(t, err)
	assert.Equal(t, cid.String(), cidE.String())
}
