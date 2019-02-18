// +build integration

/*

The identity contract has a method called execute which forwards
a call to another contract. In Ethereum it is a useful pattern especially
related to identity smart contracts.


The correct behaviour currently is tested by calling the anchorRepository
commit method. It could be any other contract or method as well to simple test the
correct implementation.

*/

package did

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/ethereum"
	id "github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// TODO will be removed after migration
func resetDefaultCentID() {
	cfg.Set("identityId", "0x010101010101")
}

func TestExecute_successful(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()
	anchorAddress := getAnchorAddress()

	// init params
	testAnchorId, _ := anchors.ToAnchorID(utils.RandomSlice(32))
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)
	proofs := [][anchors.DocumentProofLength]byte{utils.RandomByte32()}

	err := idSrv.Execute(aCtx, anchorAddress, anchors.AnchorContractABI, "commit", testAnchorId.BigInt(), testRootHash, proofs)
	assert.Nil(t, err, "Execute method calls should be successful")

	checkAnchor(t, testAnchorId, rootHash)
	resetDefaultCentID()
}

func TestExecute_fail_falseMethodName(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)

	idSrv := initIdentity()
	anchorAddress := getAnchorAddress()

	testAnchorId, _ := anchors.ToAnchorID(utils.RandomSlice(32))
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)

	proofs := [][anchors.DocumentProofLength]byte{utils.RandomByte32()}

	err := idSrv.Execute(aCtx, anchorAddress, anchors.AnchorContractABI, "fakeMethod", testAnchorId.BigInt(), testRootHash, proofs)
	assert.Error(t, err, "should throw an error because method is not existing in abi")

	resetDefaultCentID()
}

func TestExecute_fail_MissingParam(t *testing.T) {
	did := deployIdentityContract(t)
	aCtx := getTestDIDContext(t, *did)
	idSrv := initIdentity()
	anchorAddress := getAnchorAddress()

	testAnchorId, _ := anchors.ToAnchorID(utils.RandomSlice(32))
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)

	err := idSrv.Execute(aCtx, anchorAddress, anchors.AnchorContractABI, "commit", testAnchorId.BigInt(), testRootHash)
	assert.Error(t, err, "should throw an error because method is not existing in abi")
	resetDefaultCentID()
}

func checkAnchor(t *testing.T, anchorId anchors.AnchorID, expectedRootHash []byte) {
	anchorAddress := getAnchorAddress()
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
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	anchorAddress := getAnchorAddress()
	anchorContract := bindAnchorContract(t, anchorAddress)

	testAnchorId, _ := anchors.ToAnchorID(utils.RandomSlice(32))
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)

	//commit without execute method
	commitAnchorWithoutExecute(t, anchorContract, testAnchorId, testRootHash)

	opts, _ := client.GetGethCallOpts(false)
	result, err := anchorContract.GetAnchorById(opts, testAnchorId.BigInt())
	assert.Nil(t, err, "get anchor should be successful")
	assert.Equal(t, rootHash[:], result.DocumentRoot[:], "committed anchor should be the same")

}

// Checks the standard behaviour of the anchor contract
func commitAnchorWithoutExecute(t *testing.T, anchorContract *anchors.AnchorContract, anchorId anchors.AnchorID, rootHash anchors.DocumentRoot) *ethereum.WatchTransaction {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	proofs := [][anchors.DocumentProofLength]byte{utils.RandomByte32()}

	opts, err := client.GetTxOpts(cfg.GetEthereumDefaultAccountName())

	queue := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	txManager := ctx[transactions.BootstrappedService].(transactions.Manager)

	// TODO: did can be passed instead of randomCentID after CentID is DID
	_, done, err := txManager.ExecuteWithinTX(context.Background(), id.RandomCentID(), transactions.NilTxID(), "Check TX add execute",
		func(accountID id.CentID, txID transactions.TxID, txMan transactions.Manager, errOut chan<- error) {
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
	isDone := <-done
	// non async task

	assert.True(t, isDone, "add anchor commit tx should be successful ")

	return nil

}

func bindAnchorContract(t *testing.T, address common.Address) *anchors.AnchorContract {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	anchorContract, err := anchors.NewAnchorContract(address, client.GetEthClient())
	assert.Nil(t, err, "bind anchor contract should not throw an error")

	return anchorContract
}
