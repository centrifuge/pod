package ideth

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	id "github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type contract interface {

	// Ethereum Calls
	GetKey(opts *bind.CallOpts, _key [32]byte) (struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt *big.Int
	}, error)

	IsSignedWithPurpose(opts *bind.CallOpts, message [32]byte, _signature []byte, _purpose *big.Int) (bool, error)

	GetKeysByPurpose(opts *bind.CallOpts, purpose *big.Int) ([][32]byte, error)

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

func (i service) prepareTransaction(ctx context.Context, did id.DID) (contract, *bind.TransactOpts, error) {
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

func (i service) prepareCall(did id.DID) (contract, *bind.CallOpts, context.CancelFunc, error) {
	opts, cancelFunc := i.client.GetGethCallOpts(false)

	contract, err := i.bindContract(did)
	if err != nil {
		return nil, nil, nil, err
	}

	return contract, opts, cancelFunc, nil

}

func (i service) bindContract(did id.DID) (contract, error) {
	contract, err := NewIdentityContract(did.ToAddress(), i.client.GetEthClient())
	if err != nil {
		return nil, errors.New("Could not bind identity contract: %v", err)
	}

	return contract, nil

}

// NewService creates a instance of the identity service
func NewService(client ethereum.Client, txManager transactions.Manager, queue *queue.Server) id.ServiceDID {
	return service{client: client, txManager: txManager, queue: queue}
}

func logTxHash(tx *types.Transaction) {
	log.Infof("Ethereum transaction created. Hash [%x] and Nonce [%v] and Check [%v]", tx.Hash(), tx.Nonce(), tx.CheckNonce())
	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
}

// AddKey adds a key to identity contract
func (i service) AddKey(ctx context.Context, key id.KeyDID) error {
	DID, err := NewDIDFromContext(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := i.prepareTransaction(ctx, DID)
	if err != nil {
		return err
	}

	log.Info("Add key to identity contract %s", DID.ToAddress().String())
	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), DID, transactions.NilTxID(), "Check TX for add key",
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
	DID, err := NewDIDFromContext(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := i.prepareTransaction(ctx, DID)
	if err != nil {
		return err
	}

	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), DID, transactions.NilTxID(), "Check TX for add multi purpose key",
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
	DID, err := NewDIDFromContext(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := i.prepareTransaction(ctx, DID)
	if err != nil {
		return err
	}

	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), DID, transactions.NilTxID(), "Check TX for revoke key",
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
func (i service) ethereumTX(opts *bind.TransactOpts, contractMethod interface{}, params ...interface{}) func(accountID id.DID, txID transactions.TxID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID id.DID, txID transactions.TxID, txMan transactions.Manager, errOut chan<- error) {
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
func (i service) GetKey(did id.DID, key [32]byte) (*id.KeyResponse, error) {
	contract, opts, _, err := i.prepareCall(did)
	if err != nil {
		return nil, err
	}

	result, err := contract.GetKey(opts, key)

	if err != nil {
		return nil, errors.New("Could not call identity contract: %v", err)
	}

	return &id.KeyResponse{result.Key, result.Purposes, result.RevokedAt}, nil

}

// IsSignedWithPurpose verifies if a message is signed with one of the identities specific purpose keys
func (i service) IsSignedWithPurpose(did id.DID, message [32]byte, signature []byte, purpose *big.Int) (bool, error) {
	contract, opts, _, err := i.prepareCall(did)
	if err != nil {
		return false, err
	}

	return contract.IsSignedWithPurpose(opts, message, signature, purpose)

}

// RawExecute calls the execute method on the identity contract
func (i service) RawExecute(ctx context.Context, to common.Address, data []byte) error {
	DID, err := NewDIDFromContext(ctx)
	if err != nil {
		return err
	}
	contract, opts, err := i.prepareTransaction(ctx, DID)
	if err != nil {
		return err
	}

	// default: no ether should be send
	value := big.NewInt(0)

	txID, done, err := i.txManager.ExecuteWithinTX(context.Background(), DID, transactions.NilTxID(), "Check TX for execute", i.ethereumTX(opts, contract.Execute, to, value, data))
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

func (i service) GetKeysByPurpose(did id.DID, purpose *big.Int) ([][32]byte, error) {
	contract, opts, _, err := i.prepareCall(did)
	if err != nil {
		return nil, err
	}

	return contract.GetKeysByPurpose(opts, purpose)

}

// CurrentP2PKey returns the latest P2P key
func (i service) CurrentP2PKey(did id.DID) (ret string, err error) {
	keys, err := i.GetKeysByPurpose(did, &(id.KeyPurposeP2PDiscovery.Value))
	if err != nil {
		return ret, err
	}

	lastKey := keys[len(keys)-1]
	key, err := i.GetKey(did, lastKey)
	if err != nil {
		return "", err
	}

	if key.RevokedAt.Cmp(big.NewInt(0)) != 0 {
		return "", errors.New("current p2p key has been revoked")
	}

	p2pID, err := ed25519.PublicKeyToP2PKey(key.Key)
	if err != nil {
		return ret, err
	}

	return p2pID.Pretty(), nil
}

// GetClientP2PURL returns the p2p url associated with the did
func (i service) GetClientP2PURL(did id.DID) (string, error) {
	p2pID, err := i.CurrentP2PKey(did)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/ipfs/%s", p2pID), nil
}

//Exists checks if an identity contract exists
func (i service) Exists(ctx context.Context, did id.DID) error {
	return isIdentityContract(did.ToAddress(), i.client)
}

// ValidateKey checks if a given key is valid for the given centrifugeID.
func (i service) ValidateKey(ctx context.Context, did id.DID, key []byte, purpose *big.Int, validateAt *time.Time) error {
	contract, opts, _, err := i.prepareCall(did)
	if err != nil {
		return err
	}

	key32, err := utils.SliceToByte32(key)
	if err != nil {
		return err
	}

	ethKey, err := contract.GetKey(opts, key32)
	if err != nil {
		return err
	}

	// if revoked
	if ethKey.RevokedAt.Cmp(big.NewInt(0)) > 0 {
		// if a specific time for validation is provided then we validate if a revoked key was revoked before the provided time
		if validateAt != nil {
			revokedAtBlock, err := i.client.GetEthClient().BlockByNumber(ctx, ethKey.RevokedAt)
			if err != nil {
				return err
			}

			if big.NewInt(validateAt.Unix()).Cmp(revokedAtBlock.Time()) > 0 {
				return errors.New("the given key [%x] for purpose [%s] has been revoked before provided time %s", key, purpose.String(), validateAt.String())
			}
		} else {
			return errors.New("the given key [%x] for purpose [%s] has been revoked and not valid anymore", key, purpose.String())
		}
	}

	for _, p := range ethKey.Purposes {
		if p.Cmp(purpose) == 0 {
			return nil
		}
	}

	return errors.New("identity contract doesn't have a key with requested purpose")
}

// GetClientsP2PURLs returns p2p urls associated with each centIDs
// will error out at first failure
func (i service) GetClientsP2PURLs(dids []*id.DID) ([]string, error) {
	urls := make([]string, len(dids))

	for idx, did := range dids {
		url, err := i.GetClientP2PURL(*did)
		if err != nil {
			return nil, err
		}
		urls[idx] = url

	}

	return urls, nil
}

func convertAccountKeysToKeyDID(accKeys map[string]config.IDKey) (map[string]id.KeyDID, error) {
	keys := map[string]id.KeyDID{}
	for k, v := range accKeys {
		pk32, err := utils.SliceToByte32(v.PublicKey)
		if err != nil {
			return nil, err
		}
		v := id.GetPurposeByName(k).Value
		keys[k] = id.NewKey(pk32, &v, big.NewInt(id.KeyTypeECDSA))
	}
	return keys, nil
}

// AddKeysForAccount adds the keys from the config to the smart contracts
func (i service) AddKeysForAccount(acc config.Account) error {
	tctx, err := contextutil.New(context.Background(), acc)
	if err != nil {
		return err
	}

	accKeys, err := acc.GetKeys()
	if err != nil {
		return err
	}

	keys, err := convertAccountKeysToKeyDID(accKeys)
	if err != nil {
		return err
	}

	err = i.AddKey(tctx, keys[id.KeyPurposeAction.Name])
	if err != nil {
		return err
	}

	err = i.AddKey(tctx, keys[id.KeyPurposeP2PDiscovery.Name])
	if err != nil {
		return err
	}

	err = i.AddKey(tctx, keys[id.KeyPurposeSigning.Name])
	if err != nil {
		return err
	}

	return nil
}

// ValidateSignature validates a signature on a message based on identity data
func (i service) ValidateSignature(did id.DID, pubKey []byte, signature []byte, message []byte, timestamp time.Time) error {
	err := i.ValidateKey(context.Background(), did, pubKey, &(id.KeyPurposeSigning.Value), &timestamp)
	if err != nil {
		return err
	}

	if !crypto.VerifyMessage(pubKey, message, signature, crypto.CurveSecp256K1) {
		return errors.New("error when validating signature")
	}

	return nil
}

// NewDIDFromContext returns DID from context.Account
func NewDIDFromContext(ctx context.Context) (id.DID, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return id.DID{}, err
	}

	addressByte, err := tc.GetIdentityID()
	if err != nil {
		return id.DID{}, err
	}
	return id.NewDID(common.BytesToAddress(addressByte)), nil
}
