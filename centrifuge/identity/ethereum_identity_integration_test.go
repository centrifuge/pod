// +build ethereum

package identity_test

import (
	"os"
	"testing"
	"time"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

var identityService identity.Service

func TestMain(m *testing.M) {
	// Adding delay to startup (concurrency hack)
	// TODO: look for other sleep statements in tests and fix the underlying issues
	time.Sleep(time.Second + 2)

	cc.TestFunctionalEthereumBootstrap()
	config.Config.V.Set("keys.signing.publicKey", "../../example/resources/signingKey.pub.pem")
	config.Config.V.Set("keys.signing.privateKey", "../../example/resources/signingKey.key.pem")

	identityService = &identity.EthereumIdentityService{}
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestCreateAndLookupIdentity_Integration(t *testing.T) {
	centrifugeId := tools.RandomSlice(identity.CentIdByteLength)
	wrongCentrifugeId := tools.RandomSlice(identity.CentIdByteLength)
	wrongCentrifugeId[0] = 0x0
	wrongCentrifugeId[1] = 0x0
	wrongCentrifugeId[2] = 0x0
	wrongCentrifugeId[3] = 0x0

	id, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.GetCentrifugeID(), "Resulting Identity should have the same ID as the input")

	// LookupIdentityForID
	id, err = identityService.LookupIdentityForID(centrifugeId)
	assert.Nil(t, err, "should not error out when resolving identity")
	assert.Equal(t, centrifugeId, id.GetCentrifugeID(), "CentrifugeId Should match provided one")

	wrongId, err := identityService.LookupIdentityForID(wrongCentrifugeId)
	assert.NotNil(t, err, "should error out when resolving wrong identity")

	// CheckIdentityExists
	exists, err := id.CheckIdentityExists()
	assert.Nil(t, err, "should not error out when looking for correct identity")
	assert.True(t, exists)

	exists, err = identityService.CheckIdentityExists(wrongCentrifugeId)
	assert.NotNil(t, err, "should err when looking for incorrect identity")
	assert.False(t, exists)

	wrongId = identity.NewEthereumIdentity()
	wrongId.SetCentrifugeID(wrongCentrifugeId)
	exists, err = wrongId.CheckIdentityExists()
	assert.NotNil(t, err, "should error out when missing identity")
	assert.False(t, exists)

	// Add Key
	key := tools.RandomSlice(32)
	confirmations, err = id.AddKeyToIdentity(1, key)
	assert.Nil(t, err, "should not error out when adding key to identity")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchReceivedIdentity := <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.GetCentrifugeID(), "Resulting Identity should have the same ID as the input")

	recKey, err := id.GetLastKeyForPurpose(1)
	assert.Nil(t, err)
	assert.Equal(t, key, recKey)

	_, err = id.GetLastKeyForPurpose(2)
	assert.NotNil(t, err)

}

func TestCreateAndLookupIdentity_Integration_Concurrent(t *testing.T) {
	var centIds [5][]byte
	var identityConfirmations [5]<-chan *identity.WatchIdentity
	var err error
	for ix := 0; ix < 5; ix++ {
		centId := tools.RandomSlice(identity.CentIdByteLength)
		centIds[ix] = centId
		_, identityConfirmations[ix], err = identityService.CreateIdentity(centId)
		assert.Nil(t, err, "should not error out upon identity creation")
	}

	for ix := 0; ix < 5; ix++ {
		watchSingleIdentity := <-identityConfirmations[ix]
		id, err := identityService.LookupIdentityForID(watchSingleIdentity.Identity.GetCentrifugeID())
		assert.Nil(t, err, "should not error out upon identity resolution")
		assert.Equal(t, centIds[ix], id.GetCentrifugeID(), "Should have the ID that was passed into create function [%v]", id.GetCentrifugeID())
	}
}
