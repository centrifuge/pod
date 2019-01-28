// +build integration

package did

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func commitAnchorDirect(t*testing.T,anchorContract *anchors.AnchorContract, anchorId anchors.AnchorID, rootHash anchors.DocumentRoot) *ethereum.WatchTransaction  {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	proofs := [][anchors.DocumentProofLength]byte{utils.RandomByte32()}

	opts, err := client.GetTxOpts(cfg.GetEthereumDefaultAccountName())
	tx, err := client.SubmitTransactionWithRetries(anchorContract.Commit,opts,anchorId.BigInt(),rootHash,proofs)

	assert.Nil(t, err, "submit transaction should be successful")

	watchTrans := make(chan *ethereum.WatchTransaction)
	go waitForTransaction(client, tx.Hash(), watchTrans)
	return  <- watchTrans
}

func bindAnchorContract(t *testing.T, address common.Address) *anchors.AnchorContract{
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	anchorContract, err := anchors.NewAnchorContract(address, client.GetEthClient())
	assert.Nil(t, err, "bind anchor contract should not throw an error")

	return anchorContract
}


func TestExecute_successful(t *testing.T) {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	idSrv := ctx[BootstrappedDID].(Identity)
	did := deployIdentityContract(t)
	anchorAddress := getAnchorAddress()


	testAnchorId, _ := anchors.ToAnchorID(utils.RandomSlice(32))
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)

	proofs := [][anchors.DocumentProofLength]byte{utils.RandomByte32()}

	watchTrans, err := idSrv.Execute(did, anchorAddress,anchors.AnchorContractABI,"commit", testAnchorId.BigInt(),testRootHash,proofs)
	assert.Nil(t, err, "Execute method calls should be successful")

	txStatus := <-watchTrans
	assert.Equal(t, ethereum.TransactionStatusSuccess, txStatus.Status, "transaction should be successful")

	// check if anchor has been added
	opts, _ := client.GetGethCallOpts(false)
	anchorContract := bindAnchorContract(t, anchorAddress)
	result, err := anchorContract.GetAnchorById(opts,testAnchorId.BigInt())
	assert.Nil(t, err, "get anchor should be successful")
	assert.Equal(t, rootHash[:],result.DocumentRoot[:], "committed anchor should be the same")

}

func TestExecute_withAnchor_successful(t *testing.T) {
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	anchorAddress := getAnchorAddress()
	anchorContract := bindAnchorContract(t, anchorAddress)

	testAnchorId, _ := anchors.ToAnchorID(utils.RandomSlice(32))
	rootHash := utils.RandomSlice(32)
	testRootHash, _ := anchors.ToDocumentRoot(rootHash)

	//commit without execute method
	txStatus := commitAnchorDirect(t,anchorContract, testAnchorId,testRootHash)
	assert.Equal(t, ethereum.TransactionStatusSuccess, txStatus.Status, "transaction should be successful")

	opts, _ := client.GetGethCallOpts(false)
	result, err := anchorContract.GetAnchorById(opts,testAnchorId.BigInt())
	assert.Nil(t, err, "get anchor should be successful")
	assert.Equal(t, rootHash[:],result.DocumentRoot[:], "committed anchor should be the same")


}