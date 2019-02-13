package did

import (
	"context"
	"math/big"
	"strings"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/utils"

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
	GetKey(did DID, key [32]byte) (*KeyResponse, error)

	// RawExecute calls the execute method on the identity contract
	RawExecute(ctx context.Context, to common.Address, data []byte) error

	// Execute creates the abi encoding an calls the execute method on the identity contract
	Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) error

	// IsSignedWithPurpose verifies if a message is signed with one of the identities specific purpose keys
	IsSignedWithPurpose(did DID, message [32]byte, _signature []byte, _purpose *big.Int) (bool, error)

	// AddMultiPurposeKey adds a key with multiple purposes
	AddMultiPurposeKey(context context.Context, key [32]byte, purposes []*big.Int, keyType *big.Int) error

	// RevokeKey revokes an existing key in the smart contract
	RevokeKey(ctx context.Context, key [32]byte) error

	// GetClientP2PURL returns the p2p url associated with the did
	GetClientP2PURL(ctx context.Context, did DID) (string, error)

	//Exists checks if an identity contract exists
	Exists(ctx context.Context, did DID) error

	// ValidateKey checks if a given key is valid for the given centrifugeID.
	ValidateKey(ctx context.Context, did DID, key []byte, purpose int64) error

	// GetClientsP2PURLs returns p2p urls associated with each centIDs
	// will error out at first failure
	GetClientsP2PURLs(ctx context.Context, did []*DID) ([]string, error)

	// GetKeysByPurpose returns keys grouped by purpose from the identity contract.
	GetKeysByPurpose(did DID, purpose *big.Int) ([][32]byte, error)
}

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
func (i service) GetKey(did DID, key [32]byte) (*KeyResponse, error) {
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
func (i service) IsSignedWithPurpose(did DID, message [32]byte, _signature []byte, _purpose *big.Int) (bool, error) {
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


func (i service)  GetKeysByPurpose(did DID, purpose *big.Int) ([][32]byte, error) {
	contract, opts, _, err := i.prepareCall(did)
	if err != nil {
		return nil, err
	}

	return contract.GetKeysByPurpose(opts, purpose)

}


// GetClientP2PURL returns the p2p url associated with the did
func (i service)  GetClientP2PURL(ctx context.Context, did DID) (string, error) {
	// TODO implement

	return "", nil
}



//Exists checks if an identity contract exists
func (i service) Exists(ctx context.Context, did DID) error {
	return isIdentityContract(did.ToAddress(),i.client)
}

// ValidateKey checks if a given key is valid for the given centrifugeID.
func(i service) ValidateKey(ctx context.Context, did DID, key []byte, purpose int64) error {
	contract, opts, _, err := i.prepareCall(did)
	if err != nil {
		return err
	}

	key32, err := utils.SliceToByte32(key)
	if err != nil {
		return err
	}
	keys, err := contract.GetKey(opts, key32)

	for _, p := range keys.Purposes {
		if p.Cmp(big.NewInt(purpose)) == 0 {
			return nil
		}
	}

	return errors.New("identity contract doesn't have a key with requested purpose")
}

// GetClientsP2PURLs returns p2p urls associated with each centIDs
// will error out at first failure
func(i service)  GetClientsP2PURLs(ctx context.Context, did []*DID) ([]string, error) {
	// TODO implement


	return nil, nil
}

func getKeyPairsFromConfig(config config.Configuration) (map[int]Key, error) {
	keys := map[int]Key{}
	var pk []byte

	// ed25519 keys
	// KeyPurposeP2P
	pk, _, err := ed25519.GetSigningKeyPair(config.GetP2PKeyPair())
	if err != nil {
		return nil, err
	}
	pk32, err := utils.SliceToByte32(pk)
	if err != nil {
		return nil, err
	}
	keys[id.KeyPurposeP2P] = NewKey(pk32, big.NewInt(id.KeyPurposeP2P), big.NewInt(id.KeyTypeECDSA))

	// KeyPurposeSigning
	pk, _, err = ed25519.GetSigningKeyPair(config.GetSigningKeyPair())
	if err != nil {
		return nil, err
	}
	pk32, err = utils.SliceToByte32(pk)
	if err != nil {
		return nil, err
	}
	keys[id.KeyPurposeSigning] = NewKey(pk32, big.NewInt(id.KeyPurposeSigning), big.NewInt(id.KeyTypeECDSA))

	// secp256k1 keys
	// KeyPurposeEthMsgAuth
	pk, _, err = secp256k1.GetEthAuthKey(config.GetEthAuthKeyPair())
	if err != nil {
		return nil, err
	}

	address32Bytes := utils.AddressTo32Bytes(common.HexToAddress(secp256k1.GetAddress(pk)))
	keys[id.KeyPurposeEthMsgAuth] = NewKey(address32Bytes, big.NewInt(id.KeyPurposeEthMsgAuth), big.NewInt(id.KeyTypeECDSA))

	return keys, nil
}

// AddKeysFromConfig adds the keys from the config to the smart contracts
func AddKeysFromConfig(ctx map[string]interface{}, cfg config.Configuration) error {
	idSrv := ctx[BootstrappedDIDService].(Service)

	tc, err := configstore.NewAccount(cfg.GetEthereumDefaultAccountName(), cfg)
	if err != nil {
		return err
	}

	tctx, err := contextutil.New(context.Background(), tc)
	if err != nil {
		return err
	}

	keys, err := getKeyPairsFromConfig(cfg)
	if err != nil {
		return err
	}
	err = idSrv.AddKey(tctx, keys[id.KeyPurposeP2P])
	if err != nil {
		return err
	}

	err = idSrv.AddKey(tctx, keys[id.KeyPurposeSigning])
	if err != nil {
		return err
	}

	err = idSrv.AddKey(tctx, keys[id.KeyPurposeEthMsgAuth])
	if err != nil {
		return err
	}
	return nil
}
