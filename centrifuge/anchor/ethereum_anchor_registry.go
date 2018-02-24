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
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/core/types"
)

//Supported anchor schema version as stored on public registry
const ANCHOR_SCHEMA_VERSION uint = 1

type EthereumAnchorRegistry struct {
}

func SupportedSchemaVersion() (uint) {
	return ANCHOR_SCHEMA_VERSION
}

type WatchAnchorRegistered interface {
	WatchAnchorRegistered(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRegistryContractAnchorRegistered, from []common.Address, identifier [][32]byte, rootHash [][32]byte) (event.Subscription, error)
}

type RegisterAnchor interface {
	RegisterAnchor(opts *bind.TransactOpts, identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error)
}

func (ethRegistry *EthereumAnchorRegistry) RegisterAsAnchor(anchorID string, rootHash string, confirmations chan<- *Anchor) (error) {
	var err error

	ethRegistryContract, _ := getAnchorContract()
	opts, err := ethereum.GetGethTxOpts()
	if err != nil {
		return err
	}
	registerThisAnchor := generateAnchor(anchorID, rootHash)

	err = setUpRegistrationEventListener(ethRegistryContract, registerThisAnchor, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Printf("Failed to set up event listener for anchor [id: %x, hash: %x, SchemaVersion:%v]: %v", registerThisAnchor.AnchorID, registerThisAnchor.RootHash, registerThisAnchor.SchemaVersion, wError)
		return err
	}

	err = sendRegistrationTransaction(ethRegistryContract, opts, registerThisAnchor)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Printf("Failed to send Ethereum transaction to register anchor [id: %x, hash: %x, SchemaVersion:%v]: %v", registerThisAnchor.AnchorID, registerThisAnchor.RootHash, registerThisAnchor.SchemaVersion, wError)
		return err
	}

	return nil
}

// Convenience method to create a "registerable" `Anchor` from anchor ID and root hash
func generateAnchor(anchorID string, rootHash string) (*Anchor) {
	returnAnchor := &Anchor{}
	returnAnchor.AnchorID = anchorID
	returnAnchor.RootHash = rootHash
	// Rather using SchemaVersion as that's the real value that was passed around instead of calling `SupportedSchemaVersion`
	// again.
	returnAnchor.SchemaVersion = SupportedSchemaVersion()
	return returnAnchor
}

// Sends the actual transaction to register the Anchor on Ethereum registry contract
func sendRegistrationTransaction(ethRegistryContract RegisterAnchor, opts *bind.TransactOpts, anchorToBeRegistered *Anchor) (err error) {
	if len(anchorToBeRegistered.AnchorID) != 32 {
		return errors.New("AnchorID needs to be length of 32")
	}
	if len(anchorToBeRegistered.RootHash) != 32 {
		return errors.New("RootHash needs to be length of 32")
	}

	//preparation of data in specific types for the call to Ethereum
	var bMerkleRoot, bAnchorId [32]byte
	copy(bMerkleRoot[:], anchorToBeRegistered.RootHash[:32])
	copy(bAnchorId[:], anchorToBeRegistered.AnchorID[:32])
	schemaVersion := big.NewInt(int64(anchorToBeRegistered.SchemaVersion))

	tx, err := ethRegistryContract.RegisterAnchor(opts, bAnchorId, bMerkleRoot, schemaVersion)

	if err != nil {
		log.Printf("Failed to send anchor for registration [id: %x, hash: %x, SchemaVersion:%v] on registry: %v", bAnchorId, bMerkleRoot, schemaVersion, err)
		return err
	} else {
		log.Printf("Sent off the anchor [id: %x, hash: %x, SchemaVersion:%v] to registry. Ethereum transaction hash [%x]", bAnchorId, bMerkleRoot, schemaVersion, tx.Hash())
	}

	log.Printf("Transfer pending: 0x%x\n", tx.Hash())
	return
}

// Setting up the listened for the "AnchorRegistered" event to notify the upstream code about successful mining/creation
// of the anchor.
func setUpRegistrationEventListener(ethRegistryContract WatchAnchorRegistered, anchorToBeRegistered *Anchor, confirmations chan<- *Anchor) (err error) {

	//listen to this particular anchor being mined/event is triggered
	watchOpts := &bind.WatchOpts{}
	watchOpts.Context = ethereum.DefaultWaitForTransactionMiningContext()

	//only setting up a channel of 1 notification as there should always be only one notification coming for this
	//single anchor being registered
	anchorRegisteredEvents := make(chan *EthereumAnchorRegistryContractAnchorRegistered, 1)
	go waitAndRouteAnchorRegistrationEvent(anchorRegisteredEvents, watchOpts.Context, confirmations, anchorToBeRegistered)

	var bAnchorId [32]byte
	copy(bAnchorId[:], anchorToBeRegistered.AnchorID[:32])

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = ethRegistryContract.WatchAnchorRegistered(watchOpts, anchorRegisteredEvents, nil, [][32]byte{bAnchorId}, nil)
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for anchor registration", 1)
		log.Printf(wError.Error())
		panic(wError)
	}
	return
}

// Whenever the anchor registration is being noted as Ethereum event, notify the confirmations channel
func waitAndRouteAnchorRegistrationEvent(conf <-chan *EthereumAnchorRegistryContractAnchorRegistered, ctx context.Context, confirmations chan<- *Anchor, pushThisAnchor *Anchor) {
	for {
		select {
		case <-ctx.Done():
			log.Fatalf("Context closed before receiving AnchorRegistered event for anchor ID: %v, RootHash: %v", pushThisAnchor.AnchorID, pushThisAnchor.RootHash)
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
