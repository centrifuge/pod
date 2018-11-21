// +build integration

package identity_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var identityService identity.Service
var cfg *config.Configuration

func TestMain(m *testing.M) {
	// Adding delay to startup (concurrency hack)
	time.Sleep(time.Second + 2)

	ctx := cc.TestFunctionalEthereumBootstrap()
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")

	identityService = ctx[identity.BootstrappedIDService].(identity.Service)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestCreateAndLookupIdentity_Integration(t *testing.T) {
	centrifugeId, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	wrongCentrifugeId := utils.RandomSlice(identity.CentIDLength)
	wrongCentrifugeId[0] = 0x0
	wrongCentrifugeId[1] = 0x0
	wrongCentrifugeId[2] = 0x0
	wrongCentrifugeId[3] = 0x0
	wrongCentrifugeIdTyped, _ := identity.ToCentID(wrongCentrifugeId)

	id, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	// LookupIdentityForID
	id, err = identityService.LookupIdentityForID(centrifugeId)
	assert.Nil(t, err, "should not error out when resolving identity")
	assert.Equal(t, centrifugeId, id.CentID(), "CentrifugeID Should match provided one")

	_, err = identityService.LookupIdentityForID(wrongCentrifugeIdTyped)
	assert.NotNil(t, err, "should error out when resolving wrong identity")

	exists, err := identityService.CheckIdentityExists(wrongCentrifugeIdTyped)
	assert.NotNil(t, err, "should err when looking for incorrect identity")
	assert.False(t, exists)

	// Add Key
	key := utils.RandomSlice(32)
	confirmations, err = id.AddKeyToIdentity(context.Background(), 1, key)
	assert.Nil(t, err, "should not error out when adding key to identity")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchReceivedIdentity := <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	recKey, err := id.LastKeyForPurpose(1)
	assert.Nil(t, err)
	assert.Equal(t, key, recKey)

	_, err = id.LastKeyForPurpose(2)
	assert.NotNil(t, err)

}

func TestAddKeyFromConfig(t *testing.T) {
	centrifugeId, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	defaultCentrifugeId := cfg.GetString("identityId")
	cfg.Set("identityId", centrifugeId.String())
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	_, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	err = identityService.AddKeyFromConfig(identity.KeyPurposeEthMsgAuth)
	assert.Nil(t, err, "should not error out")

	cfg.Set("identityId", defaultCentrifugeId)
}

func TestAddKeyFromConfig_IdentityDoesNotExist(t *testing.T) {
	centrifugeId, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	defaultCentrifugeId := cfg.GetString("identityId")
	cfg.Set("identityId", centrifugeId.String())
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")

	err := identityService.AddKeyFromConfig(identity.KeyPurposeEthMsgAuth)
	assert.NotNil(t, err, "should error out")

	cfg.Set("identityId", defaultCentrifugeId)
}

func TestCreateAndLookupIdentity_Integration_Concurrent(t *testing.T) {
	var centIds [5]identity.CentID
	var identityConfirmations [5]<-chan *identity.WatchIdentity
	var err error
	for ix := 0; ix < 5; ix++ {
		centId, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
		centIds[ix] = centId
		_, identityConfirmations[ix], err = identityService.CreateIdentity(centId)
		assert.Nil(t, err, "should not error out upon identity creation")
	}

	for ix := 0; ix < 5; ix++ {
		watchSingleIdentity := <-identityConfirmations[ix]
		id, err := identityService.LookupIdentityForID(watchSingleIdentity.Identity.CentID())
		assert.Nil(t, err, "should not error out upon identity resolution")
		assert.Equal(t, centIds[ix], id.CentID(), "Should have the ID that was passed into create function [%v]", id.CentID())
	}
}

func TestEthereumIdentityService_GetIdentityAddress(t *testing.T) {
	centrifugeId, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	_, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")
	<-confirmations
	addr, err := identityService.GetIdentityAddress(centrifugeId)
	assert.Nil(t, err)
	assert.True(t, len(addr) == common.AddressLength)
}

func TestEthereumIdentityService_GetIdentityAddressNonExistingID(t *testing.T) {
	_, err := identityService.GetIdentityAddress(identity.RandomCentID())
	assert.NotNil(t, err)
}
