package identity

import (
	"context"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/keytools/ed25519"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("identity")

type factory interface {
	CreateIdentity(opts *bind.TransactOpts, _centrifugeId *big.Int) (*types.Transaction, error)
}

type contract interface {
	AddKey(opts *bind.TransactOpts, _key [32]byte, _kPurpose *big.Int) (*types.Transaction, error)
}

type Config interface {
	GetEthereumDefaultAccountName() string
}

// EthereumIdentityKey holds the identity related details
type EthereumIdentityKey struct {
	Key       [32]byte
	Purposes  []*big.Int
	RevokedAt *big.Int
}

// GetKey returns the public key
func (idk *EthereumIdentityKey) GetKey() [32]byte {
	return idk.Key
}

// GetPurposes returns the purposes intended for the key
func (idk *EthereumIdentityKey) GetPurposes() []*big.Int {
	return idk.Purposes
}

// GetRevokedAt returns the block at which the identity is revoked
func (idk *EthereumIdentityKey) GetRevokedAt() *big.Int {
	return idk.RevokedAt
}

// String prints the peerID extracted from the key
func (idk *EthereumIdentityKey) String() string {
	peerID, _ := ed25519.PublicKeyToP2PKey(idk.Key)
	return fmt.Sprintf("%s", peerID.Pretty())
}

type ethereumIdentity struct {
	CentID           CentID
	Contract         *EthereumIdentityContract
	RegistryContract *EthereumIdentityRegistryContract
	Config           Config
}

func newEthereumIdentity(id CentID, registryContract *EthereumIdentityRegistryContract, config Config) *ethereumIdentity {
	return &ethereumIdentity{CentID: id, RegistryContract: registryContract, Config: config}
}

// CentrifugeID sets the CentID to the Identity
func (id *ethereumIdentity) SetCentrifugeID(centID CentID) {
	id.CentID = centID
}

// String returns CentrifugeID
func (id *ethereumIdentity) String() string {
	return fmt.Sprintf("CentrifugeID [%s]", id.CentID)
}

// CentrifugeID returns the CentrifugeID
func (id *ethereumIdentity) CentrifugeID() CentID {
	return id.CentID
}

// LastKeyForPurpose returns the latest key for given purpose
func (id *ethereumIdentity) LastKeyForPurpose(keyPurpose int) (key []byte, err error) {
	idKeys, err := id.fetchKeysByPurpose(keyPurpose)
	if err != nil {
		return []byte{}, err
	}

	if len(idKeys) == 0 {
		return []byte{}, fmt.Errorf("no key found for type [%d] in ID [%s]", keyPurpose, id.CentID)
	}

	return idKeys[len(idKeys)-1].Key[:32], nil
}

// FetchKey fetches the Key from the chain
func (id *ethereumIdentity) FetchKey(key []byte) (Key, error) {
	contract, err := id.getContract()
	if err != nil {
		return nil, err
	}
	// Ignoring cancelFunc as code will block until response or timeout is triggered
	opts, _ := ethereum.GetGethCallOpts()
	key32, _ := utils.SliceToByte32(key)
	keyInstance, err := contract.GetKey(opts, key32)
	if err != nil {
		return nil, err
	}

	return &EthereumIdentityKey{
		Key:       keyInstance.Key,
		Purposes:  keyInstance.Purposes,
		RevokedAt: keyInstance.RevokedAt,
	}, nil

}

// CurrentP2PKey returns the latest P2P key
func (id *ethereumIdentity) CurrentP2PKey() (ret string, err error) {
	key, err := id.LastKeyForPurpose(KeyPurposeP2P)
	if err != nil {
		return
	}
	key32, _ := utils.SliceToByte32(key)
	p2pId, err := ed25519.PublicKeyToP2PKey(key32)
	if err != nil {
		return
	}
	ret = p2pId.Pretty()
	return
}

func (id *ethereumIdentity) findContract() (exists bool, err error) {
	if id.Contract != nil {
		return true, nil
	}

	// Ignoring cancelFunc as code will block until response or timeout is triggered
	opts, _ := ethereum.GetGethCallOpts()
	idAddress, err := id.RegistryContract.GetIdentityByCentrifugeId(opts, id.CentID.BigInt())
	if err != nil {
		return false, err
	}
	if utils.IsEmptyByteSlice(idAddress.Bytes()) {
		return false, errors.New("Identity not found by address provided")
	}

	client := ethereum.GetConnection()
	idContract, err := NewEthereumIdentityContract(idAddress, client.GetClient())
	if err == bind.ErrNoCode {
		return false, err
	}
	if err != nil {
		log.Errorf("Failed to instantiate the identity contract: %v", err)
		return false, err
	}
	id.Contract = idContract
	return true, nil
}

func (id *ethereumIdentity) getContract() (contract *EthereumIdentityContract, err error) {
	if id.Contract == nil {
		_, err := id.findContract()
		if err != nil {
			return nil, err
		}
		return id.Contract, nil
	}
	return id.Contract, nil
}

// CheckIdentityExists checks if the identity represented by id actually exists on ethereum
func (id *ethereumIdentity) CheckIdentityExists() (exists bool, err error) {
	return id.findContract()
}

// AddKeyToIdentity adds key to the purpose on chain
func (id *ethereumIdentity) AddKeyToIdentity(ctx context.Context, keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error) {
	if utils.IsEmptyByteSlice(key) || len(key) > 32 {
		log.Errorf("Can't add key to identity: empty or invalid length(>32) key for [id: %s]: %x", id, key)
		return confirmations, errors.New("Can't add key to identity: Invalid key")
	}

	ethIdentityContract, err := id.getContract()
	if err != nil {
		log.Errorf("Failed to set up event listener for identity [id: %s]: %v", id, err)
		return confirmations, err
	}

	conn := ethereum.GetConnection()
	opts, err := ethereum.GetConnection().GetTxOpts(id.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return confirmations, err
	}

	h, err := conn.GetClient().HeaderByNumber(ctx, nil)
	if err != nil {
		return confirmations, err
	}

	var keyFixed [32]byte
	copy(keyFixed[:], key)
	confirmations, err = setUpKeyRegisteredEventListener(id, keyPurpose, keyFixed, h.Number.Uint64())
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to set up event listener for identity [id: %s]: %v", id, wError)
		return confirmations, wError
	}

	err = sendKeyRegistrationTransaction(ethIdentityContract, opts, id, keyPurpose, key)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to create transaction for identity [id: %s]: %v", id, wError)
		return confirmations, wError
	}
	return confirmations, nil
}

