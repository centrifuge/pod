package identity

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("identity")

type WatchIdentityCreated interface {
	WatchIdentityCreated(opts *bind.WatchOpts, sink chan<- *EthereumIdentityFactoryContractIdentityCreated, centrifugeId [][32]byte) (event.Subscription, error)
}

type WatchKeyRegistered interface {
	WatchKeyRegistered(opts *bind.WatchOpts, sink chan<- *EthereumIdentityContractKeyRegistered, kType []*big.Int, key [][32]byte) (event.Subscription, error)
}

type IdentityFactory interface {
	CreateIdentity(opts *bind.TransactOpts, _centrifugeId [32]byte) (*types.Transaction, error)
}

type IdentityContract interface {
	AddKey(opts *bind.TransactOpts, _key [32]byte, _kType *big.Int) (*types.Transaction, error)
}

type EthereumIdentityKey struct {
	Key [32]byte
}

func (idk *EthereumIdentityKey) String() string {
	peerdId, _ := keytools.PublicKeyToP2PKey(idk.Key)
	return fmt.Sprintf("%s", peerdId.Pretty())
}

func NewEthereumIdentity() (id *EthereumIdentity) {
	id = new(EthereumIdentity)
	id.cachedKeys = make(map[int][]EthereumIdentityKey)
	return
}

type EthereumIdentity struct {
	CentrifugeId []byte
	cachedKeys   map[int][]EthereumIdentityKey
	Contract     *EthereumIdentityContract
}

func (id *EthereumIdentity) SetCentrifugeId(b []byte) error {
	if len(b) != 32 {
		return errors.New("CentrifugeId has incorrect length")
	}
	if tools.IsEmptyByteSlice(b) {
		return errors.New("CentrifugeId can't be empty")
	}
	id.CentrifugeId = b
	return nil
}

func (id *EthereumIdentity) CentrifugeIdString() string {
	return base64.StdEncoding.EncodeToString(id.CentrifugeId)
}

func (id *EthereumIdentity) CentrifugeIdB32() [32]byte {
	var b32Id [32]byte
	copy(b32Id[:], id.CentrifugeId[:32])
	return b32Id
}

func (id *EthereumIdentity) String() string {
	return fmt.Sprintf("CentrifugeId [%s]", id.CentrifugeIdString())
}

func (id *EthereumIdentity) GetCentrifugeId() []byte {
	return id.CentrifugeId
}

func (id *EthereumIdentity) GetLastKeyForType(keyType int) (key []byte, err error) {
	err = id.fetchKeysByType(keyType)
	if err != nil {
		return
	}

	if len(id.cachedKeys[keyType]) == 0 {
		return []byte{}, fmt.Errorf("No key found for type [%d] in id [%s]", keyType, id.CentrifugeIdString())
	}

	return id.cachedKeys[keyType][len(id.cachedKeys[keyType])-1].Key[:32], nil
}
func (id *EthereumIdentity) GetCurrentP2PKey() (ret string, err error) {
	key, err := id.GetLastKeyForType(KEY_TYPE_PEERID)
	if err != nil {
		return
	}
	key32, _ := tools.SliceToByte32(key)
	p2pId, err := keytools.PublicKeyToP2PKey(key32)
	if err != nil {
		return
	}
	ret = p2pId.Pretty()
	return
}

func (id *EthereumIdentity) findContract() (exists bool, err error) {
	if id.Contract != nil {
		return true, nil
	}

	ethIdentityRegistryContract, err := getIdentityRegistryContract()
	if err != nil {
		return false, err
	}
	opts := ethereum.GetGethCallOpts()
	idAddress, err := ethIdentityRegistryContract.GetIdentityByCentrifugeId(opts, id.CentrifugeIdB32())
	if err != nil {
		return false, err
	}
	if tools.IsEmptyByteSlice(idAddress.Bytes()) {
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

func (id *EthereumIdentity) getContract() (contract *EthereumIdentityContract, err error) {
	if id.Contract == nil {
		_, err := id.findContract()
		if err != nil {
			return nil, err
		}
		return id.Contract, nil
	}
	return id.Contract, nil
}

func (id *EthereumIdentity) CheckIdentityExists() (exists bool, err error) {
	exists, err = id.findContract()
	return
}

func (id *EthereumIdentity) AddKeyToIdentity(keyType int, key []byte) (confirmations chan *WatchIdentity, err error) {
	if tools.IsEmptyByteSlice(key) || len(key) != 32 {
		return confirmations, errors.New("Can't add key to identity: Inavlid key")
	}

	ethIdentityContract, err := id.getContract()
	if err != nil {
		return
	}

	confirmations, err = setUpKeyRegisteredEventListener(ethIdentityContract, id, keyType, key)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to set up event listener for identity [id: %s]: %v", id, wError)
		return
	}

	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return confirmations, err
	}

	err = sendKeyRegistrationTransaction(ethIdentityContract, opts, id, keyType, key)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to create transaction for identity [id: %s]: %v", id, wError)
		return confirmations, wError
	}
	return confirmations, nil
}

