// +build unit

package ed25519

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, nil)
	result := m.Run()
	os.Exit(result)
}

func TestPublicKeyToP2PKey(t *testing.T) {
	expectedPeerId := "12D3KooWHSED5BoCN6ogf6e5Wk1H3pH63mT2Emki7wTTaAGD6bw8"
	publicKey, err := GetPublicSigningKey("../../build/resources/signingKey.pub.pem")
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
