// +build unit

package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestManagementPurpose(t *testing.T) {
	b32, err := utils.ByteArrayTo32BytesLeftPadded([]byte{1})
	assert.NoError(t, err)
	mgmtHex := hex.EncodeToString(b32[:])
	assert.Equal(t, KeyPurposeManagement.HexValue, mgmtHex)
}

func TestActionPurpose(t *testing.T) {
	b32, err := utils.ByteArrayTo32BytesLeftPadded([]byte{2})
	assert.NoError(t, err)
	actionHex := hex.EncodeToString(b32[:])
	assert.Equal(t, KeyPurposeAction.HexValue, actionHex)
}

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

func TestKeyPurposeManagement(t *testing.T) {
	purpose := KeyPurposeManagement
	assert.Equal(t, "MANAGEMENT", purpose.Name)
	assert.Equal(t, "0000000000000000000000000000000000000000000000000000000000000001", purpose.HexValue)
	assert.Equal(t, "1", purpose.Value.String())
}

func TestKeyPurposeAction(t *testing.T) {
	purpose := KeyPurposeAction
	assert.Equal(t, "ACTION", purpose.Name)
	assert.Equal(t, "0000000000000000000000000000000000000000000000000000000000000002", purpose.HexValue)
	assert.Equal(t, "2", purpose.Value.String())
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

func TestDID_Marshaling(t *testing.T) {
	did0, err := NewDIDFromString("0x366b41162a53fd75d95d31dD6d1C4d83bD436BBe")
	did1, err := NewDIDFromString("0x8780e1143036b4c979fE253128a48074093f1987")
	assert.NoError(t, err)
	mdid0, err := did0.MarshalJSON()
	assert.NoError(t, err)

	// Unmarshal with no wrapped quotes
	err = did1.UnmarshalJSON([]byte("0x8780e1143036b4c979fE253128a48074093f1987"))
	assert.NoError(t, err)

	// Wrong DID length
	err = did1.UnmarshalJSON(append(mdid0, mdid0...))
	assert.Error(t, err)

	// nil payload
	err = did1.UnmarshalJSON(nil)
	assert.Error(t, err)

	err = did1.UnmarshalJSON(mdid0)
	assert.NoError(t, err)
	assert.Equal(t, "0x366b41162a53fd75d95d31dD6d1C4d83bD436BBe", did1.String())
}
