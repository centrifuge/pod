package repository

import (
	"context"
	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
)

type EthereumAnchorRepository struct {
}

type AnchorRepositoryContract interface {

	//transactions
	PreCommit(opts *bind.TransactOpts, anchorId *big.Int, signingRoot [32]byte, centrifugeId *big.Int, signature []byte, expirationBlock *big.Int) (*types.Transaction, error)
	Commit(opts *bind.TransactOpts, _anchorId *big.Int, _documentRoot [32]byte, _centrifugeId *big.Int, _documentProofs [][32]byte, _signature []byte) (*types.Transaction, error)

	//event name: AnchorPreCommitted
	WatchAnchorPreCommitted(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRepositoryContractAnchorPreCommitted,
		from []common.Address, anchorId []*big.Int) (event.Subscription, error)

	//event name: AnchorCommitted
	WatchAnchorCommitted(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRepositoryContractAnchorCommitted,
		from []common.Address, anchorId []*big.Int, centrifugeId []*big.Int) (event.Subscription, error)
}

func (ethRepository *EthereumAnchorRepository) PreCommitAnchor(anchorId *big.Int, signingRoot [32]byte, centrifugeId *big.Int, signature []byte, expirationBlock *big.Int) (confirmations <-chan *WatchPreCommit, err error) {

	//TODO check if parameters are valid
	ethRepositoryContract, _ := getRepositoryContract()
	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return
	}
	preCommitData, err := generatePreCommitData(anchorId, signingRoot, centrifugeId, signature, expirationBlock)
	if err != nil {
		return
	}

	confirmations, err = setUpPreCommitEventListener(ethRepositoryContract, opts.From, preCommitData)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to set up event listener for anchor [id: %x, hash: %x, SchemaVersion:%v]: %v",
			preCommitData.anchorId, preCommitData.signingRoot, preCommitData.SchemaVersion, wError)
		return
	}

	err = sendPreCommitTransaction(ethRepositoryContract, opts, preCommitData)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to send Ethereum transaction to register anchor [id: %x, hash: %x, SchemaVersion:%v]: %v",
			preCommitData.anchorId, preCommitData.signingRoot, preCommitData.SchemaVersion, wError)
		return
	}
	return confirmations, err
}

// sendPreCommitTransaction sends the actual transaction to register the Anchor on Ethereum registry contract
func sendPreCommitTransaction(contract AnchorRepositoryContract, opts *bind.TransactOpts, preCommitData *PreCommitData) (err error) {

	//preparation of data in specific types for the call to Ethereum
	schemaVersion := big.NewInt(int64(preCommitData.SchemaVersion))

	tx, err := ethereum.SubmitTransactionWithRetries(contract.PreCommit, opts, preCommitData.anchorId, preCommitData.signingRoot,
		preCommitData.centrifugeId, preCommitData.signature, preCommitData.expirationBlock, schemaVersion)

	if err != nil {
		log.Errorf("Failed to pre commit anchor[id: %x, hash: %x, SchemaVersion:%v] on registry: %v", preCommitData.anchorId, preCommitData.signingRoot, schemaVersion, err)
		return err
	} else {
		log.Infof("Sent off the anchor [id: %x, hash: %x, SchemaVersion:%v] to registry. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", preCommitData.anchorId,
			preCommitData.signingRoot, schemaVersion, tx.Hash(), tx.Nonce(), tx.CheckNonce())
	}

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
	return
}

// sendPreCommitTransaction sends the actual transaction to register the Anchor on Ethereum registry contract
func sendCommitTransaction(contract AnchorRepositoryContract, opts *bind.TransactOpts, commitData *CommitData) (err error) {

	tx, err := ethereum.SubmitTransactionWithRetries(contract.Commit, opts, commitData.anchorId, commitData.documentRoot,
		commitData.centrifugeId, commitData.documentProofs, commitData.signature)

	if err != nil {
		log.Errorf("Failed to pre commit anchor[id: %x, hash: %x", commitData.anchorId, commitData.documentRoot, err)
		return err
	} else {
		log.Infof("Sent off the anchor [id: %x, hash: %x] to registry. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", commitData.anchorId,
			commitData.documentRoot, tx.Hash(), tx.Nonce(), tx.CheckNonce())
	}

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
	return
}

