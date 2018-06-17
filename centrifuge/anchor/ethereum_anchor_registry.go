package anchor

import (
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
	"math/big"
)

//Supported anchor schema version as stored on public registry
const ANCHOR_SCHEMA_VERSION uint = 1

type EthereumAnchorRegistry struct {
}

func SupportedSchemaVersion() uint {
	return ANCHOR_SCHEMA_VERSION
}

type WatchAnchorRegistered interface {
	WatchAnchorRegistered(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRegistryContractAnchorRegistered, from []common.Address, identifier [][32]byte, rootHash [][32]byte) (event.Subscription, error)
}

type RegisterAnchor interface {
	RegisterAnchor(opts *bind.TransactOpts, identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error)
}

// RegisterAsAnchor registers the given anchorID and rootHash on the Ethereum anchor registry and submits the confirmed Anchor
// into the confirmations channel when done.
// Could error out with Fatal error in case the confirmation is never received within the timeframe of configured value
// of `ethereum.contextWaitTimeout`.
func (ethRegistry *EthereumAnchorRegistry) RegisterAsAnchor(anchorID string, rootHash string, confirmations chan<- *WatchAnchor) error {
	var err error

	ethRegistryContract, _ := getAnchorContract()
	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return err
	}
	registerThisAnchor, err := generateAnchor(anchorID, rootHash)
	if err != nil {
		return err
	}

	err = setUpRegistrationEventListener(ethRegistryContract, opts.From, registerThisAnchor, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to set up event listener for anchor [id: %x, hash: %x, SchemaVersion:%v]: %v", registerThisAnchor.AnchorID, registerThisAnchor.RootHash, registerThisAnchor.SchemaVersion, wError)
		return err
	}

	err = sendRegistrationTransaction(ethRegistryContract, opts, registerThisAnchor)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to send Ethereum transaction to register anchor [id: %x, hash: %x, SchemaVersion:%v]: %v", registerThisAnchor.AnchorID, registerThisAnchor.RootHash, registerThisAnchor.SchemaVersion, wError)
		return err
	}

	return nil
}

// generateAnchor is a convenience method to create a "registerable" `Anchor` from anchor ID and root hash
func generateAnchor(anchorID string, rootHash string) (returnAnchor *Anchor, err error) {
	err = tools.CheckLen32(anchorID, "anchorID needs to be length of 32. Got value [%v]")
	if err != nil {
		return nil, err
	}
	err = tools.CheckLen32(rootHash, "rootHash needs to be length of 32. Got value [%v]")
	if err != nil {
		return nil, err
	}

	returnAnchor = &Anchor{}
	returnAnchor.AnchorID = anchorID
	returnAnchor.RootHash = rootHash
	// Rather using SchemaVersion as that's the real value that was passed around instead of calling `SupportedSchemaVersion`
	// again.
	returnAnchor.SchemaVersion = SupportedSchemaVersion()
	return returnAnchor, nil
}

// sendRegistrationTransaction sends the actual transaction to register the Anchor on Ethereum registry contract
func sendRegistrationTransaction(ethRegistryContract RegisterAnchor, opts *bind.TransactOpts, anchorToBeRegistered *Anchor) (err error) {
	err = tools.CheckLen32(anchorToBeRegistered.AnchorID, "AnchorID needs to be length of 32. Got value [%x]")
	if err != nil {
		return err
	}
	err = tools.CheckLen32(anchorToBeRegistered.RootHash, "RootHash needs to be length of 32. Got value [%x]")
	if err != nil {
		return err
	}

	//preparation of data in specific types for the call to Ethereum
	var bMerkleRoot, bAnchorId [32]byte
	copy(bMerkleRoot[:], anchorToBeRegistered.RootHash[:32])
	copy(bAnchorId[:], anchorToBeRegistered.AnchorID[:32])
	schemaVersion := big.NewInt(int64(anchorToBeRegistered.SchemaVersion))

	tx, err := ethereum.SubmitTransactionWithRetries(ethRegistryContract.RegisterAnchor, opts, bAnchorId, bMerkleRoot, schemaVersion)

	if err != nil {
		log.Errorf("Failed to send anchor for registration [id: %x, hash: %x, SchemaVersion:%v] on registry: %v", bAnchorId, bMerkleRoot, schemaVersion, err)
		return err
	} else {
		log.Infof("Sent off the anchor [id: %x, hash: %x, SchemaVersion:%v] to registry. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", bAnchorId, bMerkleRoot, schemaVersion, tx.Hash(), tx.Nonce(), tx.CheckNonce())
	}

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
	return
}

// setUpRegistrationEventListener sets up the listened for the "AnchorRegistered" event to notify the upstream code about successful mining/creation
// of the anchor.
func setUpRegistrationEventListener(ethRegistryContract WatchAnchorRegistered, from common.Address, anchorToBeRegistered *Anchor, confirmations chan<- *WatchAnchor) (err error) {

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
	_, err = ethRegistryContract.WatchAnchorRegistered(watchOpts, anchorRegisteredEvents, []common.Address{from}, [][32]byte{bAnchorId}, nil)
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for anchor registration", 1)
		log.Panicf(wError.Error())
	}
	return
}

// waitAndRouteAnchorRegistrationEvent notififies the confirmations channel whenever the anchor registration is being noted as Ethereum event
func waitAndRouteAnchorRegistrationEvent(conf <-chan *EthereumAnchorRegistryContractAnchorRegistered, ctx context.Context, confirmations chan<- *WatchAnchor, pushThisAnchor *Anchor) {
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving AnchorRegistered event for anchor ID: %x, RootHash: %x\n", ctx, pushThisAnchor.AnchorID, pushThisAnchor.RootHash)
			confirmations <- &WatchAnchor{pushThisAnchor, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received AnchorRegistered event from: %x, identifier: %x\n", res.From, res.Identifier)
			confirmations <- &WatchAnchor{pushThisAnchor, nil}
			return
		}
	}
}

func getAnchorContract() (anchorContract *EthereumAnchorRegistryContract, err error) {
	client := ethereum.GetConnection()
	anchorContract, err = NewEthereumAnchorRegistryContract(config.Config.GetContractAddress("anchorRegistry"), client.GetClient())
	if err != nil {
		log.Fatalf("Failed to instantiate the anchor contract: %v", err)
	}
	return
}