func (id *EthereumIdentity) fetchKeysByType(keyType int) error {
	contract, err := id.getContract()
	if err != nil {
		return err
	}
	opts := ethereum.GetGethCallOpts()
	bigInt := big.NewInt(int64(keyType))
	keys, err := contract.GetKeysByType(opts, bigInt)
	if err != nil {
		return err
	}
	log.Errorf("HERE: %d %x\n", keyType, keys)
	for _, key := range keys {
		id.cachedKeys[keyType] = append(id.cachedKeys[keyType], EthereumIdentityKey{key})
	}
	return nil
}

func getIdentityFactoryContract() (identityFactoryContract *EthereumIdentityFactoryContract, err error) {
	client := ethereum.GetConnection()

	identityFactoryContract, err = NewEthereumIdentityFactoryContract(config.Config.GetContractAddress("identityFactory"), client.GetClient())
	if err != nil {
		log.Infof("Failed to instantiate the identity factory contract: %v", err)
	}
	return
}

func getIdentityRegistryContract() (identityRegistryContract *EthereumIdentityRegistryContract, err error) {
	client := ethereum.GetConnection()

	identityRegistryContract, err = NewEthereumIdentityRegistryContract(config.Config.GetContractAddress("identityRegistry"), client.GetClient())
	if err != nil {
		log.Infof("Failed to instantiate the identity registry contract: %v", err)
	}
	return
}

// sendRegistrationTransaction sends the actual transaction to add a Key on Ethereum registry contract
func sendKeyRegistrationTransaction(identityContract IdentityContract, opts *bind.TransactOpts, identity *EthereumIdentity, keyType int, key []byte) (err error) {

	//preparation of data in specific types for the call to Ethereum
	bigInt := big.NewInt(int64(keyType))
	bKey, err := tools.SliceToByte32(key)
	if err != nil {
		return err
	}

	tx, err := ethereum.SubmitTransactionWithRetries(identityContract.AddKey, opts, bKey, bigInt)
	if err != nil {
		log.Infof("Failed to send key [%v:%x] to add to CentrifugeID [%x]: %v", keyType, bKey, identity.CentrifugeId, err)
		return err
	}

	log.Infof("Sent off key [%v:%x] to add to CentrifugeID [%s]. Ethereum transaction hash [%x]", keyType, bKey, identity, tx.Hash())
	return
}

// sendIdentityCreationTransaction sends the actual transaction to create identity on Ethereum registry contract
func sendIdentityCreationTransaction(identityFactory IdentityFactory, opts *bind.TransactOpts, identityToBeCreated Identity) (err error) {
	//preparation of data in specific types for the call to Ethereum
	tx, err := ethereum.SubmitTransactionWithRetries(identityFactory.CreateIdentity, opts, identityToBeCreated.CentrifugeIdB32())

	if err != nil {
		log.Infof("Failed to send identity for creation [CentrifugeID: %s] : %v", identityToBeCreated, err)
		return err
	} else {
		log.Infof("Sent off identity creation [%s]. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", identityToBeCreated, tx.Hash(), tx.Nonce(), tx.CheckNonce())
	}

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())

	return
}

func setUpKeyRegisteredEventListener(ethCreatedContract WatchKeyRegistered, identity *EthereumIdentity, keyType int, key []byte) (confirmations chan *WatchIdentity, err error) {
	//listen to this particular key being mined/event is triggered
	ctx, cancelFunc := ethereum.DefaultWaitForTransactionMiningContext()
	watchOpts := &bind.WatchOpts{Context: ctx}

	//only setting up a channel of 1 notification as there should always be only one notification coming for this
	//single key being registered
	keyAddedEvents := make(chan *EthereumIdentityContractKeyRegistered, 1)
	confirmations = make(chan *WatchIdentity)
	go waitAndRouteKeyRegistrationEvent(keyAddedEvents, watchOpts.Context, confirmations, identity)

	b32Key, err := tools.SliceToByte32(key)
	if err != nil {
		return confirmations, err
	}
	bigInt := big.NewInt(int64(keyType))

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = ethCreatedContract.WatchKeyRegistered(watchOpts, keyAddedEvents, []*big.Int{bigInt}, [][32]byte{b32Key})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Errorf(wError.Error())
		cancelFunc()
		return confirmations, wError
	}
	return
}