func (ethRepository *EthereumAnchorRepository) CommitAnchor(anchorId *big.Int, documentRoot [32]byte, centrifugeId *big.Int, documentProofs [][32]byte, signature []byte) (confirmations <-chan *WatchCommit, err error) {

	ethRepositoryContract, _ := getRepositoryContract()
	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return
	}
	//TODO check if parameters are valid
	commitData, err := generateCommitData(anchorId, documentRoot, centrifugeId, documentProofs, signature)
	if err != nil {
		return
	}

	confirmations, err = setUpCommitEventListener(ethRepositoryContract, opts.From, commitData)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to set up event listener for anchor [id: %x, hash: %x, SchemaVersion:%v]: %v",
			commitData.anchorId, commitData.documentRoot, commitData.SchemaVersion, wError)
		return
	}

	err = sendCommitTransaction(ethRepositoryContract, opts, commitData)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to send Ethereum transaction to register anchor [id: %x, hash: %x, SchemaVersion:%v]: %v",
			commitData.anchorId, commitData.documentRoot, commitData.SchemaVersion, wError)
		return
	}
	return confirmations, err

	return nil, nil
}

func generateEventContext() (*bind.WatchOpts, context.CancelFunc) {
	//listen to this particular anchor being mined/event is triggered
	ctx, cancelFunc := ethereum.DefaultWaitForTransactionMiningContext()
	watchOpts := &bind.WatchOpts{Context: ctx}

	return watchOpts, cancelFunc

}

// setUpPreCommitEventListener sets up the listened for the "PreCommit" event to notify the upstream code about successful mining/creation
// of a pre-commit.
func setUpPreCommitEventListener(contract AnchorRepositoryContract, from common.Address, preCommitData *PreCommitData) (confirmations chan *WatchPreCommit, err error) {

	watchOpts, cancelFunc := generateEventContext()

	//there should always be only one notification coming for this
	//single anchor being registered
	anchorPreCommittedEvents := make(chan *EthereumAnchorRepositoryContractAnchorPreCommitted)
	confirmations = make(chan *WatchPreCommit)
	go waitAndRoutePreCommitEvent(anchorPreCommittedEvents, watchOpts.Context, confirmations, preCommitData)

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = contract.WatchAnchorPreCommitted(watchOpts, anchorPreCommittedEvents, []common.Address{from}, []*big.Int{preCommitData.anchorId})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for anchor registration", 1)
		log.Errorf("Failed to watch anchor registered event: %v", wError.Error())
		cancelFunc() // cancel the event router
		return confirmations, wError
	}
	return confirmations, nil
}

// setUpCommitEventListener sets up the listened for the "AnchorCommitted" event to notify the upstream code about successful mining/creation
// of a commit
func setUpCommitEventListener(contract AnchorRepositoryContract, from common.Address, commitData *CommitData) (confirmations chan *WatchCommit, err error) {

	watchOpts, cancelFunc := generateEventContext()

	//there should always be only one notification coming for this
	//single anchor being committed
	anchorCommittedEvents := make(chan *EthereumAnchorRepositoryContractAnchorCommitted)
	confirmations = make(chan *WatchCommit)
	go waitAndRouteCommitEvent(anchorCommittedEvents, watchOpts.Context, confirmations, commitData)

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = contract.WatchAnchorCommitted(watchOpts, anchorCommittedEvents, []common.Address{from}, []*big.Int{commitData.anchorId}, []*big.Int{commitData.centrifugeId})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for anchor registration", 1)
		log.Errorf("Failed to watch anchor registered event: %v", wError.Error())
		cancelFunc() // cancel the event router
		return confirmations, wError
	}
	return confirmations, nil
}

// waitAndRoutePreCommitEvent notifies the confirmations channel whenever a pre-commit is being noted as Ethereum event
func waitAndRoutePreCommitEvent(conf <-chan *EthereumAnchorRepositoryContractAnchorPreCommitted, ctx context.Context, confirmations chan<- *WatchPreCommit, preCommitData *PreCommitData) {
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving AnchorPreCommitted event for anchor ID: %x, RootHash: %x\n", ctx, preCommitData.anchorId, preCommitData.signingRoot)
			confirmations <- &WatchPreCommit{preCommitData, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received AnchorPreCommitted event from: %x\n", res.From)
			confirmations <- &WatchPreCommit{preCommitData, nil}
			return
		}
	}
}

// waitAndRoutePreCommitEvent notifies the confirmations channel whenever a pre-commit is being noted as Ethereum event
func waitAndRouteCommitEvent(conf <-chan *EthereumAnchorRepositoryContractAnchorCommitted, ctx context.Context, confirmations chan<- *WatchCommit, commitData *CommitData) {
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving AnchorPreCommitted event for anchor ID: %x, RootHash: %x\n", ctx, commitData.anchorId, commitData.documentRoot)
			confirmations <- &WatchCommit{commitData, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received AnchorCommitted event from: %x\n", res.From)
			confirmations <- &WatchCommit{commitData, nil}
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
