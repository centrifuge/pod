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

// ToAddress returns the DID as common.Address
func (d DID) ToAddress() common.Address {
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

// Service interface contains the methods to interact with the identity contract
type Service interface {
	// AddKey adds a key to identity contract
	AddKey(ctx context.Context, key Key) error

	// GetKey return a key from the identity contract
	GetKey(ctx context.Context, key [32]byte) (*KeyResponse, error)

	// RawExecute calls the execute method on the identity contract
	RawExecute(ctx context.Context, to common.Address, data []byte) error

	// Execute creates the abi encoding an calls the execute method on the identity contract
	Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) error

	// IsSignedWithPurpose verifies if a message is signed with one of the identities specific purpose keys
	IsSignedWithPurpose(ctx context.Context, message [32]byte, _signature []byte, _purpose *big.Int) (bool, error)

	// AddMultiPurposeKey adds a key with multiple purposes
	AddMultiPurposeKey(context context.Context, key [32]byte, purposes []*big.Int, keyType *big.Int) error

	// RevokeKey revokes an existing key in the smart contract
	RevokeKey(ctx context.Context, key [32]byte) error
}

type contract interface {

	// Ethereum Calls
	GetKey(opts *bind.CallOpts, _key [32]byte) (struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}, error)

	IsSignedWithPurpose(opts *bind.CallOpts, message [32]byte, _signature []byte, _purpose *big.Int) (bool, error)

	// Ethereum Transactions
	AddKey(opts *bind.TransactOpts, _key [32]byte, _purpose *big.Int, _keyType *big.Int) (*types.Transaction, error)

	Execute(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error)

	AddMultiPurposeKey(opts *bind.TransactOpts, _key [32]byte, _purposes []*big.Int, _keyType *big.Int) (*types.Transaction, error)

	RevokeKey(opts *bind.TransactOpts, _key [32]byte) (*types.Transaction, error)
}

type service struct {
	client    ethereum.Client
	txManager transactions.Manager
	queue     *queue.Server
}

func (i service) prepareTransaction(ctx context.Context, did DID) (contract, *bind.TransactOpts, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, nil, err
	}

	opts, err := i.client.GetTxOpts(tc.GetEthereumDefaultAccountName())
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

func (i service) prepareCall(did DID) (contract, *bind.CallOpts, context.CancelFunc, error) {
	opts, cancelFunc := i.client.GetGethCallOpts(false)

	contract, err := i.bindContract(did)
	if err != nil {
		return nil, nil, nil, err
	}

	return contract, opts, cancelFunc, nil

}

func (i service) bindContract(did DID) (contract, error) {
	contract, err := NewIdentityContract(did.ToAddress(), i.client.GetEthClient())
	if err != nil {
		return nil, errors.New("Could not bind identity contract: %v", err)
	}

	return contract, nil

}

// NewService creates a instance of the identity service
func NewService(client ethereum.Client, txManager transactions.Manager, queue *queue.Server) Service {
	return service{client: client, txManager: txManager, queue: queue}
}

func logTxHash(tx *types.Transaction) {
	log.Infof("Ethereum transaction created. Hash [%x] and Nonce [%v] and Check [%v]", tx.Hash(), tx.Nonce(), tx.CheckNonce())
	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
}

func (i service) getDID(ctx context.Context) (did DID, err error) {
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

// AddKey adds a key to identity contract
func (i service) AddKey(ctx context.Context, key Key) error {
	did, err := i.getDID(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := i.prepareTransaction(ctx, did)
	if err != nil {
		return err
	}

	// TODO: did can be passed instead of randomCentID after CentID is DID
	log.Info("Add key to identity contract %s", did.ToAddress().String())
	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), id.RandomCentID(), uuid.Nil, "Check TX for add key",
		i.ethereumTX(opts, contract.AddKey, key.GetKey(), key.GetPurpose(), key.GetType()))
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

// AddMultiPurposeKey adds a key with multiple purposes
func (i service) AddMultiPurposeKey(ctx context.Context, key [32]byte, purposes []*big.Int, keyType *big.Int) error {
	did, err := i.getDID(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := i.prepareTransaction(ctx, did)
	if err != nil {
		return err
	}

	// TODO: did can be passed instead of randomCentID after CentID is DID
	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), id.RandomCentID(), uuid.Nil, "Check TX for add multi purpose key",
		i.ethereumTX(opts, contract.AddMultiPurposeKey, key, purposes, keyType))
	if err != nil {
		return err
	}

	isDone := <-done
	// non async task
	if !isDone {
		return errors.New("add key multi purpose  TX failed: txID:%s", txID.String())

	}
	return nil
}

// RevokeKey revokes an existing key in the smart contract
func (i service) RevokeKey(ctx context.Context, key [32]byte) error {
	did, err := i.getDID(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := i.prepareTransaction(ctx, did)
	if err != nil {
		return err
	}

	// TODO: did can be passed instead of randomCentID after CentID is DID
	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), id.RandomCentID(), uuid.Nil, "Check TX for revoke key",
		i.ethereumTX(opts, contract.RevokeKey, key))
	if err != nil {
		return err
	}

	isDone := <-done
	// non async task
	if !isDone {
		return errors.New("revoke key TX failed: txID:%s", txID.String())

	}
	return nil
}

// ethereumTX is submitting an Ethereum transaction and starts a task to wait for the transaction result
func (i service) ethereumTX(opts *bind.TransactOpts, contractMethod interface{}, params ...interface{}) func(accountID id.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID id.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
		ethTX, err := i.client.SubmitTransactionWithRetries(contractMethod, opts, params...)
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

// GetKey return a key from the identity contract
func (i service) GetKey(ctx context.Context, key [32]byte) (*KeyResponse, error) {
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

// IsSignedWithPurpose verifies if a message is signed with one of the identities specific purpose keys
func (i service) IsSignedWithPurpose(ctx context.Context, message [32]byte, _signature []byte, _purpose *big.Int) (bool, error) {
	did, err := i.getDID(ctx)
	if err != nil {
		return false, err
	}
	contract, opts, _, err := i.prepareCall(did)
	if err != nil {
		return false, err
	}

	return contract.IsSignedWithPurpose(opts, message, _signature, _purpose)

}

// RawExecute calls the execute method on the identity contract
func (i service) RawExecute(ctx context.Context, to common.Address, data []byte) error {
	did, err := i.getDID(ctx)
	if err != nil {
		return err
	}
	contract, opts, err := i.prepareTransaction(ctx, did)
	if err != nil {
		return err
	}

	// default: no ether should be send
	value := big.NewInt(0)

	// TODO: did can be passed instead of randomCentID after CentID is DID
	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), id.RandomCentID(), uuid.Nil, "Check TX for execute", i.ethereumTX(opts, contract.Execute, to, value, data))
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

// Execute creates the abi encoding an calls the execute method on the identity contract
func (i service) Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) error {
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
