package anchor

import (
	//"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	//"github.com/ethereum/go-ethereum/common"
	//"github.com/spf13/viper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"log"
	"math/big"
	"github.com/go-errors/errors"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"context"
)

//Supported anchor schema version as stored on public registry
const ANCHOR_SCHEMA_VERSION uint = 1

type EthereumAnchorRegistry struct {
}

func SupportedSchemaVersion() (uint) {
	return ANCHOR_SCHEMA_VERSION
}

func (ethRegistry *EthereumAnchorRegistry) RegisterAsAnchor(anchorID string, rootHash string, confirmations chan<- *Anchor) (error) {
	var err error

	ethRegistryContract, _ := getAnchorContract()
	opts, err := ethereum.GetGethTxOpts()
	if err != nil {
		return err
	}
	registerThisAnchor := ethRegistry.generateAnchor(anchorID, rootHash)

	err = setUpRegistrationEventListener(ethRegistryContract, registerThisAnchor, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Fatalf("Failed to set up event listener for anchor [id: %x, hash: %x, schemaVersion:%v]: %v", registerThisAnchor.anchorID, registerThisAnchor.rootHash, registerThisAnchor.schemaVersion, wError)
		return err
	}

	err = sendRegistrationTransaction(ethRegistryContract, opts, registerThisAnchor)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Fatalf("Failed to send Ethereum transaction to register anchor [id: %x, hash: %x, schemaVersion:%v]: %v", registerThisAnchor.anchorID, registerThisAnchor.rootHash, registerThisAnchor.schemaVersion, wError)
		return err
	}

	return nil
}

// Convenience method to create a "registerable" `Anchor` from anchor ID and root hash
func (ethRegistry *EthereumAnchorRegistry) generateAnchor(anchorID string, rootHash string) (*Anchor) {
	returnAnchor := &Anchor{}
	returnAnchor.anchorID = anchorID
	returnAnchor.rootHash = rootHash
	// Rather using schemaVersion as that's the real value that was passed around instead of calling `SupportedSchemaVersion`
	// again.
	returnAnchor.schemaVersion = SupportedSchemaVersion()
	return returnAnchor
}

// Sends the actual transaction to register the Anchor on Ethereum registry contract
func sendRegistrationTransaction(ethRegistryContract *EthereumAnchorRegistryContract, opts *bind.TransactOpts, anchorToBeRegistered *Anchor) (err error) {
	//preparation of data in specific types for the call to Ethereum
	var bMerkleRoot, bAnchorId [32]byte
	copy(bMerkleRoot[:], anchorToBeRegistered.rootHash[:32])
	copy(bAnchorId[:], anchorToBeRegistered.anchorID[:32])
	schemaVersion := big.NewInt(int64(anchorToBeRegistered.schemaVersion))

	tx, err := ethRegistryContract.RegisterAnchor(opts, bAnchorId, bMerkleRoot, schemaVersion)

	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Fatalf("Failed to send anchor for registration [id: %x, hash: %x, schemaVersion:%v] on registry: %v", bAnchorId, bMerkleRoot, schemaVersion, wError)
		return err
	} else {
		log.Printf("Sent off the anchor [id: %x, hash: %x, schemaVersion:%v] to registry. Ethereum transaction hash [%x]", bAnchorId, bMerkleRoot, schemaVersion, tx.Hash())
	}

	log.Printf("Transfer pending: 0x%x\n", tx.Hash())
	return
}


// Setting up the listened for the "AnchorRegistered" event to notify the upstream code about successful mining/creation
// of the anchor.
func setUpRegistrationEventListener(ethRegistryContract *EthereumAnchorRegistryContract, anchorToBeRegistered *Anchor, confirmations chan<- *Anchor) (err error) {

	//listen to this particular anchor being mined/event is triggered
	watchOpts := &bind.WatchOpts{}
	watchOpts.Context = ethereum.DefaultWaitForTransactionMiningContext()

	//only setting up a channel of 1 notification as there should always be only one notification coming for this
	//single anchor being registered
	anchorRegisteredEvents := make(chan *EthereumAnchorRegistryContractAnchorRegistered, 1)
	go waitAndRouteAnchorRegistrationEvent(anchorRegisteredEvents, watchOpts.Context, confirmations, anchorToBeRegistered)

	var bAnchorId [32]byte
	copy(bAnchorId[:], anchorToBeRegistered.anchorID[:32])

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = ethRegistryContract.WatchAnchorRegistered(watchOpts, anchorRegisteredEvents, nil, [][32]byte{bAnchorId}, nil)
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for anchor registration: %v", 1)
		log.Fatalf(wError.Error())
		panic(wError)
	}
	return
}

// Whenever the anchor registration is being noted as Ethereum event, notify the confirmations channel
func waitAndRouteAnchorRegistrationEvent(conf <-chan *EthereumAnchorRegistryContractAnchorRegistered, ctx context.Context, confirmations chan<- *Anchor, pushThisAnchor *Anchor) {
	for {
		select {
		case <-ctx.Done():
			log.Fatalf("Context closed before receiving AnchorRegistered event for anchor ID: %v, rootHash: %v", pushThisAnchor.anchorID, pushThisAnchor.rootHash)
			return
		case res := <-conf:
			log.Printf("Received AnchorRegistered event from: %x, identifier: %x", res.From, res.Identifier)
			confirmations <- pushThisAnchor
			return
		}
	}
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
