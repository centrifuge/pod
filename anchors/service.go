package anchors

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/transactions"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

type anchorRepositoryContract interface {
	PreCommit(opts *bind.TransactOpts, _anchorID *big.Int, signingRoot [32]byte, expirationBlock *big.Int) (*types.Transaction, error)
	Commit(opts *bind.TransactOpts, anchorID *big.Int, documentRoot [32]byte, documentProofs [][32]byte) (*types.Transaction, error)
	GetAnchorById(opts *bind.CallOpts, anchorID *big.Int) (struct {
		AnchorID     *big.Int
		DocumentRoot [32]byte
	}, error)
}

type service struct {
	config                   Config
	anchorRepositoryContract anchorRepositoryContract
	client                   ethereum.Client
	queue                    *queue.Server
	txManager                transactions.Manager
}

func newService(config Config, anchorContract anchorRepositoryContract, queue *queue.Server, client ethereum.Client, txManager transactions.Manager) AnchorRepository {
	return &service{config: config, anchorRepositoryContract: anchorContract, client: client, queue: queue, txManager: txManager}
}

// GetDocumentRootOf takes an anchorID and returns the corresponding documentRoot from the chain.
func (s *service) GetDocumentRootOf(anchorID AnchorID) (docRoot DocumentRoot, err error) {
	// Ignoring cancelFunc as code will block until response or timeout is triggered
	opts, _ := s.client.GetGethCallOpts(false)
	r, err := s.anchorRepositoryContract.GetAnchorById(opts, anchorID.BigInt())
	return r.DocumentRoot, err
}

// PreCommitAnchor will call the transaction PreCommit on the smart contract
func (s *service) PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot, centID identity.DID, signature []byte, expirationBlock *big.Int) (confirmations <-chan *WatchPreCommit, err error) {
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
func (s service) ethereumTX(opts *bind.TransactOpts, contractMethod interface{}, params ...interface{}) func(accountID identity.DID, txID transactions.TxID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID identity.DID, txID transactions.TxID, txMan transactions.Manager, errOut chan<- error) {

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

// getDID returns DID from context.Account
// TODO use did.NewDIDFromContext as soon as IDConfig is deleted
func getDID(ctx context.Context) (identity.DID, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return identity.DID{}, err
	}

	addressByte, err := tc.GetIdentityID()
	if err != nil {
		return identity.DID{}, err
	}
	return identity.NewDID(common.BytesToAddress(addressByte)), nil
}

// CommitAnchor will send a commit transaction to Ethereum.
func (s *service) CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, documentProofs [][32]byte) (chan bool, error) {
	did, err := getDID(ctx)
	if err != nil {
		return nil, err
	}

	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	uuid := contextutil.TX(ctx)

	conn := s.client
	opts, err := conn.GetTxOpts(tc.GetEthereumDefaultAccountName())
	if err != nil {
		return nil, err
	}

	h, err := conn.GetEthClient().HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	cd := NewCommitData(h.Number.Uint64(), anchorID, documentRoot, documentProofs)

	log.Info("Add Anchor to Commit %s from did:%s", anchorID.BigInt().String(), did.ToAddress().String())
	_, done, err := s.txManager.ExecuteWithinTX(ctx, did, uuid, "Check TX for anchor commit",
		s.ethereumTX(opts, s.anchorRepositoryContract.Commit, cd.AnchorID.BigInt(), cd.DocumentRoot, cd.DocumentProofs))
	if err != nil {
		return nil, err
	}

	return done, nil

}

// sendPreCommitTransaction sends the actual transaction to the ethereum node.
func sendPreCommitTransaction(contract anchorRepositoryContract, opts *bind.TransactOpts, preCommitData *PreCommitData) error {

	//TODO implement
	return nil
}
