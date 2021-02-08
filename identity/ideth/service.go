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
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type contract interface {

	// Ethereum Calls
	GetKey(opts *bind.CallOpts, _key [32]byte) (struct {
		Key       [32]byte
		Purposes  []*big.Int
		RevokedAt uint32
	}, error)

	GetKeysByPurpose(opts *bind.CallOpts, purpose *big.Int) (struct {
		KeysByPurpose [][32]byte
		KeyTypes      []*big.Int
		KeysRevokedAt []uint32
	}, error)

	// Ethereum Transactions
	AddKey(opts *bind.TransactOpts, _key [32]byte, _purpose *big.Int, _keyType *big.Int) (*types.Transaction, error)

	Execute(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error)

	AddMultiPurposeKey(opts *bind.TransactOpts, _key [32]byte, _purposes []*big.Int, _keyType *big.Int) (*types.Transaction, error)

	RevokeKey(opts *bind.TransactOpts, _key [32]byte) (*types.Transaction, error)
}

func methodToOp(method string) config.ContractOp {
	m := map[string]config.ContractOp{
		"mint":         config.NftMint,
		"commit":       config.AnchorCommit,
		"preCommit":    config.AnchorPreCommit,
		"transferFrom": config.NftTransferFrom,
		"store":        config.AssetStore,
		"update":       config.PushToOracle,
	}
	return m[method]
}

type service struct {
	client     ethereum.Client
	jobManager jobs.Manager
	queue      *queue.Server
	config     identity.Config
	dispatcher jobsv2.Dispatcher
}

func (s service) prepareTransaction(ctx context.Context) (contract, *bind.TransactOpts, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, nil, err
	}

	opts, err := s.client.GetTxOpts(ctx, tc.GetEthereumDefaultAccountName())
	if err != nil {
		log.Infof("Failed to get txOpts from Ethereum client: %v", err)
		return nil, nil, err
	}

	contract, err := s.bindContract(identity.NewDID(common.BytesToAddress(tc.GetIdentityID())))
	if err != nil {
		return nil, nil, err
	}

	return contract, opts, nil
}

func (s service) prepareCall(did identity.DID) (contract, *bind.CallOpts, context.CancelFunc, error) {
	opts, cancelFunc := s.client.GetGethCallOpts(false)

	contract, err := s.bindContract(did)
	if err != nil {
		return nil, nil, nil, err
	}

	return contract, opts, cancelFunc, nil
}

func (s service) bindContract(did identity.DID) (contract, error) {
	contract, err := NewIdentityContract(did.ToAddress(), s.client.GetEthClient())
	if err != nil {
		return nil, errors.New("Could not bind identity contract: %v", err)
	}

	return contract, nil
}

// NewService creates a instance of the identity service
func NewService(client ethereum.Client, dispatcher jobsv2.Dispatcher, jobsMgr jobs.Manager, queue *queue.Server,
	conf identity.Config) identity.Service {
	return service{client: client, dispatcher: dispatcher, queue: queue, config: conf, jobManager: jobsMgr}
}

func logTxHash(tx *types.Transaction) {
	log.Infof("Ethereum transaction created. Hash [%x] and Nonce [%v] and Check [%v]", tx.Hash(), tx.Nonce(), tx.CheckNonce())
	log.Infof("Transfer pending: 0x%x\n", tx.Hash())
}

func (s service) submitAndWait(
	did identity.DID, contractMethod interface{}, opts *bind.TransactOpts, params ...interface{}) error {
	ethTX, err := s.client.SubmitTransactionWithRetries(contractMethod, opts, params...)
	if err != nil {
		return fmt.Errorf("failed to submit eth transacion: %w", err)
	}

	// register a random runner and wait for the job and deregister once done
	rfn := hexutil.Encode(utils.RandomSlice(32))
	s.dispatcher.RegisterRunnerFunc(rfn, func([]interface{}, map[string]interface{}) (interface{}, error) {
		return ethereum.IsTxnSuccessful(context.Background(), s.client.GetEthClient(), ethTX.Hash())
	})
	// TODO(ved): need to unregister the runner func
	job := gocelery.NewRunnerFuncJob("", rfn, nil, nil, time.Time{})
	res, err := s.dispatcher.Dispatch(did, job)
	if err != nil {
		return fmt.Errorf("failed to dispatch job: %w", err)
	}

	_, err = res.Await(context.Background())
	return err
}

