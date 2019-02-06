// +build integration

package did

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/testingutils/config"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/stretchr/testify/assert"
)

func getTestKey() Key {
	return &key{Key: utils.RandomByte32(), Purpose: utils.ByteSliceToBigInt([]byte{123}), Type: utils.ByteSliceToBigInt([]byte{123})}
}

func initIdentity(config config.Configuration, client ethereum.Client, did *DID) Identity {
	return NewIdentity(config, client, did)
}

func deployIdentityContract(t *testing.T) *DID {
	service := ctx[BootstrappedDIDService].(Service)
	accountCtx := testingconfig.CreateAccountContext(t, cfg)
	did, err := service.CreateIdentity(accountCtx)
	assert.Nil(t, err, "create identity should be successful")

	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	time.Sleep(2000 * time.Millisecond)

	contractCode, err := client.GetEthClient().CodeAt(context.Background(), did.toAddress(), nil)
	assert.Nil(t, err, "should be successful to get the contract code")

	assert.Equal(t, true, len(contractCode) > 3000, "current contract code should be arround 3378 bytes")
	return did

}

func TestAddKey_successful(t *testing.T) {
	did := deployIdentityContract(t)
	idSrv := initIdentity(cfg, ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client), did)

	testKey := getTestKey()

	watchTrans, err := idSrv.AddKey(testKey)
	assert.Nil(t, err, "add key should be successful")

	txStatus := <-watchTrans
	assert.Equal(t, ethereum.TransactionStatusSuccess, txStatus.Status, "transactions should be successful")

	response, err := idSrv.GetKey(testKey.GetKey())
	assert.Nil(t, err, "get Key should be successful")

	assert.Equal(t, testKey.GetPurpose(), response.Purposes[0], "key should have the same purpose")
}

func TestAddKey_fail(t *testing.T) {
	testKey := getTestKey()
	did := NewDIDFromString("0x123")
	idSrv := initIdentity(cfg, ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client), &did)

	watchTrans, err := idSrv.AddKey(testKey)
	assert.Nil(t, err, "add key should be successful")

	txStatus := <-watchTrans
	// contract is not existing but status is successful
	assert.Equal(t, ethereum.TransactionStatusSuccess, txStatus.Status, "transactions")

	_, err = idSrv.GetKey(testKey.GetKey())
	assert.Error(t, err, "no contract code at given address")

}
