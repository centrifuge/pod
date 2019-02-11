// +build integration

package did

import (
	"context"
	"testing"

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
	cfg.Set("identityId", did.toAddress().String())
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

	contractCode, err := client.GetEthClient().CodeAt(context.Background(), did.toAddress(), nil)
	assert.Nil(t, err, "should be successful to get the contract code")

	assert.Equal(t, true, len(contractCode) > 3000, "current contract code should be arround 3378 bytes")
	return did

}

func TestServiceAddKey_successful(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()

	testKey := getTestKey()

	err := idSrv.AddKey(aCtx, testKey)
	assert.Nil(t, err, "add key should be successful")

	response, err := idSrv.GetKey(aCtx, testKey.GetKey())
	assert.Nil(t, err, "get Key should be successful")

	assert.Equal(t, testKey.GetPurpose(), response.Purposes[0], "key should have the same purpose")
	resetDefaultCentID()
}

func TestServiceAddKey_fail(t *testing.T) {
	testKey := getTestKey()
	did := NewDIDFromString("0x123")
	aCtx := getTestDIDContext(t, did)
	idSrv := initIdentity()

	err := idSrv.AddKey(aCtx, testKey)
	assert.Nil(t, err, "add key should be successful")

	_, err = idSrv.GetKey(aCtx, testKey.GetKey())
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
	signed, err := idSrv.IsSignedWithPurpose(aCtx, msg, signature, purpose)
	assert.Nil(t, err, "sign verify should not throw an error")
	assert.True(t, signed, "signature should be correct")

	//false purpose
	falsePurpose := utils.ByteSliceToBigInt([]byte{42})
	signed, err = idSrv.IsSignedWithPurpose(aCtx, msg, signature, falsePurpose)
	assert.Nil(t, err, "sign verify should not throw an error")
	assert.False(t, signed, "signature should be false (wrong purpose)")

	//false keypair
	_, sk2, _ := secp256k1.GenerateSigningKeyPair()
	signature, err = secp256k1.SignEthereum(msg[:], sk2)
	assert.Nil(t, err, "should sign a message")
	signed, err = idSrv.IsSignedWithPurpose(aCtx, msg, signature, purpose)
	assert.Nil(t, err, "sign verify should not throw an error")
	assert.False(t, signed, "signature should be wrong key pair")

}
