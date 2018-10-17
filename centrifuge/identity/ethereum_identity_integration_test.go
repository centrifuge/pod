// +build integration

package identity_test

import (
	"context"
	"os"
	"testing"
	"time"

	"log"

	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var identityService identity.Service

func TestMain(m *testing.M) {
	// Adding delay to startup (concurrency hack)
	// TODO: look for other sleep statements in tests and fix the underlying issues
	time.Sleep(time.Second + 2)

	cc.DONT_USE_FOR_UNIT_TESTS_TestFunctionalEthereumBootstrap()
	config.Config.V.Set("keys.signing.publicKey", "../../example/resources/signingKey.pub.pem")
	config.Config.V.Set("keys.signing.privateKey", "../../example/resources/signingKey.key.pem")

	identityService = identity.IDService
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestCreateAndLookupIdentity_Integration(t *testing.T) {
	centrifugeId, _ := identity.ToCentID(tools.RandomSlice(identity.CentIDLength))
	wrongCentrifugeId := tools.RandomSlice(identity.CentIDLength)
	wrongCentrifugeId[0] = 0x0
	wrongCentrifugeId[1] = 0x0
	wrongCentrifugeId[2] = 0x0
	wrongCentrifugeId[3] = 0x0
	wrongCentrifugeIdTyped, _ := identity.ToCentID(wrongCentrifugeId)

	id, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.GetCentrifugeID(), "Resulting Identity should have the same ID as the input")

	// LookupIdentityForID
	id, err = identityService.LookupIdentityForID(centrifugeId)
	assert.Nil(t, err, "should not error out when resolving identity")
	assert.Equal(t, centrifugeId, id.GetCentrifugeID(), "CentrifugeID Should match provided one")

	_, err = identityService.LookupIdentityForID(wrongCentrifugeIdTyped)
	assert.NotNil(t, err, "should error out when resolving wrong identity")

	exists, err := identityService.CheckIdentityExists(wrongCentrifugeIdTyped)
	assert.NotNil(t, err, "should err when looking for incorrect identity")
	assert.False(t, exists)

	// Add Key
	key := tools.RandomSlice(32)
	confirmations, err = id.AddKeyToIdentity(context.Background(), 1, key)
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

func TestAddKeyFromConfig(t *testing.T) {
	centrifugeId, _ := identity.ToCentID(tools.RandomSlice(identity.CentIDLength))
	defaultCentrifugeId := config.Config.V.GetString("identityId")
	config.Config.V.Set("identityId", centrifugeId.String())
	config.Config.V.Set("keys.ethauth.publicKey", "../../example/resources/ethauth.pub.pem")
	config.Config.V.Set("keys.ethauth.privateKey", "../../example/resources/ethauth.key.pem")
	_, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.GetCentrifugeID(), "Resulting Identity should have the same ID as the input")

	err = identity.AddKeyFromConfig(identity.KeyPurposeEthMsgAuth)
	assert.Nil(t, err, "should not error out")

	config.Config.V.Set("identityId", defaultCentrifugeId)
}

func TestAddKeyFromConfig_IdentityDoesNotExist(t *testing.T) {
	centrifugeId, _ := identity.ToCentID(tools.RandomSlice(identity.CentIDLength))
	defaultCentrifugeId := config.Config.V.GetString("identityId")
	config.Config.V.Set("identityId", centrifugeId.String())
	config.Config.V.Set("keys.ethauth.publicKey", "../../example/resources/ethauth.pub.pem")
	config.Config.V.Set("keys.ethauth.privateKey", "../../example/resources/ethauth.key.pem")

	err := identity.AddKeyFromConfig(identity.KeyPurposeEthMsgAuth)
	assert.NotNil(t, err, "should error out")

	config.Config.V.Set("identityId", defaultCentrifugeId)
}

func TestCreateAndLookupIdentity_Integration_Concurrent(t *testing.T) {
	var centIds [5]identity.CentID
	var identityConfirmations [5]<-chan *identity.WatchIdentity
	var err error
	for ix := 0; ix < 5; ix++ {
		centId, _ := identity.ToCentID(tools.RandomSlice(identity.CentIDLength))
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

func TestEthereumIdentityService_GetIdentityAddress(t *testing.T) {
	centrifugeId, _ := identity.ToCentID(tools.RandomSlice(identity.CentIDLength))
	_, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")
	<-confirmations
	addr, err := identityService.GetIdentityAddress(centrifugeId)
	assert.Nil(t, err)
	assert.True(t, len(addr) == common.AddressLength)
}

func TestEthereumIdentityService_GetIdentityAddressNonExistingID(t *testing.T) {
	addr, err := identityService.GetIdentityAddress(identity.NewRandomCentID())
	log.Printf("TestEthereumIdentityService_GetIdentityAddressNonExistingID address %x , err %v")
	assert.NotNil(t, err)
	assert.True(t, len(addr) == 0)
}