// AddKey adds a key to identity contract
func (s service) AddKey(ctx context.Context, key identity.Key) error {
	did, err := NewDIDFromContext(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := s.prepareTransaction(ctx)
	if err != nil {
		return err
	}

	opts.GasLimit = s.config.GetEthereumGasLimit(config.IDAddKey)
	log.Info("Add key to identity contract %s", did.ToAddress().String())
	err = s.submitAndWait(did, contract.AddKey, opts, key.GetKey(), key.GetPurpose(), key.GetType())
	if err != nil {
		return fmt.Errorf("failed to add key to contact: %w", err)
	}
	return nil
}

// AddMultiPurposeKey adds a key with multiple purposes
func (s service) AddMultiPurposeKey(ctx context.Context, key [32]byte, purposes []*big.Int, keyType *big.Int) error {
	did, err := NewDIDFromContext(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := s.prepareTransaction(ctx)
	if err != nil {
		return err
	}

	opts.GasLimit = s.config.GetEthereumGasLimit(config.IDAddKey)
	err = s.submitAndWait(did, contract.AddMultiPurposeKey, opts, key, purposes, keyType)
	if err != nil {
		return fmt.Errorf("failed to add multi purpose key: %w", err)
	}
	return nil
}

// RevokeKey revokes an existing key in the smart contract
func (s service) RevokeKey(ctx context.Context, key [32]byte) error {
	did, err := NewDIDFromContext(ctx)
	if err != nil {
		return err
	}

	contract, opts, err := s.prepareTransaction(ctx)
	if err != nil {
		return err
	}

	opts.GasLimit = s.config.GetEthereumGasLimit(config.IDRevokeKey)
	err = s.submitAndWait(did, contract.RevokeKey, opts, key)
	if err != nil {
		return fmt.Errorf("failed tot revoke key: %w", err)
	}

	return nil
}

// ethereumTX is submitting an Ethereum transaction and starts a task to wait for the transaction result
func (s service) ethereumTX(opts *bind.TransactOpts, contractMethod interface{}, params ...interface{}) func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
	return func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
		ethTX, err := s.client.SubmitTransactionWithRetries(contractMethod, opts, params...)
		if err != nil {
			errOut <- err
			return
		}
		logTxHash(ethTX)

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

// GetKey return a key from the identity contract
func (s service) GetKey(did identity.DID, key [32]byte) (*identity.KeyResponse, error) {
	contract, opts, _, err := s.prepareCall(did)
	if err != nil {
		return nil, err
	}

	result, err := contract.GetKey(opts, key)
	if err != nil {
		return nil, errors.New("Could not call identity contract: %v", err)
	}

	return &identity.KeyResponse{Key: result.Key, Purposes: result.Purposes, RevokedAt: result.RevokedAt}, nil
}

// rawExecute calls the execute method on the identity contract
// TODO once we clean up transaction to not use higher level deps we can change back the return to be transactions.txID
func (s service) rawExecute(ctx context.Context, to common.Address, data []byte, gasLimit uint64) (txID identity.IDTX, done chan error, err error) {
	jobID := contextutil.Job(ctx)
	did, err := NewDIDFromContext(ctx)
	if err != nil {
		return jobs.NilJobID(), nil, err
	}
	contract, opts, err := s.prepareTransaction(ctx)
	if err != nil {
		return jobs.NilJobID(), nil, err
	}
	opts.GasLimit = gasLimit

	// default: no ether should be send
	value := big.NewInt(0)
	return s.jobManager.ExecuteWithinJob(contextutil.Copy(ctx), did, jobID, "Check Job for execute", s.ethereumTX(opts, contract.Execute, to, value, data))
}

// Execute creates the abi encoding an calls the execute method on the identity contract
// TODO once we clean up transaction to not use higher level deps we can change back the return to be transactions.txID
func (s service) Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) (txID identity.IDTX, done chan error, err error) {
	abiObj, err := abi.JSON(strings.NewReader(contractAbi))
	if err != nil {
		return jobs.NilJobID(), nil, err
	}

	// Pack encodes the parameters and additionally checks if the method and arguments are defined correctly
	data, err := abiObj.Pack(methodName, args...)
	if err != nil {
		return jobs.NilJobID(), nil, err
	}

	return s.rawExecute(ctx, to, data, s.config.GetEthereumGasLimit(methodToOp(methodName)))
}

func (s service) ExecuteAsync(
	ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) (tx *types.Transaction, err error) {
	abiObj, err := abi.JSON(strings.NewReader(contractAbi))
	if err != nil {
		return nil, err
	}

	// Pack encodes the parameters and additionally checks if the method and arguments are defined correctly
	data, err := abiObj.Pack(methodName, args...)
	if err != nil {
		return nil, err
	}

	contract, opts, err := s.prepareTransaction(ctx)
	if err != nil {
		return nil, err
	}

	opts.GasLimit = s.config.GetEthereumGasLimit(methodToOp(methodName))

	// default: no ether should be send
	value := big.NewInt(0)
	txn, err := s.client.SubmitTransactionWithRetries(contract.Execute, opts, to, value, data)
	if err != nil {
		return nil, err
	}

	return txn, err
}

func (s service) GetKeysByPurpose(did identity.DID, purpose *big.Int) ([]identity.Key, error) {
	contract, opts, _, err := s.prepareCall(did)
	if err != nil {
		return nil, err
	}

	keyStruct, err := contract.GetKeysByPurpose(opts, purpose)
	if err != nil {
		return nil, err
	}

	var keyResp []identity.Key
	for i, k := range keyStruct.KeysByPurpose {
		keyResp = append(keyResp, identity.NewKey(k, purpose, keyStruct.KeyTypes[i], keyStruct.KeysRevokedAt[i]))
	}
	return keyResp, nil
}

// CurrentP2PKey returns the latest P2P key
func (s service) CurrentP2PKey(did identity.DID) (ret string, err error) {
	keys, err := s.GetKeysByPurpose(did, &(identity.KeyPurposeP2PDiscovery.Value))
	if err != nil {
		return ret, err
	}

	if len(keys) == 0 {
		return "", errors.New("missing p2p key")
	}

	lastKey := keys[len(keys)-1]
	key, err := s.GetKey(did, lastKey.GetKey())
	if err != nil {
		return "", err
	}

	if key.RevokedAt != 0 {
		return "", errors.New("current p2p key has been revoked")
	}

	p2pID, err := ed25519.PublicKeyToP2PKey(key.Key)
	if err != nil {
		return ret, err
	}

	return p2pID.Pretty(), nil
}

// GetClientP2PURL returns the p2p url associated with the did
func (s service) GetClientP2PURL(did identity.DID) (string, error) {
	p2pID, err := s.CurrentP2PKey(did)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/ipfs/%s", p2pID), nil
}

// Exists checks if an identity contract exists
func (s service) Exists(ctx context.Context, did identity.DID) error {
	return isIdentityContract(did.ToAddress(), s.client)
}

// ValidateKey checks if a given key is valid for the given centrifugeID.
func (s service) ValidateKey(ctx context.Context, did identity.DID, key []byte, purpose *big.Int, validateAt *time.Time) error {
	contract, opts, _, err := s.prepareCall(did)
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
	if ethKey.RevokedAt > 0 {
		// if a specific time for validation is provided then we validate if a revoked key was revoked before the provided time
		if validateAt != nil {
			revokedAtBlock, err := s.client.GetBlockByNumber(ctx, big.NewInt(int64(ethKey.RevokedAt)))
			if err != nil {
				return err
			}

			if validateAt.Unix() > int64(revokedAtBlock.Time()) {
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
func (s service) GetClientsP2PURLs(dids []*identity.DID) ([]string, error) {
	urls := make([]string, len(dids))

	for idx, did := range dids {
		url, err := s.GetClientP2PURL(*did)
		if err != nil {
			return nil, err
		}
		urls[idx] = url
	}

	return urls, nil
}

// ValidateSignature validates a signature on a message based on identity data
func (s service) ValidateSignature(did identity.DID, pubKey []byte, signature []byte, message []byte, timestamp time.Time) error {
	err := s.ValidateKey(context.Background(), did, pubKey, &(identity.KeyPurposeSigning.Value), &timestamp)
	if err != nil {
		return err
	}

	if !crypto.VerifyMessage(pubKey, message, signature, crypto.CurveSecp256K1) {
		return ErrSignature
	}

	return nil
}

// NewDIDFromContext returns DID from context.Account
func NewDIDFromContext(ctx context.Context) (identity.DID, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return identity.DID{}, err
	}

	addressByte := tc.GetIdentityID()
	return identity.NewDID(common.BytesToAddress(addressByte)), nil
}
