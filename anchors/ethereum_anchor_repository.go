package anchors

import (
	"context"
	"math/big"

	"time"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
)

type anchorRepositoryContract interface {
	PreCommit(opts *bind.TransactOpts, anchorID *big.Int, signingRoot [32]byte, centID *big.Int, signature []byte, expirationBlock *big.Int) (*types.Transaction, error)
	Commit(opts *bind.TransactOpts, anchorID *big.Int, documentRoot [32]byte, centID *big.Int, documentProofs [][32]byte, signatures []byte) (*types.Transaction, error)
	Commits(opts *bind.CallOpts, anchorID *big.Int) (docRoot [32]byte, err error)
}

type watchAnchorPreCommitted interface {
	//event name: AnchorPreCommitted
	WatchAnchorPreCommitted(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRepositoryContractAnchorPreCommitted,
		from []common.Address, anchorID []*big.Int) (event.Subscription, error)
}

type ethereumAnchorRepository struct {
	config                   Config
	anchorRepositoryContract anchorRepositoryContract
	gethClientFinder         func() ethereum.Client
	queue                    *queue.Server
}

func newEthereumAnchorRepository(config Config, anchorRepositoryContract anchorRepositoryContract, queue *queue.Server, gethClientFinder func() ethereum.Client) AnchorRepository {
	return &ethereumAnchorRepository{config: config, anchorRepositoryContract: anchorRepositoryContract, gethClientFinder: gethClientFinder, queue: queue}
}

// GetDocumentRootOf takes an anchorID and returns the corresponding documentRoot from the chain.
func (ethRepository *ethereumAnchorRepository) GetDocumentRootOf(anchorID AnchorID) (docRoot DocumentRoot, err error) {
	// Ignoring cancelFunc as code will block until response or timeout is triggered
	opts, _ := ethRepository.gethClientFinder().GetGethCallOpts()
	return ethRepository.anchorRepositoryContract.Commits(opts, anchorID.BigInt())
}

// PreCommitAnchor will call the transaction PreCommit on the smart contract
func (ethRepository *ethereumAnchorRepository) PreCommitAnchor(anchorID AnchorID, signingRoot DocumentRoot, centID identity.CentID, signature []byte, expirationBlock *big.Int) (confirmations <-chan *WatchPreCommit, err error) {
	ethRepositoryContract := ethRepository.anchorRepositoryContract
	opts, err := ethereum.GetClient().GetTxOpts(ethRepository.config.GetEthereumDefaultAccountName())
	if err != nil {
		return confirmations, err
	}

	preCommitData := newPreCommitData(anchorID, signingRoot, centID, signature, expirationBlock)
	if err != nil {
		return confirmations, err
	}

	err = sendPreCommitTransaction(ethRepositoryContract, opts, preCommitData)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to send Ethereum pre-commit transaction [id: %x, signingRoot: %x, SchemaVersion:%v]: %v",
			preCommitData.AnchorID, preCommitData.SigningRoot, preCommitData.SchemaVersion, wError)
		return confirmations, err
	}

	return confirmations, err
}

// CommitAnchor will send a commit transaction to Ethereum.
func (ethRepository *ethereumAnchorRepository) CommitAnchor(anchorID AnchorID, documentRoot DocumentRoot, centID identity.CentID, documentProofs [][32]byte, signature []byte) (confirmations <-chan *WatchCommit, err error) {
	conn := ethereum.GetClient()
	opts, err := conn.GetTxOpts(ethRepository.config.GetEthereumDefaultAccountName())
	if err != nil {
		return nil, err
	}

	h, err := conn.GetEthClient().HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	cd := NewCommitData(h.Number.Uint64(), anchorID, documentRoot, centID, documentProofs, signature)
	confirmations, err = ethRepository.setUpCommitEventListener(ethRepository.config.GetEthereumContextWaitTimeout(), opts.From, cd)

	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to set up event listener for commit transaction [id: %x, hash: %x]: %v",
			cd.AnchorID, cd.DocumentRoot, wError)
		return
	}

	err = sendCommitTransaction(ethRepository.anchorRepositoryContract, opts, cd)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to send Ethereum commit transaction[id: %x, hash: %x, SchemaVersion:%v]: %v",
			cd.AnchorID, cd.DocumentRoot, cd.SchemaVersion, wError)
		return
	}
	return confirmations, err
}

// sendPreCommitTransaction sends the actual transaction to the ethereum node.
func sendPreCommitTransaction(contract anchorRepositoryContract, opts *bind.TransactOpts, preCommitData *PreCommitData) error {

	//preparation of data in specific types for the call to Ethereum
	schemaVersion := big.NewInt(int64(preCommitData.SchemaVersion))

	tx, err := ethereum.GetClient().SubmitTransactionWithRetries(contract.PreCommit, opts, preCommitData.AnchorID, preCommitData.SigningRoot,
		preCommitData.CentrifugeID, preCommitData.Signature, preCommitData.ExpirationBlock, schemaVersion)

	if err != nil {
		return err
	}

	log.Infof("Sent off transaction pre-commit [id: %x, hash: %x, SchemaVersion:%v] to registry. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", preCommitData.AnchorID,
		preCommitData.SigningRoot, schemaVersion, tx.Hash(), tx.Nonce(), tx.CheckNonce())

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
	return nil
}

