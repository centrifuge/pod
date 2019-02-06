// +build integration

package did

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"

	"github.com/centrifuge/go-centrifuge/testingutils/config"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/stretchr/testify/assert"
)

func getTestKey() Key {
	return &key{Key: utils.RandomByte32(), Purpose: utils.ByteSliceToBigInt([]byte{123}), Type: utils.ByteSliceToBigInt([]byte{123})}
}

func initIdentity(config config.Configuration) Identity {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	txManager := ctx[transactions.BootstrappedService].(transactions.Manager)
	queue := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	return NewIdentity(config, client, txManager, queue)
}

func getTestDIDContext(t *testing.T, did DID) context.Context {
	cfg.Set("identityId", did.toAddress().String())
	cfg.Set("keys.ethauth.publicKey", "../../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../../build/resources/ethauth.key.pem")
	aCtx := testingconfig.CreateAccountContext(t, cfg)

	return aCtx

}

func deployIdentityContract(t *testing.T) *DID {
	service := ctx[BootstrappedDIDService].(Service)
	accountCtx := testingconfig.CreateAccountContext(t, cfg)
	did, err := service.CreateIdentity(accountCtx)
	assert.Nil(t, err, "create identity should be successful")

	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	contractCode, err := client.GetEthClient().CodeAt(context.Background(), did.toAddress(), nil)
	assert.Nil(t, err, "should be successful to get the contract code")

	assert.Equal(t, true, len(contractCode) > 3000, "current contract code should be arround 3378 bytes")
	return did

}

func TestAddKey_successful(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity(cfg)

	testKey := getTestKey()

	err := idSrv.AddKey(aCtx, testKey)
	assert.Nil(t, err, "add key should be successful")

	response, err := idSrv.GetKey(aCtx, testKey.GetKey())
	assert.Nil(t, err, "get Key should be successful")

	assert.Equal(t, testKey.GetPurpose(), response.Purposes[0], "key should have the same purpose")
	resetDefaultCentID()
}

func TestAddKey_fail(t *testing.T) {
	testKey := getTestKey()
	did := NewDIDFromString("0x123")
	aCtx := getTestDIDContext(t, did)
	idSrv := initIdentity(cfg)

	err := idSrv.AddKey(aCtx, testKey)
	assert.Nil(t, err, "add key should be successful")

	_, err = idSrv.GetKey(aCtx, testKey.GetKey())
	assert.Error(t, err, "no contract code at given address")
	resetDefaultCentID()

}
