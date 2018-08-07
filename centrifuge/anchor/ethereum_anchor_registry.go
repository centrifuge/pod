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
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
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
func (ethRegistry *EthereumAnchorRegistry) RegisterAsAnchor(anchorID [32]byte, rootHash [32]byte, confirmations chan<- *WatchAnchor) (err error) {
	if tools.IsEmptyByte32(anchorID) {
		err = errors.New("Can not work with empty anchor ID")
		return
	}
	if tools.IsEmptyByte32(rootHash) {
		err = errors.New("Can not work with empty root hash")
		return
	}

	ethRegistryContract, _ := getAnchorContract()
	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return
	}
	registerThisAnchor, err := generateAnchor(anchorID, rootHash)
	if err != nil {
		return
	}

	err = setUpRegistrationEventListener(ethRegistryContract, opts.From, registerThisAnchor, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to set up event listener for anchor [id: %x, hash: %x, SchemaVersion:%v]: %v", registerThisAnchor.AnchorID, registerThisAnchor.RootHash, registerThisAnchor.SchemaVersion, wError)
		return
	}

	err = sendRegistrationTransaction(ethRegistryContract, opts, registerThisAnchor)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to send Ethereum transaction to register anchor [id: %x, hash: %x, SchemaVersion:%v]: %v", registerThisAnchor.AnchorID, registerThisAnchor.RootHash, registerThisAnchor.SchemaVersion, wError)
		return
	}
	return
}

// generateAnchor is a convenience method to create a "registerable" `Anchor` from anchor ID and root hash
func generateAnchor(anchorID [32]byte, rootHash [32]byte) (returnAnchor *Anchor, err error) {
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

	//preparation of data in specific types for the call to Ethereum
	schemaVersion := big.NewInt(int64(anchorToBeRegistered.SchemaVersion))

	tx, err := ethereum.SubmitTransactionWithRetries(ethRegistryContract.RegisterAnchor, opts, anchorToBeRegistered.AnchorID, anchorToBeRegistered.RootHash, schemaVersion)

	if err != nil {
		log.Errorf("Failed to send anchor for registration [id: %x, hash: %x, SchemaVersion:%v] on registry: %v", anchorToBeRegistered.AnchorID, anchorToBeRegistered.RootHash, schemaVersion, err)
		return err
	} else {
		log.Infof("Sent off the anchor [id: %x, hash: %x, SchemaVersion:%v] to registry. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", anchorToBeRegistered.AnchorID, anchorToBeRegistered.RootHash, schemaVersion, tx.Hash(), tx.Nonce(), tx.CheckNonce())
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
	cancel := make(chan interface{})
	go waitAndRouteAnchorRegistrationEvent(anchorRegisteredEvents, watchOpts.Context, confirmations, anchorToBeRegistered, cancel)

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = ethRegistryContract.WatchAnchorRegistered(watchOpts, anchorRegisteredEvents, []common.Address{from}, [][32]byte{anchorToBeRegistered.AnchorID}, nil)
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for anchor registration", 1)
		log.Errorf("Failed to watch anchor registered event: %v", wError.Error())
		// stop the awaiting go routine
		utils.SendNonBlocking(true, cancel)
		return wError
	}
	return
}

// waitAndRouteAnchorRegistrationEvent notifies the confirmations channel whenever the anchor registration is being noted as Ethereum event
func waitAndRouteAnchorRegistrationEvent(conf <-chan *EthereumAnchorRegistryContractAnchorRegistered,
	ctx context.Context, confirmations chan<- *WatchAnchor, pushThisAnchor *Anchor, cancel <-chan interface{}) {
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
		case <-cancel:
			log.Info("Anchor registration event handling was cancelled by upstream")
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
