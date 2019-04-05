package anchors

import (
	"context"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type anchorRepositoryContract interface {
	PreCommit(opts *bind.TransactOpts, anchorId *big.Int, signingRoot [32]byte) (*types.Transaction, error)
	Commit(opts *bind.TransactOpts, anchorId *big.Int, documentRoot [32]byte, proof [32]byte) (*types.Transaction, error)
	GetAnchorById(opts *bind.CallOpts, id *big.Int) (struct {
		AnchorId     *big.Int
		DocumentRoot [32]byte
		BlockNumber  uint32
	}, error)
	HasValidPreCommit(opts *bind.CallOpts, anchorId *big.Int) (bool, error)
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

// HasValidPreCommit checks if the given anchorID has a valid pre-commit
func (s *service) HasValidPreCommit(anchorID AnchorID) bool {
	// Ignoring cancelFunc as code will block until response or timeout is triggered
	opts, _ := s.client.GetGethCallOpts(false)
	r, err := s.anchorRepositoryContract.HasValidPreCommit(opts, anchorID.BigInt())
	if err != nil {
		return false
	}
	return r
}

// GetAnchorData takes an anchorID and returns the corresponding documentRoot from the chain.
func (s *service) GetAnchorData(anchorID AnchorID) (docRoot DocumentRoot, anchoredTime time.Time, err error) {
	// Ignoring cancelFunc as code will block until response or timeout is triggered
	opts, _ := s.client.GetGethCallOpts(false)
	r, err := s.anchorRepositoryContract.GetAnchorById(opts, anchorID.BigInt())
	if err != nil {
		return docRoot, anchoredTime, err
	}

	blk, err := s.client.GetBlockByNumber(context.Background(), big.NewInt(int64(r.BlockNumber)))
	if err != nil {
		return docRoot, anchoredTime, err
	}

	return r.DocumentRoot, time.Unix(blk.Time().Int64(), 0), err
}

// PreCommitAnchor will call the transaction PreCommit on the smart contract
func (s *service) PreCommitAnchor(ctx context.Context, anchorID AnchorID, signingRoot DocumentRoot) (confirmations chan bool, err error) {
	did, err := getDID(ctx)
	if err != nil {
		return nil, err
	}

	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	txID := contextutil.TX(ctx)

	conn := s.client
	opts, err := conn.GetTxOpts(ctx, tc.GetEthereumDefaultAccountName())
	if err != nil {
		return nil, err
	}

	opts.GasLimit = s.config.GetEthereumGasLimit(config.AnchorPreCommit)
	pc := newPreCommitData(anchorID, signingRoot)
	if err != nil {
		return confirmations, err
	}

	log.Infof("Add Anchor to Pre-commit %s from did:%s", anchorID.String(), did.ToAddress().String())
	_, done, err := s.txManager.ExecuteWithinTX(ctx, did, txID, "Check TX for anchor commit",
		s.ethereumTX(opts, s.anchorRepositoryContract.PreCommit, pc.AnchorID.BigInt(), pc.SigningRoot))
	if err != nil {
		return nil, err
	}

	return done, nil
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
func (s *service) CommitAnchor(ctx context.Context, anchorID AnchorID, documentRoot DocumentRoot, proof [32]byte) (chan bool, error) {
	did, err := getDID(ctx)
	if err != nil {
		return nil, err
	}

	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	txID := contextutil.TX(ctx)

	conn := s.client
	opts, err := conn.GetTxOpts(ctx, tc.GetEthereumDefaultAccountName())
	if err != nil {
		return nil, err
	}

	opts.GasLimit = s.config.GetEthereumGasLimit(config.AnchorCommit)
	h, err := conn.GetEthClient().HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	cd := NewCommitData(h.Number.Uint64(), anchorID, documentRoot, proof)

	log.Infof("Add Anchor to Commit %s from did:%s", anchorID.String(), did.ToAddress().String())
	_, done, err := s.txManager.ExecuteWithinTX(ctx, did, txID, "Check TX for anchor commit",
		s.ethereumTX(opts, s.anchorRepositoryContract.Commit, cd.AnchorID.BigInt(), cd.DocumentRoot, cd.DocumentProof))
	if err != nil {
		return nil, err
	}

	return done, nil
}