// fetchKeysByPurpose fetches keys from chain matching purpose
func (id *ethereumIdentity) fetchKeysByPurpose(keyPurpose int) ([]EthereumIdentityKey, error) {
	contract, err := id.getContract()
	if err != nil {
		return nil, err
	}
	// Ignoring cancelFunc as code will block until response or timeout is triggered
	opts, _ := ethereum.GetGethCallOpts()
	bigInt := big.NewInt(int64(keyPurpose))
	keys, err := contract.GetKeysByPurpose(opts, bigInt)
	if err != nil {
		return nil, err
	}
	log.Infof("fetched keys: %d %x\n", keyPurpose, keys)

	var ids []EthereumIdentityKey
	for _, key := range keys {
		ids = append(ids, EthereumIdentityKey{Key: key})
	}

	return ids, nil
}

// sendRegistrationTransaction sends the actual transaction to add a Key on Ethereum registry contract
func sendKeyRegistrationTransaction(identityContract contract, opts *bind.TransactOpts, identity *ethereumIdentity, keyPurpose int, key []byte) (err error) {

	//preparation of data in specific types for the call to Ethereum
	bigInt := big.NewInt(int64(keyPurpose))
	bKey, err := utils.SliceToByte32(key)
	if err != nil {
		return err
	}

	tx, err := ethereum.GetConnection().SubmitTransactionWithRetries(identityContract.AddKey, opts, bKey, bigInt)
	if err != nil {
		log.Infof("Failed to send key [%v:%x] to add to CentrifugeID [%x]: %v", keyPurpose, bKey, identity.CentID, err)
		return err
	}

	log.Infof("Sent off key [%v:%x] to add to CentrifugeID [%s]. Ethereum transaction hash [%x]", keyPurpose, bKey, identity, tx.Hash())
	return
}

// sendIdentityCreationTransaction sends the actual transaction to create identity on Ethereum registry contract
func sendIdentityCreationTransaction(identityFactory factory, opts *bind.TransactOpts, identityToBeCreated Identity) (err error) {
	//preparation of data in specific types for the call to Ethereum
	tx, err := ethereum.GetConnection().SubmitTransactionWithRetries(identityFactory.CreateIdentity, opts, identityToBeCreated.CentrifugeID().BigInt())

	if err != nil {
		log.Infof("Failed to send identity for creation [CentrifugeID: %s] : %v", identityToBeCreated, err)
		return err
	} else {
		log.Infof("Sent off identity creation [%s]. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", identityToBeCreated, tx.Hash(), tx.Nonce(), tx.CheckNonce())
	}

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())

	return
}

// setUpKeyRegisteredEventListener listens for Identity creation
func setUpKeyRegisteredEventListener(identity Identity, keyPurpose int, key [32]byte, bh uint64) (confirmations chan *WatchIdentity, err error) {
	confirmations = make(chan *WatchIdentity)
	centId := identity.CentrifugeID()
	if err != nil {
		return nil, err
	}
	asyncRes, err := queue.Queue.DelayKwargs(keyRegistrationConfirmationTaskName,
		map[string]interface{}{
			centIDParam:      centId,
			keyParam:         key,
			keyPurposeParam:  keyPurpose,
			blockHeightParam: bh,
		})
	if err != nil {
		return nil, err
	}
	go waitAndRouteKeyRegistrationEvent(asyncRes, confirmations, identity)
	return confirmations, nil
}

