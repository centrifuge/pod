// +build unit

package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestP2PDiscoveryPurposeHash(t *testing.T) {
	h := sha256.New()
	_, err := h.Write([]byte("CENTRIFUGE@P2P_DISCOVERY"))
	assert.NoError(t, err)
	hb := h.Sum(nil)
	assert.Len(t, hb, 32)
	p2pHex := hex.EncodeToString(hb)
	assert.Equal(t, KeyPurposeP2PDiscovery.HexValue, p2pHex)
}

func TestSigningPurposeHash(t *testing.T) {
	h := sha256.New()
	_, err := h.Write([]byte("CENTRIFUGE@SIGNING"))
	assert.NoError(t, err)
	hb := h.Sum(nil)
	assert.Len(t, hb, 32)
	p2pHex := hex.EncodeToString(hb)
	assert.Equal(t, KeyPurposeSigning.HexValue, p2pHex)
}

func TestKeyPurposeP2PDiscovery(t *testing.T) {
	purpose := KeyPurposeP2PDiscovery
	assert.Equal(t, "P2P_DISCOVERY", purpose.Name)
	assert.Equal(t, "88dbd1f0b244e515ab5aee93b5dee6a2d8e326576a583822635a27e52e5b591e", purpose.HexValue)
	assert.Equal(t, "61902935868658303950246481358903666251099779839440421743915568792957869578526", purpose.Value.String())
}

func TestKeyPurposeSigning(t *testing.T) {
	purpose := KeyPurposeSigning
	assert.Equal(t, "SIGNING", purpose.Name)
	assert.Equal(t, "774a43710604e3ce8db630136980a6ba5a65b5e6686ee51009ed5f3fded6ea7e", purpose.HexValue)
	assert.Equal(t, "53956441128315394338673222674654929973131976200905067808864911710716608047742", purpose.Value.String())
}
