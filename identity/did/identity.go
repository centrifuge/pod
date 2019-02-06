package did

import (
	"context"
	"math/big"
	"strings"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/satori/go.uuid"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	id "github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DID stores the identity address of the user
type DID common.Address

func (d DID) toAddress() common.Address {
	return common.Address(d)
}

// NewDID returns a DID based on a common.Address
func NewDID(address common.Address) DID {
	return DID(address)
}

// NewDIDFromString returns a DID based on a hex string
func NewDIDFromString(address string) DID {
	return DID(common.HexToAddress(address))
}

// Identity interface contains the methods to interact with the identity contract
type Identity interface {
	// AddKey adds a key to identity contract
	AddKey(ctx context.Context, key Key) error

	// GetKey return a key from the identity contract
	GetKey(ctx context.Context, key [32]byte) (*KeyResponse, error)

	// RawExecute calls the execute method on the identity contract
	RawExecute(ctx context.Context, to common.Address, data []byte) error

	// Execute creates the abi encoding an calls the execute method on the identity contract
	Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) error
}

type contract interface {

	// calls
	GetKey(opts *bind.CallOpts, _key [32]byte) (struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}, error)

	// transactions
	AddKey(opts *bind.TransactOpts, _key [32]byte, _purpose *big.Int, _keyType *big.Int) (*types.Transaction, error)

	Execute(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error)
}

type identity struct {
	config    id.Config
	client    ethereum.Client
	txManager transactions.Manager
	queue     *queue.Server
}

func (i identity) prepareTransaction(did DID) (contract, *bind.TransactOpts, error) {
	opts, err := i.client.GetTxOpts(i.config.GetEthereumDefaultAccountName())
	if err != nil {
		log.Infof("Failed to get txOpts from Ethereum client: %v", err)
		return nil, nil, err
	}

	contract, err := i.bindContract(did)
	if err != nil {
		return nil, nil, err
	}

	return contract, opts, nil

}

func (i identity) prepareCall(did DID) (contract, *bind.CallOpts, context.CancelFunc, error) {
	opts, cancelFunc := i.client.GetGethCallOpts(false)

	contract, err := i.bindContract(did)
	if err != nil {
		return nil, nil, nil, err
	}

	return contract, opts, cancelFunc, nil

}

func (i identity) bindContract(did DID) (contract, error) {
	contract, err := NewIdentityContract(did.toAddress(), i.client.GetEthClient())
	if err != nil {
		return nil, errors.New("Could not bind identity contract: %v", err)
	}

	return contract, nil

}

// NewIdentity creates a instance of an identity
func NewIdentity(config id.Config, client ethereum.Client, txManager transactions.Manager, queue *queue.Server) Identity {
	return identity{config: config, client: client, txManager: txManager, queue: queue}
}

func logTxHash(tx *types.Transaction) {
	log.Infof("Ethereum transaction created. Hash [%x] and Nonce [%v] and Check [%v]", tx.Hash(), tx.Nonce(), tx.CheckNonce())
	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
}

func (i identity) getDID(ctx context.Context) (did DID, err error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return did, err
	}

	addressByte, err := tc.GetIdentityID()
	if err != nil {
		return did, err
	}
	did = NewDID(common.BytesToAddress(addressByte))
	return did, nil

}

func (i identity) AddKey(ctx context.Context, key Key) error {
	did, err := i.getDID(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := i.prepareTransaction(did)
	if err != nil {
		return err
	}

	// TODO: did can be passed instead of randomCentID after CentID is DID
	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), id.RandomCentID(), uuid.Nil, "Check TX for add key", i.addKeyTX(opts, contract, key))
	if err != nil {
		return err
	}

	isDone := <-done
	// non async task
	if !isDone {
		return errors.New("add key  TX failed: txID:%s", txID.String())

	}
	return nil

}

func (i identity) addKeyTX(opts *bind.TransactOpts, identityContract contract, key Key) func(accountID id.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID id.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
		ethTX, err := i.client.SubmitTransactionWithRetries(identityContract.AddKey, opts, key.GetKey(), key.GetPurpose(), key.GetType())
		if err != nil {
			errOut <- err
			return
		}
		logTxHash(ethTX)

		res, err := ethereum.QueueEthTXStatusTask(accountID, txID, ethTX.Hash(), i.queue)
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

func (i identity) rawExecuteTX(opts *bind.TransactOpts, identityContract contract, to common.Address, value *big.Int, data []byte) func(accountID id.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID id.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
		ethTX, err := i.client.SubmitTransactionWithRetries(identityContract.Execute, opts, to, value, data)
		if err != nil {
			errOut <- err
			return
		}
		logTxHash(ethTX)

		res, err := ethereum.QueueEthTXStatusTask(accountID, txID, ethTX.Hash(), i.queue)
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

func (i identity) GetKey(ctx context.Context, key [32]byte) (*KeyResponse, error) {
	did, err := i.getDID(ctx)
	if err != nil {
		return nil, err
	}
	contract, opts, _, err := i.prepareCall(did)
	if err != nil {
		return nil, err
	}

	result, err := contract.GetKey(opts, key)

	if err != nil {
		return nil, errors.New("Could not call identity contract: %v", err)
	}

	return &KeyResponse{result.Key, result.Purposes, result.RevokedAt}, nil

}

func (i identity) RawExecute(ctx context.Context, to common.Address, data []byte) error {
	did, err := i.getDID(ctx)
	if err != nil {
		return err
	}
	contract, opts, err := i.prepareTransaction(did)
	if err != nil {
		return err
	}

	// default: no ether should be send
	value := big.NewInt(0)

	// TODO: did can be passed instead of randomCentID after CentID is DID
	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), id.RandomCentID(), uuid.Nil, "Check TX for execute", i.rawExecuteTX(opts, contract, to, value, data))
	if err != nil {
		return err
	}

	isDone := <-done
	// non async task
	if !isDone {
		return errors.New("raw execute TX failed: txID:%s", txID.String())

	}
	return nil

}

func (i identity) Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) error {
	abi, err := abi.JSON(strings.NewReader(contractAbi))
	if err != nil {
		return err
	}

	// Pack encodes the parameters and additionally checks if the method and arguments are defined correctly
	data, err := abi.Pack(methodName, args...)
	if err != nil {
		return err
	}
	return i.RawExecute(ctx, to, data)
}
