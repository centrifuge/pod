package repository

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
)



type EthereumAnchorRepository struct {
}



// TODO check behaviour
type WatchAnchorRegisteredEVENT interface {
	WatchAnchorRegisteredEVENT(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRepositoryContractAnchorCommitted, from []common.Address, identifier [][32]byte, rootHash [][32]byte) (event.Subscription, error)
}

type RegisterAnchor interface {
	RegisterAnchor(opts *bind.TransactOpts, identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error)
}

//
func (ethRepository *EthereumAnchorRepository) PreCommit(anchorId *big.Int, signingRoot [32]byte, centrifugeId *big.Int, signature []byte, expirationBlock *big.Int) (confirmations <-chan *WatchPreCommit, err error) {

	//TODO check parameters
/*
	ethRepositoryContract, _ := getRepositoryContract()
	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return
	}
	registerThisAnchor, err := generateAnchor(anchorID, rootHash)
	if err != nil {
		return
	}

	confirmations, err = setUpPreCommitEventListener(ethRepositoryContract, opts.From, registerThisAnchor)
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
	}*/
	return nil,nil
}

func (ethRepository *EthereumAnchorRepository) Commit(anchorId *big.Int, documentRoot [32]byte, centrifugeId *big.Int, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error) {
	//TODO implement Commit
	return nil,nil
}

/*
// sendRegistrationTransaction sends the actual transaction to register the Anchor on Ethereum registry contract
func setUpPreCommitEventListener(ethRegistryContract RegisterAnchor, opts *bind.TransactOpts, anchorToBeRegistered *Anchor) (err error) {

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

// setUpRegistrationEventListener sets up the listened for the "AnchorRegisteredEVENT" event to notify the upstream code about successful mining/creation
// of the anchor.
func setUpRegistrationEventListener(ethRegistryContract WatchAnchorRegisteredEVENT, from common.Address, anchorToBeRegistered *Anchor) (confirmations chan *WatchAnchor, err error) {
	//listen to this particular anchor being mined/event is triggered
	ctx, cancelFunc := ethereum.DefaultWaitForTransactionMiningContext()
	watchOpts := &bind.WatchOpts{Context: ctx}

	//there should always be only one notification coming for this
	//single anchor being registered
	AnchorRegisteredEVENTEvents := make(chan *EthereumAnchorRegistryContractAnchorRegisteredEVENT)
	confirmations = make(chan *WatchAnchor)
	go waitAndRouteAnchorRegistrationEvent(AnchorRegisteredEVENTEvents, watchOpts.Context, confirmations, anchorToBeRegistered)

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = ethRegistryContract.WatchAnchorRegisteredEVENT(watchOpts, AnchorRegisteredEVENTEvents, []common.Address{from}, [][32]byte{anchorToBeRegistered.AnchorID}, nil)
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for anchor registration", 1)
		log.Errorf("Failed to watch anchor registered event: %v", wError.Error())
		cancelFunc() // cancel the event router
		return confirmations, wError
	}
	return confirmations, nil
}

// waitAndRouteAnchorRegistrationEvent notifies the confirmations channel whenever the anchor registration is being noted as Ethereum event
func waitAndRouteAnchorRegistrationEvent(conf <-chan *EthereumAnchorRegistryContractAnchorRegisteredEVENT, ctx context.Context, confirmations chan<- *WatchAnchor, pushThisAnchor *Anchor) {
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving AnchorRegisteredEVENT event for anchor ID: %x, RootHash: %x\n", ctx, pushThisAnchor.AnchorID, pushThisAnchor.RootHash)
			confirmations <- &WatchAnchor{pushThisAnchor, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received AnchorRegisteredEVENT event from: %x, identifier: %x\n", res.From, res.Identifier)
			confirmations <- &WatchAnchor{pushThisAnchor, nil}
			return
		}
	}
}

func getRepositoryContract() (repositoryContract *EthereumAnchorRepositoryContract, err error) {
	client := ethereum.GetConnection()

	// TODO add parameter anchorRepository to config
	repositoryContract, err = NewEthereumAnchorRepositoryContract(config.Config.GetContractAddress("anchorRegistry"), client.GetClient())
	if err != nil {
		log.Fatalf("Failed to instantiate the anchor contract: %v", err)
	}
	return
}
*/