// sendCommitTransaction sends the actual transaction to register the Anchor on Ethereum registry contract
func sendCommitTransaction(contract anchorRepositoryContract, opts *bind.TransactOpts, commitData *CommitData) error {
	tx, err := ethereum.GetClient().SubmitTransactionWithRetries(contract.Commit, opts, commitData.AnchorID.BigInt(), commitData.DocumentRoot,
		commitData.CentrifugeID.BigInt(), commitData.DocumentProofs, commitData.Signature)

	if err != nil {
		return err
	}

	log.Infof("Sent off the anchor [id: %x, hash: %x] to registry. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", commitData.AnchorID,
		commitData.DocumentRoot, tx.Hash(), tx.Nonce(), tx.CheckNonce())
	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
	return nil
}

// TODO: This method is only used by setUpPreCommitEventListener below, it will be changed soon so we can remove the hardcoded `time.Second` and use the global one
func generateEventContext() (*bind.WatchOpts, context.CancelFunc) {
	//listen to this particular anchor being mined/event is triggered
	ctx, cancelFunc := ethereum.DefaultWaitForTransactionMiningContext(time.Second)
	watchOpts := &bind.WatchOpts{Context: ctx}

	return watchOpts, cancelFunc

}

// setUpPreCommitEventListener sets up the listened for the "PreCommit" event to notify the upstream code
// about successful mining/creation of a pre-commit.
func setUpPreCommitEventListener(contractEvent watchAnchorPreCommitted, from common.Address, preCommitData *PreCommitData) (confirmations chan *WatchPreCommit, err error) {
	watchOpts, cancelFunc := generateEventContext()

	//there should always be only one notification coming for this
	//single anchor being registered
	anchorPreCommittedEvents := make(chan *EthereumAnchorRepositoryContractAnchorPreCommitted)
	confirmations = make(chan *WatchPreCommit)
	go waitAndRoutePreCommitEvent(watchOpts.Context, anchorPreCommittedEvents, confirmations, preCommitData)

	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = contractEvent.WatchAnchorPreCommitted(watchOpts, anchorPreCommittedEvents, []common.Address{from}, []*big.Int{preCommitData.AnchorID.BigInt()})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for anchor registration", 1)
		log.Errorf("Failed to watch anchor registered event: %v", wError.Error())
		cancelFunc() // cancel the event router
		return confirmations, wError
	}
	return confirmations, nil
}

// setUpCommitEventListener sets up the listened for the "AnchorCommitted" event to notify the upstream code
// about successful mining/creation of a commit
func (ethRepository *ethereumAnchorRepository) setUpCommitEventListener(timeout time.Duration, from common.Address, commitData *CommitData) (confirmations chan *WatchCommit, err error) {
	confirmations = make(chan *WatchCommit)
	asyncRes, err := ethRepository.queue.EnqueueJob(anchorRepositoryConfirmationTaskName, map[string]interface{}{
		anchorIDParam: commitData.AnchorID,
		addressParam:  from,
		centIDParam:   commitData.CentrifugeID,
		blockHeight:   commitData.BlockHeight,
	})
	if err != nil {
		return nil, err
	}

	go waitAndRouteCommitEvent(timeout, asyncRes, confirmations, commitData)
	return confirmations, nil
}

// waitAndRoutePreCommitEvent notifies the confirmations channel whenever a pre-commit is being noted as Ethereum event
func waitAndRoutePreCommitEvent(ctx context.Context, conf <-chan *EthereumAnchorRepositoryContractAnchorPreCommitted, confirmations chan<- *WatchPreCommit, preCommitData *PreCommitData) {
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving AnchorPreCommitted event for anchor ID: %x, DocumentRoot: %x\n", ctx, preCommitData.AnchorID, preCommitData.SigningRoot)
			confirmations <- &WatchPreCommit{preCommitData, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received AnchorPreCommitted event from: %x\n", res.From)
			confirmations <- &WatchPreCommit{preCommitData, nil}
			return
		}
	}
}

// waitAndRouteCommitEvent notifies the confirmations channel whenever a commit is being noted as Ethereum event
func waitAndRouteCommitEvent(timeout time.Duration, asyncResult queue.TaskResult, confirmations chan<- *WatchCommit, commitData *CommitData) {
	_, err := asyncResult.Get(timeout)
	confirmations <- &WatchCommit{commitData, err}
}
