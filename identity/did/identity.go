package did

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
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
	AddKey(ctx context.Context, key Key) (chan *ethereum.WatchTransaction, error)

	// GetKey return a key from the identity contract
	GetKey(ctx context.Context, key [32]byte) (*KeyResponse, error)

	// RawExecute calls the execute method on the identity contract
	RawExecute(ctx context.Context, to common.Address, data []byte) (chan *ethereum.WatchTransaction, error)

	// Execute creates the abi encoding an calls the execute method on the identity contract
	Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) (chan *ethereum.WatchTransaction, error)
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
	client ethereum.Client
}

func (i identity) prepareTransaction(ctx context.Context, did DID) (contract, *bind.TransactOpts, error) {
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
func NewIdentity(client ethereum.Client) Identity {
	return identity{client: client}
}

// TODO: will be replaced with statusTask
func waitForTransaction(client ethereum.Client, txHash common.Hash, txStatus chan *ethereum.WatchTransaction) {
	time.Sleep(3000 * time.Millisecond)
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		txStatus <- &ethereum.WatchTransaction{Error: err}
	}
	txStatus <- &ethereum.WatchTransaction{Status: receipt.Status}

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

func (i identity) AddKey(ctx context.Context, key Key) (chan *ethereum.WatchTransaction, error) {
	did, err := i.getDID(ctx)
	if err != nil {
		return nil, err
	}

	contract, opts, err := i.prepareTransaction(ctx, did)
	if err != nil {
		return nil, err
	}

	tx, err := i.client.SubmitTransactionWithRetries(contract.AddKey, opts, key.GetKey(), key.GetPurpose(), key.GetType())
	if err != nil {
		log.Infof("could not addKey to identity contract: %v[txHash: %s] : %v", tx.Hash(), err)
		return nil, errors.New("could not addKey to identity contract: %v", err)
	}
	logTxHash(tx)

	txStatus := make(chan *ethereum.WatchTransaction)

	// TODO will be replaced with transaction Status task
	go waitForTransaction(i.client, tx.Hash(), txStatus)

	return txStatus, nil

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

func (i identity) RawExecute(ctx context.Context, to common.Address, data []byte) (chan *ethereum.WatchTransaction, error) {
	did, err := i.getDID(ctx)
	if err != nil {
		return nil, err
	}
	contract, opts, err := i.prepareTransaction(ctx, did)
	if err != nil {
		return nil, err
	}

	// default: no ether should be send
	value := big.NewInt(0)

	tx, err := i.client.SubmitTransactionWithRetries(contract.Execute, opts, to, value, data)
	if err != nil {
		log.Infof("could not call execute method on identity contract: %v[txHash: %s] toAddress: %s : %v", tx.Hash(), to.String(), err)
		return nil, errors.New("could not execute to identity contract: %v", err)
	}
	logTxHash(tx)

	txStatus := make(chan *ethereum.WatchTransaction)
	// TODO will be replaced with transaction Status task
	go waitForTransaction(i.client, tx.Hash(), txStatus)

	return txStatus, nil

}

func (i identity) Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) (chan *ethereum.WatchTransaction, error) {
	abi, err := abi.JSON(strings.NewReader(contractAbi))
	if err != nil {
		return nil, err
	}

	// Pack encodes the parameters and additionally checks if the method and arguments are defined correctly
	data, err := abi.Pack(methodName, args...)
	if err != nil {
		return nil, err
	}
	return i.RawExecute(ctx, to, data)
}