// setUpRegistrationEventListener sets up the listened for the "IdentityCreated" event to notify the upstream code about successful mining/creation
// of the identity.
func setUpRegistrationEventListener(identityToBeCreated Identity, blockHeight uint64) (confirmations chan *WatchIdentity, err error) {
	confirmations = make(chan *WatchIdentity)
	bCentId := identityToBeCreated.CentrifugeID()
	if err != nil {
		return nil, err
	}
	asyncRes, err := queue.Queue.DelayKwargs(idRegistrationConfirmationTaskName, map[string]interface{}{centIDParam: bCentId, blockHeightParam: blockHeight})
	if err != nil {
		return nil, err
	}
	go waitAndRouteIdentityRegistrationEvent(asyncRes, confirmations, identityToBeCreated)
	return confirmations, nil
}

// waitAndRouteKeyRegistrationEvent notifies the confirmations channel whenever the key has been added to the identity and has been noted as Ethereum event
func waitAndRouteKeyRegistrationEvent(asyncRes *gocelery.AsyncResult, confirmations chan<- *WatchIdentity, pushThisIdentity Identity) {
	_, err := asyncRes.Get(ethereum.GetDefaultContextTimeout())
	confirmations <- &WatchIdentity{Identity: pushThisIdentity, Error: err}
}

// waitAndRouteIdentityRegistrationEvent notifies the confirmations channel whenever the identity creation is being noted as Ethereum event
func waitAndRouteIdentityRegistrationEvent(asyncRes *gocelery.AsyncResult, confirmations chan<- *WatchIdentity, pushThisIdentity Identity) {
	_, err := asyncRes.Get(ethereum.GetDefaultContextTimeout())
	confirmations <- &WatchIdentity{pushThisIdentity, err}
}

// EthereumidentityService implements `Service`
type EthereumIdentityService struct {
	config           Config
	factoryContract  *EthereumIdentityFactoryContract
	registryContract *EthereumIdentityRegistryContract
}

// NewEthereumIdentityService creates a new NewEthereumIdentityService given the config and the smart contracts
func NewEthereumIdentityService(config Config, factoryContract *EthereumIdentityFactoryContract, registryContract *EthereumIdentityRegistryContract) Service {
	return &EthereumIdentityService{config: config, factoryContract: factoryContract, registryContract: registryContract}
}

// CheckIdentityExists checks if the identity represented by id actually exists on ethereum
func (ids *EthereumIdentityService) CheckIdentityExists(centrifugeID CentID) (exists bool, err error) {
	id := newEthereumIdentity(centrifugeID, ids.registryContract, ids.config)
	exists, err = id.CheckIdentityExists()
	return
}

// CreateIdentity creates an identity representing the id on ethereum
func (ids *EthereumIdentityService) CreateIdentity(centrifugeID CentID) (id Identity, confirmations chan *WatchIdentity, err error) {
	log.Infof("Creating Identity [%x]", centrifugeID)
	id = newEthereumIdentity(centrifugeID, ids.registryContract, ids.config)
	conn := ethereum.GetConnection()
	opts, err := conn.GetTxOpts(ids.config.GetEthereumDefaultAccountName())
	if err != nil {
		return nil, confirmations, err
	}

	h, err := conn.GetClient().HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, confirmations, err
	}

	confirmations, err = setUpRegistrationEventListener(id, h.Number.Uint64())
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to set up event listener for identity [mockID: %s]: %v", id, wError)
		return nil, confirmations, wError
	}

	err = sendIdentityCreationTransaction(ids.factoryContract, opts, id)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to create transaction for identity [mockID: %s]: %v", id, wError)
		return nil, confirmations, wError
	}
	return id, confirmations, nil
}

// GetIdentityAddress gets the address of the ethereum identity contract for the given CentID
func (ids *EthereumIdentityService) GetIdentityAddress(centID CentID) (common.Address, error) {
	// Ignoring cancelFunc as code will block until response or timeout is triggered
	opts, _ := ethereum.GetGethCallOpts()
	address, err := ids.registryContract.GetIdentityByCentrifugeId(opts, centID.BigInt())
	if err != nil {
		return common.Address{}, err
	}

	if utils.IsEmptyAddress(address) {
		return common.Address{}, errors.New("No address found for centID")
	}
	return address, nil
}

// LookupIdentityForID looks up if the identity for given CentID exists on ethereum
func (ids *EthereumIdentityService) LookupIdentityForID(centrifugeID CentID) (Identity, error) {
	id := newEthereumIdentity(centrifugeID, ids.registryContract, ids.config)
	exists, err := id.CheckIdentityExists()
	if !exists {
		return id, fmt.Errorf("identity [%s] does not exist with err [%v]", id.CentID, err)
	}

	if err != nil {
		return id, err
	}

	return id, nil
}