// setUpRegistrationEventListener sets up the listened for the "IdentityCreated" event to notify the upstream code about successful mining/creation
// of the identity.
func setUpRegistrationEventListener(ethCreatedContract WatchIdentityCreated, identityToBeCreated Identity) (confirmations chan *WatchIdentity, err error) {
	//listen to this particular identity being mined/event is triggered
	ctx, cancelFunc := ethereum.DefaultWaitForTransactionMiningContext()
	watchOpts := &bind.WatchOpts{Context: ctx}

	//only setting up a channel of 1 notification as there should always be only one notification coming for this
	//single identity being registered
	identityCreatedEvents := make(chan *EthereumIdentityFactoryContractIdentityCreated, 1)
	confirmations = make(chan *WatchIdentity)
	go waitAndRouteIdentityRegistrationEvent(identityCreatedEvents, watchOpts.Context, confirmations, identityToBeCreated)

	bCentId := identityToBeCreated.CentrifugeIdB32()

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = ethCreatedContract.WatchIdentityCreated(watchOpts, identityCreatedEvents, [][32]byte{bCentId})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Errorf(wError.Error())
		cancelFunc()
	}
	return
}

// waitAndRouteKeyRegistrationEvent notifies the confirmations channel whenever the key has been added to the identity and has been noted as Ethereum event
func waitAndRouteKeyRegistrationEvent(conf <-chan *EthereumIdentityContractKeyRegistered, ctx context.Context, confirmations chan<- *WatchIdentity, pushThisIdentity Identity) {
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving KeyRegistered event for Identity ID: %x\n", ctx, pushThisIdentity)
			confirmations <- &WatchIdentity{pushThisIdentity, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received KeyRegistered event from [%s] for keyType: %x and value: %x\n", pushThisIdentity, res.KType, res.Key)
			confirmations <- &WatchIdentity{pushThisIdentity, nil}
			return
		}
	}
}

// waitAndRouteIdentityRegistrationEvent notifies the confirmations channel whenever the identity creation is being noted as Ethereum event
func waitAndRouteIdentityRegistrationEvent(conf <-chan *EthereumIdentityFactoryContractIdentityCreated, ctx context.Context, confirmations chan<- *WatchIdentity, pushThisIdentity Identity) {
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving IdentityCreated event for Identity ID: %x\n", ctx, pushThisIdentity)
			confirmations <- &WatchIdentity{pushThisIdentity, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received IdentityCreated event from: %x, identifier: %x\n", res.CentrifugeId, res.Identity)
			confirmations <- &WatchIdentity{pushThisIdentity, nil}
			return
		}
	}
}

func NewEthereumIdentityService() IdentityService {
	return &EthereumIdentityService{}
}

// EthereumidentityService implements `IdentityService`
type EthereumIdentityService struct {
}

func (ids *EthereumIdentityService) CheckIdentityExists(centrifugeId []byte) (exists bool, err error) {
	if tools.IsEmptyByteSlice(centrifugeId) || len(centrifugeId) != 32 {
		return false, errors.New("centrifugeId empty or of incorrect length")
	}
	id := NewEthereumIdentity()
	id.CentrifugeId = centrifugeId
	exists, err = id.CheckIdentityExists()
	return
}

func (ids *EthereumIdentityService) CreateIdentity(centrifugeId []byte) (id Identity, confirmations chan *WatchIdentity, err error) {
	log.Infof("Creating Identity [%v]", centrifugeId)

	id = NewEthereumIdentity()
	id.SetCentrifugeId(centrifugeId)

	ethIdentityFactoryContract, err := getIdentityFactoryContract()
	if err != nil {
		return
	}
	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return nil, confirmations, err
	}

	confirmations, err = setUpRegistrationEventListener(ethIdentityFactoryContract, id)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to set up event listener for identity [id: %s]: %v", id, wError)
		return nil, confirmations, wError
	}

	err = sendIdentityCreationTransaction(ethIdentityFactoryContract, opts, id)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to create transaction for identity [id: %s]: %v", id, wError)
		return nil, confirmations, wError
	}
	return id, confirmations, nil
}

func (ids *EthereumIdentityService) LookupIdentityForId(centrifugeId []byte) (Identity, error) {
	instanceId := NewEthereumIdentity()
	err := instanceId.SetCentrifugeId(centrifugeId)
	if err != nil {
		return instanceId, err
	}

	exists, err := instanceId.CheckIdentityExists()

	if !exists {
		return instanceId, fmt.Errorf("Identity [%s] does not exist", instanceId.CentrifugeIdString())
	}

	if err != nil {
		return instanceId, err
	}

	return instanceId, nil
}
