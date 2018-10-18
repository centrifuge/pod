// +build unit

package ed25519keys

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers)
	result := m.Run()
	os.Exit(result)
}

func TestPublicKeyToP2PKey(t *testing.T) {
	expectedPeerId := "12D3KooWHSED5BoCN6ogf6e5Wk1H3pH63mT2Emki7wTTaAGD6bw8"
	publicKey, err := GetPublicSigningKey("../../../example/resources/signingKey.pub.pem")
	assert.Nil(t, err)

	var bPk [32]byte
	copy(bPk[:], publicKey)
	peerId, err := PublicKeyToP2PKey(bPk)
	assert.Nil(t, err, "Should not error out")
	assert.Equal(t, expectedPeerId, peerId.Pretty())

}

func TestGetSigningKeyPairFromConfig(t *testing.T) {
	pub := config.Config.V.Get("keys.signing.publicKey")
	pri := config.Config.V.Get("keys.signing.privateKey")

	// bad public key path
	config.Config.V.Set("keys.signing.publicKey", "bad path")
	pubK, priK, err := GetSigningKeyPairFromConfig()
	assert.Error(t, err)
	assert.Nil(t, priK)
	assert.Nil(t, pubK)
	assert.Contains(t, err.Error(), "failed to read public key")
	config.Config.V.Set("keys.signing.publicKey", pub)

	// bad private key path
	config.Config.V.Set("keys.signing.privateKey", "bad path")
	pubK, priK, err = GetSigningKeyPairFromConfig()
	assert.Error(t, err)
	assert.Nil(t, priK)
	assert.Nil(t, pubK)
	assert.Contains(t, err.Error(), "failed to read private key")
	config.Config.V.Set("keys.signing.privateKey", pri)

	// success
	pubK, priK, err = GetSigningKeyPairFromConfig()
	assert.Nil(t, err)
	assert.NotNil(t, pubK)
	assert.NotNil(t, priK)
}

func TestGetIDConfig(t *testing.T) {
	pub := config.Config.V.Get("keys.signing.publicKey")

	// failed keys
	config.Config.V.Set("keys.signing.publicKey", "bad path")
	id, err := GetIDConfig()
	assert.Error(t, err)
	assert.Nil(t, id)
	assert.Contains(t, err.Error(), "failed to get signing keys")
	config.Config.V.Set("keys.signing.publicKey", pub)

	// failed identity
	gID := config.Config.V.Get("identityId")
	config.Config.V.Set("identityId", "bad id")
	id, err = GetIDConfig()
	assert.Error(t, err)
	assert.Nil(t, id)
	assert.Contains(t, err.Error(), "can't read identityId from config")
	config.Config.V.Set("identityId", gID)

	// success
	id, err = GetIDConfig()
	assert.Nil(t, err)
	assert.NotNil(t, id)
	nID, err := config.Config.GetIdentityID()
	assert.Nil(t, err)
	assert.Equal(t, id.ID, nID)
}
