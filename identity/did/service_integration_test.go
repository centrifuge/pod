// +build integration

package did

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/common"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"

	"github.com/centrifuge/go-centrifuge/testingutils/config"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/stretchr/testify/assert"
)

func getTestKey() Key {
	return &key{Key: utils.RandomByte32(), Purpose: utils.ByteSliceToBigInt([]byte{123}), Type: utils.ByteSliceToBigInt([]byte{123})}
}

func initIdentity() Service {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	txManager := ctx[transactions.BootstrappedService].(transactions.Manager)
	queue := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	return NewService(client, txManager, queue)
}

func getTestDIDContext(t *testing.T, did DID) context.Context {
	cfg.Set("identityId", did.ToAddress().String())
	cfg.Set("keys.ethauth.publicKey", "../../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../../build/resources/ethauth.key.pem")
	aCtx := testingconfig.CreateAccountContext(t, cfg)

	return aCtx

}

func deployIdentityContract(t *testing.T) *DID {
	factory := ctx[BootstrappedDIDFactory].(Factory)
	accountCtx := testingconfig.CreateAccountContext(t, cfg)
	did, err := factory.CreateIdentity(accountCtx)
	assert.Nil(t, err, "create identity should be successful")

	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	contractCode, err := client.GetEthClient().CodeAt(context.Background(), did.ToAddress(), nil)
	assert.Nil(t, err, "should be successful to get the contract code")

	assert.Equal(t, true, len(contractCode) > 3000, "current contract code should be around 3378 bytes")
	return did

}

func addKey(aCtx context.Context, t *testing.T, did DID, idSrv Service, testKey Key) {
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
	did := NewDIDFromString("0x123")
	aCtx := getTestDIDContext(t, did)
	idSrv := initIdentity()

	err := idSrv.AddKey(aCtx, testKey)
	assert.Nil(t, err, "add key should be successful")

	_, err = idSrv.GetKey(did, testKey.GetKey())
	assert.Error(t, err, "no contract code at given address")
	resetDefaultCentID()

}

func TestService_IsSignedWithPurpose(t *testing.T) {
	// create keys
	pk, sk, err := secp256k1.GenerateSigningKeyPair()
	address := common.HexToAddress(secp256k1.GetAddress(pk))
	address32Bytes := utils.AddressTo32Bytes(address)
	assert.Nil(t, err, "should convert a address to 32 bytes")

	// purpose
	purpose := utils.ByteSliceToBigInt([]byte{123})
	assert.Nil(t, err, "should generate signing key pair")

	// deploy identity and add key with purpose
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()
	key := &key{Key: address32Bytes, Purpose: purpose, Type: utils.ByteSliceToBigInt([]byte{123})}
	err = idSrv.AddKey(aCtx, key)
	assert.Nil(t, err, "add key should be successful")

	// sign a msg with keypair
	msg := utils.RandomByte32()
	signature, err := secp256k1.SignEthereum(msg[:], sk)
	assert.Nil(t, err, "should sign a message")

	//correct signature and purpose
	signed, err := idSrv.IsSignedWithPurpose(*did, msg, signature, purpose)
	assert.Nil(t, err, "sign verify should not throw an error")
	assert.True(t, signed, "signature should be correct")

	//false purpose
	falsePurpose := utils.ByteSliceToBigInt([]byte{42})
	signed, err = idSrv.IsSignedWithPurpose(*did, msg, signature, falsePurpose)
	assert.Nil(t, err, "sign verify should not throw an error")
	assert.False(t, signed, "signature should be false (wrong purpose)")

	//false keypair
	_, sk2, _ := secp256k1.GenerateSigningKeyPair()
	signature, err = secp256k1.SignEthereum(msg[:], sk2)
	assert.Nil(t, err, "should sign a message")
	signed, err = idSrv.IsSignedWithPurpose(*did, msg, signature, purpose)
	assert.Nil(t, err, "sign verify should not throw an error")
	assert.False(t, signed, "signature should be wrong key pair")
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
	assert.Equal(t, utils.ByteSliceToBigInt([]byte{0}), response.RevokedAt, "key should be not revoked")

	idSrv.RevokeKey(aCtx, testKey.GetKey())

	//check if key is revoked
	response, err = idSrv.GetKey(*did, testKey.GetKey())
	assert.Nil(t, err, "get Key should be successful")
	assert.NotEqual(t, utils.ByteSliceToBigInt([]byte{0}), response.RevokedAt, "key should be revoked")

	resetDefaultCentID()
}

func TestExists(t *testing.T) {
	did := deployIdentityContract(t)

	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	err := idSrv.Exists(aCtx, *did)
	assert.Nil(t, err, "identity contract should exist")

	err = idSrv.Exists(aCtx, NewDIDFromString("0x123"))
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

	var purpose int64
	purpose = 123 // test purpose

	err := idSrv.ValidateKey(aCtx, *did, utils.Byte32ToSlice(key32), purpose)
	assert.Nil(t, err, "key with purpose should exist")

	purpose = 1 //false purpose
	err = idSrv.ValidateKey(aCtx, *did, utils.Byte32ToSlice(key32), purpose)
	assert.Error(t, err, "key with purpose should not exist")
	resetDefaultCentID()

}

func addP2PKeyTestGetClientP2PURL(t *testing.T) (*DID, string) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	p2pKey := utils.RandomByte32()

	testKey := &key{Key: p2pKey, Purpose: utils.ByteSliceToBigInt([]byte{identity.KeyPurposeP2P}), Type: utils.ByteSliceToBigInt([]byte{123})}
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

	urls, err := idSrv.GetClientsP2PURLs([]*DID{didA, didB})
	assert.Nil(t, err)

	assert.Equal(t, urlA, urls[0], "p2p url should be the same")
	assert.Equal(t, urlB, urls[1], "p2p url should be the same")
	resetDefaultCentID()

}
