package did

import (
	"context"
	"math/big"
	"time"

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
	AddKey(did *DID, key Key) (chan *ethereum.WatchTransaction, error)
	// GetKey return a key from the identity contract
	GetKey(did *DID, key [32]byte) (*KeyResponse, error)
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
}

type identity struct {
	contract contract
	config   id.Config
	client   ethereum.Client
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
	opts, cancelFunc := i.client.GetGethCallOpts()

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
func NewIdentity(config id.Config, client ethereum.Client) Identity {
	return identity{config: config, client: client}
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
	log.Infof("Sent off identity creation Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", tx.Hash(), tx.Nonce(), tx.CheckNonce())
	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
}

func (i identity) AddKey(did *DID, key Key) (chan *ethereum.WatchTransaction, error) {

	contract, opts, err := i.prepareTransaction(*did)
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

func (i identity) GetKey(did *DID, key [32]byte) (*KeyResponse, error) {
	contract, opts, _, err := i.prepareCall(*did)
	if err != nil {
		return nil, err
	}

	result, err := contract.GetKey(opts, key)

	if err != nil {
		return nil, errors.New("Could not call identity contract: %v", err)
	}

	return &KeyResponse{result.Key, result.Purposes, result.RevokedAt}, nil

}
