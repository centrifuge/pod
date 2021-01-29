// +build integration

/*

The identity contract has a method called execute which forwards
a call to another contract. In Ethereum it is a useful pattern especially
related to identity smart contracts.
*/

package ideth

import (
	"testing"

	id "github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// TODO will be removed after migration
func resetDefaultCentID() {
	cfg.Set("identityId", "0x010101010101")
}

func TestExecute_successful(t *testing.T) {
	t.SkipNow() // TODO enable
	did := DeployIdentity(t, ctx, cfg)
	aCtx := getTestDIDContext(t, did)
	idSrv := initIdentity()

	// add node Ethereum address as a action key
	// only an action key can use the execute method
	ethAccount, err := cfg.GetEthereumAccount(cfg.GetEthereumDefaultAccountName())
	assert.Nil(t, err)
	actionAddress := ethAccount.Address

	// add action key
	actionKey := utils.AddressTo32Bytes(common.HexToAddress(actionAddress))
	key := id.NewKey(actionKey, &(id.KeyPurposeAction.Value), utils.ByteSliceToBigInt([]byte{123}), 0)
	err = idSrv.AddKey(aCtx, key)
	assert.NoError(t, err)
}
