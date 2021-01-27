// +build integration

package ideth

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func getTestKey() identity.Key {
	return identity.NewKey(utils.RandomByte32(), utils.ByteSliceToBigInt([]byte{123}), utils.ByteSliceToBigInt([]byte{123}), 0)
}

func initIdentity() identity.Service {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	jobManager := ctx[jobs.BootstrappedService].(jobs.Manager)
	queue := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	return NewService(client, jobManager, queue, cfg)
}

func getTestDIDContext(t *testing.T, did identity.DID) context.Context {
	cfg.Set("identityId", did.ToAddress().String())
	aCtx := testingconfig.CreateAccountContext(t, cfg)
	return aCtx
}

func deployIdentityContract(t *testing.T) *identity.DID {
	factory := ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	accountCtx := testingconfig.CreateAccountContext(t, cfg)
	did, err := factory.CreateIdentity(accountCtx)
	assert.Nil(t, err, "create identity should be successful")

	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	contractCode, err := client.GetEthClient().CodeAt(context.Background(), did.ToAddress(), nil)
	assert.Nil(t, err, "should be successful to get the contract code")

	assert.Equal(t, true, len(contractCode) > 3000, "current contract code should be around 3378 bytes")
	return did
}

func addKey(aCtx context.Context, t *testing.T, did identity.DID, idSrv identity.Service, testKey identity.Key) {
	err := idSrv.AddKey(aCtx, testKey)
	assert.Nil(t, err, "add key should be successful")

	response, err := idSrv.GetKey(did, testKey.GetKey())
	assert.Nil(t, err, "get Key should be successful")
	assert.Equal(t, testKey.GetPurpose(), response.Purposes[0], "key should have the same purpose")
}

func TestServiceAddKey_successful(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	testKey := getTestKey()
	addKey(aCtx, t, *did, idSrv, testKey)

	resetDefaultCentID()
}

func TestServiceAddKey_fail(t *testing.T) {
	testKey := getTestKey()
	did := identity.NewDID(common.BytesToAddress(utils.RandomSlice(20)))
	aCtx := getTestDIDContext(t, did)
	idSrv := initIdentity()

	err := idSrv.AddKey(aCtx, testKey)
	assert.Nil(t, err, "add key should be successful")

	_, err = idSrv.GetKey(did, testKey.GetKey())
	assert.Error(t, err, "no contract code at given address")
	resetDefaultCentID()

}

func TestService_AddMultiPurposeKey(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	key := utils.RandomByte32()
	purposeOne := utils.ByteSliceToBigInt([]byte{123})
	purposeTwo := utils.ByteSliceToBigInt([]byte{42})
	purposes := []*big.Int{purposeOne, purposeTwo}
	keyType := utils.ByteSliceToBigInt([]byte{137})

	err := idSrv.AddMultiPurposeKey(aCtx, key, purposes, keyType)
	assert.Nil(t, err, "add key with multiple purposes should be successful")

	response, err := idSrv.GetKey(*did, key)
	assert.Nil(t, err, "get Key should be successful")

	assert.Equal(t, purposeOne, response.Purposes[0], "key should have the same first purpose")
	assert.Equal(t, purposeTwo, response.Purposes[1], "key should have the same second purpose")
	resetDefaultCentID()
}

func TestService_RevokeKey(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	testKey := getTestKey()
	addKey(aCtx, t, *did, idSrv, testKey)

	response, err := idSrv.GetKey(*did, testKey.GetKey())
	assert.Equal(t, uint32(0), response.RevokedAt, "key should be not revoked")

	err = idSrv.RevokeKey(aCtx, testKey.GetKey())
	assert.NoError(t, err)

	//check if key is revoked
	response, err = idSrv.GetKey(*did, testKey.GetKey())
	assert.Nil(t, err, "get Key should be successful")
	assert.NotEqual(t, uint32(0), response.RevokedAt, "key should be revoked")

	resetDefaultCentID()
}

func TestExists(t *testing.T) {
	did := deployIdentityContract(t)

	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	err := idSrv.Exists(aCtx, *did)
	assert.Nil(t, err, "identity contract should exist")

	err = idSrv.Exists(aCtx, identity.NewDID(common.BytesToAddress(utils.RandomSlice(20))))
	assert.Error(t, err, "identity contract should not exist")
	resetDefaultCentID()
}

func TestValidateKey(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	testKey := getTestKey()
	addKey(aCtx, t, *did, idSrv, testKey)

	key32 := testKey.GetKey()

	var purpose *big.Int
	purpose = big.NewInt(123) // test purpose

	err := idSrv.ValidateKey(aCtx, *did, utils.Byte32ToSlice(key32), purpose, nil)
	assert.Nil(t, err, "key with purpose should exist")

	purpose = big.NewInt(1) // false purpose
	err = idSrv.ValidateKey(aCtx, *did, utils.Byte32ToSlice(key32), purpose, nil)
	assert.Error(t, err, "key with purpose should not exist")
	resetDefaultCentID()
}

func TestValidateKey_revoked(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	testKey := getTestKey()
	addKey(aCtx, t, *did, idSrv, testKey)

	err := idSrv.RevokeKey(aCtx, testKey.GetKey())
	assert.NoError(t, err)

	key32 := testKey.GetKey()

	var purpose *big.Int
	purpose = big.NewInt(123) // test purpose

	err = idSrv.ValidateKey(aCtx, *did, utils.Byte32ToSlice(key32), purpose, nil)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "for purpose [123] has been revoked and not valid anymore")
	}

	beforeRevocation := time.Now().Add(-20 * time.Second)
	err = idSrv.ValidateKey(aCtx, *did, utils.Byte32ToSlice(key32), purpose, &beforeRevocation)
	assert.NoError(t, err)

	afterRevocation := time.Now()
	err = idSrv.ValidateKey(aCtx, *did, utils.Byte32ToSlice(key32), purpose, &afterRevocation)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "for purpose [123] has been revoked before provided time")
	}
	resetDefaultCentID()
}

func addP2PKeyTestGetClientP2PURL(t *testing.T) (*identity.DID, string) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	p2pKey := utils.RandomByte32()

	testKey := identity.NewKey(p2pKey, &(identity.KeyPurposeP2PDiscovery.Value), utils.ByteSliceToBigInt([]byte{123}), 0)
	addKey(aCtx, t, *did, idSrv, testKey)

	url, err := idSrv.GetClientP2PURL(*did)
	assert.Nil(t, err, "should return p2p url")

	p2pID, err := ed25519.PublicKeyToP2PKey(p2pKey)
	assert.Nil(t, err)

	expectedUrl := fmt.Sprintf("/ipfs/%s", p2pID.Pretty())

	assert.Equal(t, expectedUrl, url, "ipfs url not correct")
	return did, url
}

func TestGetClientP2PURL(t *testing.T) {
	addP2PKeyTestGetClientP2PURL(t)
	resetDefaultCentID()
}

func TestGetClientP2PURLs(t *testing.T) {
	didA, urlA := addP2PKeyTestGetClientP2PURL(t)
	didB, urlB := addP2PKeyTestGetClientP2PURL(t)
	idSrv := initIdentity()

	urls, err := idSrv.GetClientsP2PURLs([]*identity.DID{didA, didB})
	assert.Nil(t, err)

	assert.Equal(t, urlA, urls[0], "p2p url should be the same")
	assert.Equal(t, urlB, urls[1], "p2p url should be the same")
	resetDefaultCentID()
}
