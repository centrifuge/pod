// +build ethereum

package identity_test

import (
	"testing"
	"github.com/spf13/viper"
	"os"
	"github.com/stretchr/testify/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
)

func TestMain(m *testing.M) {
	//for now set up the env vars manually in integration test
	//TODO move to generalized config once it is available
	viper.BindEnv("ethereum.gethSocket", "CENT_ETHEREUM_GETH_SOCKET")
	viper.BindEnv("ethereum.gasLimit", "CENT_ETHEREUM_GASLIMIT")
	viper.BindEnv("ethereum.gasPrice", "CENT_ETHEREUM_GASPRICE")
	viper.BindEnv("ethereum.contextWaitTimeout", "CENT_ETHEREUM_CONTEXTWAITTIMEOUT")
	viper.BindEnv("ethereum.accounts.main.password", "CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD")
	viper.BindEnv("ethereum.accounts.main.key", "CENT_ETHEREUM_ACCOUNTS_MAIN_KEY")
	viper.BindEnv("identity.ethereum.identityFactoryAddress", "CENT_IDENTITY_ETHEREUM_IDENTITYFACTORYADDRESS")
	viper.BindEnv("identity.ethereum.identityRegistryAddress", "CENT_IDENTITY_ETHEREUM_IDENTITYREGISTRYADDRESS")
	viper.Set("identity.ethereum.enabled", "true")

	result := m.Run()
	os.Exit(result)
}

func TestCreateAndResolveIdentity_Integration(t *testing.T) {
	centrifugeId := tools.RandomString32()
	nodePeerId := tools.RandomByte32()
	var m = make(map[int][]identity.IdentityKey)
	confirmations := make(chan *identity.Identity, 1)
	m[1] = append(m[1], identity.IdentityKey{nodePeerId})
	identityObj := identity.Identity{ CentrifugeId: centrifugeId, Keys: m }

	err := identity.CreateIdentity(identityObj, confirmations)
	assert.Nil(t, err, "should not error out when creating identity")

	registeredIdentity := <-confirmations
	assert.Equal(t, centrifugeId, registeredIdentity.CentrifugeId, "Resulting Identity should have the same ID as the input")

	id, err := identity.ResolveP2PIdentityForId(centrifugeId, 1)
	assert.Nil(t, err, "should not error out when resolving identity")
	assert.Equal(t, centrifugeId, id.CentrifugeId, "CentrifugeId Should match provided one")
	assert.Equal(t, 0, len(id.Keys), "Identity Should have empty map of keys")
}

func TestCheckIdentityExists_Integration(t *testing.T) {
	centrifugeId := tools.RandomString32()
	nodePeerId := tools.RandomByte32()
	var m = make(map[int][]identity.IdentityKey)
	confirmations := make(chan *identity.Identity, 1)
	m[1] = append(m[1], identity.IdentityKey{nodePeerId})
	identityObj := identity.Identity{ CentrifugeId: centrifugeId, Keys: m }

	err := identity.CreateIdentity(identityObj, confirmations)
	assert.Nil(t, err, "should not error out when creating identity")

	registeredIdentity := <-confirmations
	assert.Equal(t, centrifugeId, registeredIdentity.CentrifugeId, "Resulting Identity should have the same ID as the input")

	exists, err := identity.CheckIdentityExists(centrifugeId)
	assert.Nil(t, err, "should not error out when looking for correct identity")
	assert.Equal(t, true, exists, "Identity Should Exists")

	wrongCentrifugeId := tools.RandomString32()
	exists, err = identity.CheckIdentityExists(wrongCentrifugeId)
	assert.Nil(t, err, "should not error out when missing identity")
	assert.Equal(t, false, exists, "Identity Should Exists")
}

func TestCreateIdentityAndAddKey_Integration(t *testing.T) {
	centrifugeId := tools.RandomString32()
	nodePeerId := tools.RandomByte32()
	var m = make(map[int][]identity.IdentityKey)
	confirmations := make(chan *identity.Identity, 1)
	m[1] = append(m[1], identity.IdentityKey{nodePeerId})
	identityObj := identity.Identity{ CentrifugeId: centrifugeId, Keys: m }

	err := identity.CreateIdentity(identityObj, confirmations)
	assert.Nil(t, err, "should not error out when creating identity")

	registeredIdentity := <-confirmations
	assert.Equal(t, centrifugeId, registeredIdentity.CentrifugeId, "Resulting Identity should have the same ID as the input")

	id, err := identity.ResolveP2PIdentityForId(centrifugeId, 1)

	assert.Nil(t, err, "should not error out when resolving identity")
	assert.Equal(t, centrifugeId, id.CentrifugeId, "CentrifugeId Should match provided one")
	assert.Equal(t, 0, len(id.Keys), "Identity Should have empty map of keys")

	err = identity.AddKeyToIdentity(identityObj, 1, confirmations)
	assert.Nil(t, err, "should not error out when adding key to identity")

	receivedIdentity := <-confirmations
	assert.Equal(t, centrifugeId, receivedIdentity.CentrifugeId, "Resulting Identity should have the same ID as the input")
	assert.Equal(t, 1, len(receivedIdentity.Keys), "Resulting Identity Key Map should have expected length")
	assert.Equal(t,1, len(receivedIdentity.Keys[1]), "Resulting Identity Key Type list should have expected length")
	assert.Equal(t, m[1][0].Key, receivedIdentity.Keys[1][0].Key, "Resulting Identity Key should match the one requested")

	// Double check that Key Exists in Identity
	id, err = identity.ResolveP2PIdentityForId(centrifugeId, 1)

	assert.Nil(t, err, "should not error out when resolving identity")
	assert.Equal(t, centrifugeId, id.CentrifugeId, "CentrifugeId Should match provided one")
	assert.Equal(t, 1, len(id.Keys), "Identity Should have empty map of keys")
	assert.Equal(t, m[1][0].Key, id.Keys[1][0].Key, "Resulting Identity Key should match the one requested")
}

func TestCreateAndResolveIdentity_Integration_Concurrent(t *testing.T) {
	var submittedIds [5]string
	nodePeerId := tools.RandomByte32()
	var m = make(map[int][]identity.IdentityKey)
	m[1] = append(m[1], identity.IdentityKey{nodePeerId})
	howMany := cap(submittedIds)
	confirmations := make(chan *identity.Identity, howMany)

	for ix := 0; ix < howMany; ix++ {
		centId := tools.RandomString32()
		identityObj := identity.Identity{ CentrifugeId: centId, Keys: m }
		submittedIds[ix] = centId

		err := identity.CreateIdentity(identityObj, confirmations)
		assert.Nil(t, err, "should not error out upon identity creation")
	}

	for ix := 0; ix < howMany; ix++ {
		singleIdentity := <-confirmations
		id, err := identity.ResolveP2PIdentityForId(singleIdentity.CentrifugeId, 1)
		assert.Nil(t, err, "should not error out upon identity resolution")
		assert.Contains(t, submittedIds, id.CentrifugeId , "Should have the ID that was passed into create function [%v]", id.CentrifugeId)
	}
}