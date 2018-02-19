package anchor

import (
	//"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	//"github.com/ethereum/go-ethereum/common"
	//"github.com/spf13/viper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"log"
	"math/big"
	//"fmt"
	"github.com/go-errors/errors"
	//"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"fmt"
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

	//listen to this particular anchor being mined/event is triggered
	watchOpts := &bind.WatchOpts{}
	watchOpts.Context = ethereum.DefaultWaitForTransactionMiningContext()

	sinkNo := make(chan *EthereumAnchorRegistryContractAnchorRegistered, 100)
	go waitForTransaction(sinkNo, watchOpts.Context)

	_, err = ethRegistryContract.WatchAnchorRegistered(watchOpts, sinkNo,nil,([][32]byte{bAnchorId}),nil)
	if err != nil{
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for anchor registration: %v",1)
		log.Fatalf(wError.Error())
		panic(wError)
	}

	tx, err := ethRegistryContract.RegisterAnchor(opts, bAnchorId, bMerkleRoot, schemaVersion)

	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Fatalf("Failed to register the anchor [id: %x, hash: %x, schemaVersion:%v] on registry: %v", bAnchorId, bMerkleRoot, schemaVersion, wError)
		//log.Fatalf(wError.(*errors.Error).ErrorStack())
		return ret, err
	} else {
		log.Printf("Sent off the anchor [id: %x, hash: %x, schemaVersion:%v] to registry. Ethereum transaction hash [%x]", bAnchorId, bMerkleRoot, schemaVersion, tx.Hash())
	}
	log.Printf("Transfer pending: 0x%x\n", tx.Hash())

	select {

	case <-watchOpts.Context.Done():
		fmt.Println("waited long enough")
	}

	ret.anchorID = anchor.anchorID
	ret.rootHash = anchor.rootHash
	ret.schemaVersion = anchor.schemaVersion
	return ret, nil
}

func waitForTransaction(conf <-chan *EthereumAnchorRegistryContractAnchorRegistered, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return // returning not to leak the goroutine
		case res := <-conf:
			log.Printf("Received log output from: %x, identifier: %x",res.From,res.Identifier)
		}
	}
}

func (ethRegistry *EthereumAnchorRegistry) RegisterAsAnchor(anchorID string, rootHash string) (Anchor, error) {
	ret, err := ethRegistry.RegisterAnchor(&Anchor{anchorID, rootHash, ANCHOR_SCHEMA_VERSION})
	return ret, err
}

func getAnchorContract() (anchorContract *EthereumAnchorRegistryContract, err error) {
	// Instantiate the contract and display its name
	client := ethereum.GetConnection()

	anchorContract, err = NewEthereumAnchorRegistryContract(common.HexToAddress("0x995ef27e64cb9ef07eb6f9d255a3951ef20416fd"), client.GetClient())
	if err != nil {
		log.Fatalf("Failed to instantiate the witness contract contract: %v", err)
	}
	return
}
