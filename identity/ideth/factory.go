package ideth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("identity")

const identityCreatedEventName = "IdentityCreated(address)"

type factory struct {
	factoryAddress  common.Address
	factoryContract *FactoryContract
	client          ethereum.Client
	jobManager      jobs.Manager
	queue           *queue.Server
	config          identity.Config
}

// NewFactory returns a new identity factory service
func NewFactory(factoryContract *FactoryContract, client ethereum.Client, jobManager jobs.Manager,
	queue *queue.Server, factoryAddress common.Address, conf identity.Config) identity.Factory {
	return &factory{factoryAddress: factoryAddress, factoryContract: factoryContract, client: client, jobManager: jobManager, queue: queue, config: conf}
}

func (s *factory) getNonceAt(ctx context.Context, address common.Address) (uint64, error) {
	// TODO: add blockNumber of the transaction which created the contract
	return s.client.GetEthClient().NonceAt(ctx, s.factoryAddress, nil)
}

// CalculateCreatedAddress calculates the Ethereum address based on address and nonce
func CalculateCreatedAddress(address common.Address, nonce uint64) common.Address {
	// How is a Ethereum address calculated:
	// See https://ethereum.stackexchange.com/questions/760/how-is-the-address-of-an-ethereum-contract-computed
	return crypto.CreateAddress(address, nonce)
}

func (s *factory) createIdentityTX(opts *bind.TransactOpts) func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
	return func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
		ethTX, err := s.client.SubmitTransactionWithRetries(s.factoryContract.CreateIdentity, opts)
		if err != nil {
			errOut <- err
			log.Infof("Failed to send identity for creation: %v", err)
			return
		}

		log.Infof("Sent off identity creation Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", ethTX.Hash(), ethTX.Nonce(), ethTX.CheckNonce())
		log.Infof("Transfer pending: 0x%x\n", ethTX.Hash())

		res, err := ethereum.QueueEthTXStatusTaskWithValue(accountID, jobID, ethTX.Hash(), s.queue, &jobs.JobValue{Key: identityCreatedEventName, KeyIdx: 0})
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

func (s *factory) CalculateIdentityAddress(ctx context.Context) (*common.Address, error) {
	nonce, err := s.getNonceAt(ctx, s.factoryAddress)
	if err != nil {
		return nil, err
	}

	identityAddress := CalculateCreatedAddress(s.factoryAddress, nonce)
	log.Infof("Calculated Address of the identity contract: 0x%x\n", identityAddress)
	return &identityAddress, nil
}

func isIdentityContract(identityAddress common.Address, client ethereum.Client) error {
	contractCode, err := client.GetEthClient().CodeAt(context.Background(), identityAddress, nil)
	if err != nil {
		return err
	}

	if len(contractCode) == 0 {
		return errors.New("bytecode for deployed identity contract %s not correct", identityAddress.String())
	}

	return nil
}

func (s *factory) IdentityExists(did *identity.DID) (exists bool, err error) {
	opts, cancel := s.client.GetGethCallOpts(false)
	defer cancel()
	valid, err := s.factoryContract.CreatedIdentity(opts, did.ToAddress())
	if err != nil {
		return false, err
	}
	return valid, nil
}

func (s *factory) CreateIdentity(ctx context.Context) (did *identity.DID, err error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	opts, err := s.client.GetTxOpts(ctx, tc.GetEthereumDefaultAccountName())
	if err != nil {
		log.Infof("Failed to get txOpts from Ethereum client: %v", err)
		return nil, err
	}

	opts.GasLimit = s.config.GetEthereumGasLimit(config.IDCreate)
	calcIdentityAddress, err := s.CalculateIdentityAddress(ctx)
	if err != nil {
		return nil, err
	}

	createdDID := identity.NewDID(*calcIdentityAddress)

	jobID, done, err := s.jobManager.ExecuteWithinJob(contextutil.Copy(ctx), createdDID, jobs.NilJobID(), "Check Job for create identity status", s.createIdentityTX(opts))
	if err != nil {
		return nil, err
	}

	err = <-done
	// non async task
	if err != nil {
		return nil, errors.New("Create Identity Job failed: jobID:%s with error [%s]", jobID.String(), err)
	}

	tx, err := s.jobManager.GetJob(createdDID, jobID)
	if err != nil {
		return nil, err
	}
	idCreated, ok := tx.Values[identityCreatedEventName]
	if !ok {
		return nil, errors.New("Couldn't find value for %s", identityCreatedEventName)
	}
	createdAddr := common.BytesToAddress(idCreated.Value)
	log.Infof("ID Created with address: %s", createdAddr.Hex())

	if calcIdentityAddress.Hex() != createdAddr.Hex() {
		log.Infof("[Recovered] Found race condition creating identity, calculatedDID[%s] vs createdDID[%s]", calcIdentityAddress.Hex(), createdAddr.Hex())
	}

	createdDID = identity.NewDID(createdAddr)
	exists, err := s.IdentityExists(&createdDID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("Identity %s not found in factory registry", createdDID.String())
	}

	return &createdDID, nil
}

type factoryV2 struct {
	factoryAddress  common.Address
	factoryContract *FactoryContract
	client          ethereum.Client
	config          identity.Config
}

func (f factoryV2) IdentityExists(did identity.DID) (exists bool, err error) {
	opts, cancel := f.client.GetGethCallOpts(false)
	defer cancel()
	valid, err := f.factoryContract.CreatedIdentity(opts, did.ToAddress())
	if err != nil {
		return false, err
	}
	return valid, nil
}

func (f factoryV2) NextIdentityAddress() (did identity.DID, err error) {
	nonce, err := f.client.GetEthClient().PendingNonceAt(context.Background(), f.factoryAddress)
	if err != nil {
		return did, fmt.Errorf("failed to fetch identity factory nonce: %w", err)
	}

	addr := CalculateCreatedAddress(f.factoryAddress, nonce)
	return identity.NewDID(addr), nil
}

func (f factoryV2) CreateIdentity(ethAccount string, manager common.Address, keys []identity.Key) (transaction *types.
	Transaction, err error) {
	opts, err := f.client.GetTxOpts(context.Background(), ethAccount)
	if err != nil {
		log.Infof("Failed to get txOpts from Ethereum client: %v", err)
		return nil, err
	}

	opts.GasLimit = f.config.GetEthereumGasLimit(config.IDCreate)
	ethKeys, purposes := convertKeysToEth(keys)
	ethTX, err := f.client.SubmitTransactionWithRetries(
		f.factoryContract.CreateIdentityFor, opts, manager, ethKeys, purposes)
	if err != nil {
		return nil, fmt.Errorf("failed to submit eth transaction: %w", err)
	}

	log.Infof("Sent off identity creation Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", ethTX.Hash(), ethTX.Nonce(), ethTX.CheckNonce())
	return ethTX, nil
}

func convertKeysToEth(keys []identity.Key) (ethKeys [][32]byte, purposes []*big.Int) {
	for _, k := range keys {
		ethKeys = append(ethKeys, k.GetKey())
		purposes = append(purposes, k.GetPurpose())
	}

	return ethKeys, purposes
}
