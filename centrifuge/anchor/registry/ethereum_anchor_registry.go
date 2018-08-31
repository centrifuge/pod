package registry

import (
	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
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
func (ethRegistry *EthereumAnchorRegistry) RegisterAsAnchor(anchorID [32]byte, rootHash [32]byte) (confirmations <-chan *WatchAnchor, err error) {
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

	confirmations, err = setUpRegistrationEventListener(opts.From, registerThisAnchor)
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

	//preparation of data in specific types for the   call to Ethereum
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
func setUpRegistrationEventListener(from common.Address, anchorToBeRegistered *Anchor) (confirmations chan *WatchAnchor, err error) {
	confirmations = make(chan *WatchAnchor)
	asyncRes, err := queue.Queue.DelayKwargs(AnchoringConfirmationTaskName, map[string]interface{}{
		AnchorIdParam: anchorToBeRegistered.AnchorID,
		AddressParam:  from,
	})
	if err != nil {
		return nil, err
	}
	go waitAndRouteAnchorRegistrationEvent(asyncRes, confirmations, anchorToBeRegistered)
	return confirmations, nil
}

// waitAndRouteAnchorRegistrationEvent notifies the confirmations channel whenever the anchor registration is being noted as Ethereum event
func waitAndRouteAnchorRegistrationEvent(asyncResult *gocelery.AsyncResult, confirmations chan<- *WatchAnchor, pushThisAnchor *Anchor) {
	_, err := asyncResult.Get(ethereum.GetDefaultContextTimeout())
	confirmations <- &WatchAnchor{pushThisAnchor, err}
}

func getAnchorContract() (anchorContract *EthereumAnchorRegistryContract, err error) {
	client := ethereum.GetConnection()
	anchorContract, err = NewEthereumAnchorRegistryContract(config.Config.GetContractAddress("anchorRegistry"), client.GetClient())
	if err != nil {
		log.Fatalf("Failed to instantiate the anchor contract: %v", err)
	}
	return
}
