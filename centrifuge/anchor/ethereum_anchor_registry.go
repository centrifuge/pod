package anchor

import (
	//"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	//"github.com/ethereum/go-ethereum/common"
	//"github.com/spf13/viper"
	"github.com/ethereum/go-ethereum/common"
	goEthereum "github.com/ethereum/go-ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"log"
	"math/big"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/ethereum/go-ethereum/core/types"
	"context"
	"time"
)

//Supported anchor schema version as stored on public registry
const ANCHOR_SCHEMA_VERSION uint = 1

type EthereumAnchorRegistry struct {
}

func (ethRegistry *EthereumAnchorRegistry) RegisterAnchor(anchor *Anchor) (Anchor, error) {
	ret := Anchor{}
	var err error

	ethRegistryContract, _ := getAnchorContract()
	opts, err := ethereum.GetGethTxOpts()
	if err != nil {
		return ret, err
	}

	var bMerkleRoot, bAnchorId [32]byte

	copy(bMerkleRoot[:], anchor.rootHash[:32])
	copy(bAnchorId[:], anchor.anchorID[:32])
	var schemaVersion = big.NewInt(int64(anchor.schemaVersion))

	tx, err := ethRegistryContract.RegisterAnchor(opts, bAnchorId, bMerkleRoot, schemaVersion)
	//tx, err := contract.WitnessDocument(opts, wes.doc.Identifier, wes.doc.WitnessRoot)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Fatalf("Failed to register the anchor [id:%v, hash:%v, schemaVersion:%v] on registry: %v", bAnchorId, bMerkleRoot, schemaVersion, wError)
		//log.Fatalf(wError.(*errors.Error).ErrorStack())
		return ret, err
	}
	log.Printf("Transfer pending: 0x%x\n", tx.Hash())

	waitForIt := make(chan bool, 1)
	go waitForTransaction(tx, waitForIt)
	<-waitForIt

	ret.anchorID = anchor.anchorID
	ret.rootHash = anchor.rootHash
	ret.schemaVersion = anchor.schemaVersion
	return ret, nil
}

func waitForTransaction(pendingTransaction *types.Transaction, conf chan bool) {
	client := ethereum.GetConnection().GetClient()

	ticker := time.NewTicker(500 * time.Millisecond)

	for range ticker.C {
		receipt, err := client.TransactionReceipt(context.TODO(), pendingTransaction.Hash())
		if err != nil {
			if err == goEthereum.NotFound {
				//the transaction has not yet been mined
			} else {
				log.Fatalf("Failed to get transaction status for transaction [%v] on registry: %v", pendingTransaction.Hash().String(), err)
			}

		}
		if receipt != nil {
			break
		}
	}
	fmt.Printf("Transaction completed: 0x%x\n", pendingTransaction.Hash())
	conf <- true
}

func (ethRegistry *EthereumAnchorRegistry) RegisterAsAnchor(anchorID string, rootHash string) (Anchor, error) {
	ret, err := ethRegistry.RegisterAnchor(&Anchor{anchorID, rootHash, ANCHOR_SCHEMA_VERSION})
	return ret, err
}

func getAnchorContract() (anchorContract *EthereumAnchorRegistryContract, err error) {
	// Instantiate the contract and display its name
	client := ethereum.GetConnection()

	anchorContract, err = NewEthereumAnchorRegistryContract(common.HexToAddress("0x6e54f75413ddaf374188fd16b8e0888fee76715e"), client.GetClient())
	if err != nil {
		log.Fatalf("Failed to instantiate the witness contract contract: %v", err)
	}
	return
}
