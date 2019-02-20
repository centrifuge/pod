package anchors

import (
	"context"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/satori/go.uuid"
	"math/big"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

type anchorRepositoryContract interface {
	PreCommit(opts *bind.TransactOpts, _anchorId *big.Int, _signingRoot [32]byte, _expirationBlock *big.Int) (*types.Transaction, error)
	Commit(opts *bind.TransactOpts, _anchorId *big.Int, _documentRoot [32]byte, _documentProofs [][32]byte) (*types.Transaction, error)
	Commits(opts *bind.CallOpts, anchorID *big.Int) (docRoot [32]byte, err error)
}





type watchAnchorPreCommitted interface {
	//event name: AnchorPreCommitted
	WatchAnchorPreCommitted(opts *bind.WatchOpts, sink chan<- *AnchorContractAnchorPreCommitted,
		from []common.Address, anchorID []*big.Int) (event.Subscription, error)
}

type service struct {
	config                   Config
	anchorRepositoryContract anchorRepositoryContract
	client         ethereum.Client
	queue                    *queue.Server
	txManager transactions.Manager
}

func newService(config Config, anchorContract anchorRepositoryContract, queue *queue.Server, client ethereum.Client,txManager transactions.Manager) AnchorRepository {
	return &service{config: config, anchorRepositoryContract: anchorContract, client: client, queue: queue, txManager: txManager}
}

// GetDocumentRootOf takes an anchorID and returns the corresponding documentRoot from the chain.
func (s *service) GetDocumentRootOf(anchorID AnchorID) (docRoot DocumentRoot, err error) {
	// Ignoring cancelFunc as code will block until response or timeout is triggered
	opts, _ := s.client.GetGethCallOpts(false)
	return s.anchorRepositoryContract.Commits(opts, anchorID.BigInt())
}

// PreCommitAnchor will call the transaction PreCommit on the smart contract
func (s *service) PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot, centID identity.CentID, signature []byte, expirationBlock *big.Int) (confirmations <-chan *WatchPreCommit, err error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	ethRepositoryContract := s.anchorRepositoryContract
	opts, err := ethereum.GetClient().GetTxOpts(tc.GetEthereumDefaultAccountName())
	if err != nil {
		return confirmations, err
	}

	preCommitData := newPreCommitData(anchorID, signingRoot, centID, signature, expirationBlock)
	if err != nil {
		return confirmations, err
	}

	err = sendPreCommitTransaction(ethRepositoryContract, opts, preCommitData)
	if err != nil {
		wError := errors.New("%v", err)
		log.Errorf("Failed to send Ethereum pre-commit transaction [id: %x, signingRoot: %x, SchemaVersion:%v]: %v",
			preCommitData.AnchorID, preCommitData.SigningRoot, preCommitData.SchemaVersion, wError)
		return confirmations, err
	}

	return confirmations, err
}

// ethereumTX is submitting an Ethereum transaction and starts a task to wait for the transaction result
func (s service) ethereumTX(opts *bind.TransactOpts, contractMethod interface{}, params ...interface{}) func(accountID identity.DID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID identity.DID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {

		ethTX, err := s.client.SubmitTransactionWithRetries(contractMethod, opts, params...)
		if err != nil {
			errOut <- err
			return
		}


		res, err := ethereum.QueueEthTXStatusTask(accountID, txID, ethTX.Hash(), s.queue)
		if err != nil {
			errOut <- err
			return
		}

		_, err = res.Get(txMan.GetDefaultTaskTimeout())
		if err != nil {
			errOut <- err
			return
		}
		errOut <- nil
	}
}

// TODO move func to utils or account
func (s service) getDID(ctx context.Context) (did identity.DID, err error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return did, err
	}

	addressByte, err := tc.GetIdentityID()
	if err != nil {
		return did, err
	}
	did = identity.NewDID(common.BytesToAddress(addressByte))
	return did, nil

}

// CommitAnchor will send a commit transaction to Ethereum.
func (s *service) CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, documentProofs [][32]byte)  error {
	did,err := s.getDID(ctx)
	if err != nil {
		return err
	}

	tc, err := contextutil.Account(ctx)
	if err != nil {
		return err
	}

	conn := s.client
	opts, err := conn.GetTxOpts(tc.GetEthereumDefaultAccountName())
	if err != nil {
		return err
	}

	h, err := conn.GetEthClient().HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}

	cd := NewCommitData(h.Number.Uint64(), anchorID, documentRoot, documentProofs)

	log.Info("Add Anchor to Commit %s from did:%s", anchorID.BigInt().String(), did.ToAddress().String())
	txID, done, err := s.txManager.ExecuteWithinTX(context.Background(), did, uuid.Nil, "Check TX for anchor commit",
		s.ethereumTX(opts, s.anchorRepositoryContract.Commit, cd.AnchorID.BigInt(), cd.DocumentRoot, cd.DocumentProofs))
	if err != nil {
		return err
	}

	isDone := <-done
	// non async task
	if !isDone {
		return errors.New("add key  TX failed: txID:%s", txID.String())

	}
	return err
}

// sendPreCommitTransaction sends the actual transaction to the ethereum node.
func sendPreCommitTransaction(contract anchorRepositoryContract, opts *bind.TransactOpts, preCommitData *PreCommitData) error {

	//preparation of data in specific types for the call to Ethereum
	schemaVersion := big.NewInt(int64(preCommitData.SchemaVersion))

	//TODO old parameters needs an update
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
	anchorPreCommittedEvents := make(chan *AnchorContractAnchorPreCommitted)
	confirmations = make(chan *WatchPreCommit)
	go waitAndRoutePreCommitEvent(watchOpts.Context, anchorPreCommittedEvents, confirmations, preCommitData)

	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = contractEvent.WatchAnchorPreCommitted(watchOpts, anchorPreCommittedEvents, []common.Address{from}, []*big.Int{preCommitData.AnchorID.BigInt()})
	if err != nil {
		wError := errors.New("Could not subscribe to event logs for anchor registration: %v", err)
		log.Errorf("Failed to watch anchor registered event: %v", wError.Error())
		cancelFunc() // cancel the event router
		return confirmations, wError
	}
	return confirmations, nil
}

// setUpCommitEventListener sets up the listened for the "AnchorCommitted" event to notify the upstream code
// about successful mining/creation of a commit
func (s *service) setUpCommitEventListener(timeout time.Duration, from common.Address, commitData *CommitData) (confirmations chan *WatchCommit, err error) {
	confirmations = make(chan *WatchCommit)
	asyncRes, err := s.queue.EnqueueJob(anchorRepositoryConfirmationTaskName, map[string]interface{}{
		anchorIDParam: commitData.AnchorID,
		addressParam:  from,
		blockHeight:   commitData.BlockHeight,
	})
	if err != nil {
		return nil, err
	}

	go waitAndRouteCommitEvent(timeout, asyncRes, confirmations, commitData)
	return confirmations, nil
}

// waitAndRoutePreCommitEvent notifies the confirmations channel whenever a pre-commit is being noted as Ethereum event
func waitAndRoutePreCommitEvent(ctx context.Context, conf <-chan *AnchorContractAnchorPreCommitted, confirmations chan<- *WatchPreCommit, preCommitData *PreCommitData) {
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
