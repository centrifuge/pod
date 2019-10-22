// +build integration

/*

The identity contract has a method called execute which forwards
a call to another contract. In Ethereum it is a useful pattern especially
related to identity smart contracts.


The correct behaviour currently is tested by calling the anchorRepository
commit method. It could be any other contract or method as well to simple test the
correct implementation.

*/

package ideth

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/ethereum"
	id "github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// TODO will be removed after migration
func resetDefaultCentID() {
	cfg.Set("identityId", "0x010101010101")
}

func getAnchorAddress(cfg config.Configuration) common.Address {
	return cfg.GetContractAddress(config.AnchorRepo)
}

func TestExecute_successful(t *testing.T) {
	t.SkipNow() // TODO remove once pointing anchoring to cent-chain module
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()
	anchorAddress := getAnchorAddress(cfg)

	// add node Ethereum address as a action key
	// only an action key can use the execute method
	ethAccount, err := cfg.GetEthereumAccount(cfg.GetEthereumDefaultAccountName())
	assert.Nil(t, err)
	actionAddress := ethAccount.Address

	//add action key
	actionKey := utils.AddressTo32Bytes(common.HexToAddress(actionAddress))
	key := id.NewKey(actionKey, &(id.KeyPurposeAction.Value), utils.ByteSliceToBigInt([]byte{123}), 0)
	err = idSrv.AddKey(aCtx, key)
	assert.NoError(t, err)

	// init params
	preimage, hashed, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	testAnchorIdPreimage, _ := anchors.ToAnchorID(preimage)
	testAnchorId, _ := anchors.ToAnchorID(hashed)
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)
	proofs := utils.RandomByte32()

	// call execute
	_, done, err := idSrv.Execute(aCtx, anchorAddress, anchors.AnchorContractABI, "commit", testAnchorIdPreimage.BigInt(), testRootHash, proofs)
	assert.Nil(t, err, "Execute method calls should be successful")
	doneErr := <-done
	assert.NoError(t, doneErr)

	checkAnchor(t, testAnchorId, rootHash)
	resetDefaultCentID()
}

func TestExecute_fail_falseMethodName(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)

	idSrv := initIdentity()
	anchorAddress := getAnchorAddress(cfg)

	testAnchorId, _ := anchors.ToAnchorID(utils.RandomSlice(32))
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)

	proofs := utils.RandomByte32()

	_, _, err := idSrv.Execute(aCtx, anchorAddress, anchors.AnchorContractABI, "fakeMethod", testAnchorId.BigInt(), testRootHash, proofs)
	assert.Error(t, err, "should throw an error because method is not existing in abi")
	resetDefaultCentID()
}

func TestExecute_fail_MissingParam(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()
	anchorAddress := getAnchorAddress(cfg)

	testAnchorId, _ := anchors.ToAnchorID(utils.RandomSlice(32))
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)

	_, _, err := idSrv.Execute(aCtx, anchorAddress, anchors.AnchorContractABI, "commit", testAnchorId.BigInt(), testRootHash)
	assert.Error(t, err, "should throw an error because wrong params as per abi")
	resetDefaultCentID()
}

func checkAnchor(t *testing.T, anchorId anchors.AnchorID, expectedRootHash []byte) {
	anchorAddress := getAnchorAddress(cfg)
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	// check if anchor has been added
	opts, _ := client.GetGethCallOpts(false)
	anchorContract := bindAnchorContract(t, anchorAddress)
	result, err := anchorContract.GetAnchorById(opts, anchorId.BigInt())
	assert.Nil(t, err, "get anchor should be successful")
	assert.Equal(t, expectedRootHash[:], result.DocumentRoot[:], "committed anchor should be the same")
}

// Checks the standard behaviour of the anchor contract
func TestAnchorWithoutExecute_successful(t *testing.T) {
	t.SkipNow() // TODO remove once pointing anchoring to cent-chain module
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	anchorAddress := getAnchorAddress(cfg)
	anchorContract := bindAnchorContract(t, anchorAddress)

	preimage, hashed, err := crypto.GenerateHashPair(32)
	assert.NoError(t, err)
	testAnchorIdPreimage, _ := anchors.ToAnchorID(preimage)
	testAnchorId, _ := anchors.ToAnchorID(hashed)
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)

	//commit without execute method
	commitAnchorWithoutExecute(t, anchorContract, testAnchorIdPreimage, testRootHash)

	opts, _ := client.GetGethCallOpts(false)
	result, err := anchorContract.GetAnchorById(opts, testAnchorId.BigInt())
	assert.Nil(t, err, "get anchor should be successful")
	assert.Equal(t, rootHash[:], result.DocumentRoot[:], "committed anchor should be the same")

}

// Checks the standard behaviour of the anchor contract
func commitAnchorWithoutExecute(t *testing.T, anchorContract *anchors.AnchorContract, anchorId anchors.AnchorID, rootHash anchors.DocumentRoot) *ethereum.WatchTransaction {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	proofs := utils.RandomByte32()

	opts, err := client.GetTxOpts(context.Background(), cfg.GetEthereumDefaultAccountName())
	assert.NoError(t, err)
	queue := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	txManager := ctx[jobs.BootstrappedService].(jobs.Manager)

	_, done, err := txManager.ExecuteWithinJob(context.Background(), testingidentity.GenerateRandomDID(), jobs.NilJobID(), "Check TX add execute",
		func(accountID id.DID, txID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
			ethTX, err := client.SubmitTransactionWithRetries(anchorContract.Commit, opts, anchorId.BigInt(), rootHash, proofs)
			if err != nil {
				errOut <- err
				return
			}
			logTxHash(ethTX)

			res, err := ethereum.QueueEthTXStatusTask(accountID, txID, ethTX.Hash(), queue)
			if err != nil {
				errOut <- err
				return
			}

			_, err = res.Get(txMan.GetDefaultTaskTimeout())
			if err != nil {
				errOut <- err
				return
			}
			errOut <- nil
		})
	assert.Nil(t, err, "add anchor commit tx should be successful ")
	doneErr := <-done
	assert.Nil(t, doneErr, "add anchor commit tx should be successful ")

	return nil

}

func bindAnchorContract(t *testing.T, address common.Address) *anchors.AnchorContract {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	anchorContract, err := anchors.NewAnchorContract(address, client.GetEthClient())
	assert.Nil(t, err, "bind anchor contract should not throw an error")

	return anchorContract
}
