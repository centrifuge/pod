package identity

import (
	"context"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/centrifuge/go-centrifuge/centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("identity")

type WatchKeyAdded interface {
	WatchKeyAdded(opts *bind.WatchOpts, sink chan<- *EthereumIdentityContractKeyAdded, key [][32]byte, purpose []*big.Int) (event.Subscription, error)
}

type IdentityFactory interface {
	CreateIdentity(opts *bind.TransactOpts, _centrifugeId *big.Int) (*types.Transaction, error)
}

type IdentityContract interface {
	AddKey(opts *bind.TransactOpts, _key [32]byte, _kPurpose *big.Int) (*types.Transaction, error)
}

type EthereumIdentityKey struct {
	Key       [32]byte
	Purposes  []*big.Int
	RevokedAt *big.Int
}

func (idk *EthereumIdentityKey) GetKey() [32]byte {
	return idk.Key
}

func (idk *EthereumIdentityKey) GetPurposes() []*big.Int {
	return idk.Purposes
}

func (idk *EthereumIdentityKey) GetRevokedAt() *big.Int {
	return idk.RevokedAt
}

func (idk *EthereumIdentityKey) String() string {
	peerID, _ := ed25519.PublicKeyToP2PKey(idk.Key)
	return fmt.Sprintf("%s", peerID.Pretty())
}

type EthereumIdentity struct {
	CentrifugeId CentID
	Contract     *EthereumIdentityContract
}

func (id *EthereumIdentity) CentrifugeID(cenId CentID) {
	id.CentrifugeId = cenId
}

func (id *EthereumIdentity) CentrifugeIDBytes() CentID {
	var idBytes [CentIDByteLength]byte
	copy(idBytes[:], id.CentrifugeId[:CentIDByteLength])
	return idBytes
}

func (id *EthereumIdentity) String() string {
	return fmt.Sprintf("CentrifugeID [%s]", id.CentrifugeId)
}

func (id *EthereumIdentity) GetCentrifugeID() CentID {
	return id.CentrifugeId
}

func (id *EthereumIdentity) GetLastKeyForPurpose(keyPurpose int) (key []byte, err error) {
	idKeys, err := id.fetchKeysByPurpose(keyPurpose)
	if err != nil {
		return []byte{}, err
	}

	if len(idKeys) == 0 {
		return []byte{}, fmt.Errorf("no key found for type [%d] in ID [%s]", keyPurpose, id.CentrifugeId)
	}

	return idKeys[len(idKeys)-1].Key[:32], nil
}

func (id *EthereumIdentity) FetchKey(key []byte) (Key, error) {
	contract, err := id.getContract()
	if err != nil {
		return nil, err
	}
	opts := ethereum.GetGethCallOpts()
	key32, _ := tools.SliceToByte32(key)
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

func (id *EthereumIdentity) GetCurrentP2PKey() (ret string, err error) {
	key, err := id.GetLastKeyForPurpose(KeyPurposeP2p)
	if err != nil {
		return
	}
	key32, _ := tools.SliceToByte32(key)
	p2pId, err := ed25519.PublicKeyToP2PKey(key32)
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
	idAddress, err := ethIdentityRegistryContract.GetIdentityByCentrifugeId(opts, id.CentrifugeId.BigInt())
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
	return id.findContract()
}

func (id *EthereumIdentity) AddKeyToIdentity(keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error) {
	if tools.IsEmptyByteSlice(key) || len(key) > 32 {
		log.Errorf("Can't add key to identity: empty or invalid length(>32) key for [id: %s]: %x", id, key)
		return confirmations, errors.New("Can't add key to identity: Invalid key")
	}

	ethIdentityContract, err := id.getContract()
	if err != nil {
		log.Errorf("Failed to set up event listener for identity [id: %s]: %v", id, err)
		return confirmations, err
	}

	confirmations, err = setUpKeyRegisteredEventListener(ethIdentityContract, id, keyPurpose, key)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to set up event listener for identity [id: %s]: %v", id, wError)
		return confirmations, wError
	}

	opts, err := ethereum.GetConnection().GetTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return confirmations, err
	}

	err = sendKeyRegistrationTransaction(ethIdentityContract, opts, id, keyPurpose, key)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Errorf("Failed to create transaction for identity [id: %s]: %v", id, wError)
		return confirmations, wError
	}
	return confirmations, nil
}

func (id *EthereumIdentity) fetchKeysByPurpose(keyPurpose int) ([]EthereumIdentityKey, error) {
	contract, err := id.getContract()
	if err != nil {
		return nil, err
	}
	opts := ethereum.GetGethCallOpts()
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
func sendKeyRegistrationTransaction(identityContract IdentityContract, opts *bind.TransactOpts, identity *EthereumIdentity, keyPurpose int, key []byte) (err error) {

	//preparation of data in specific types for the call to Ethereum
	bigInt := big.NewInt(int64(keyPurpose))
	bKey, err := tools.SliceToByte32(key)
	if err != nil {
		return err
	}

	tx, err := ethereum.SubmitTransactionWithRetries(identityContract.AddKey, opts, bKey, bigInt)
	if err != nil {
		log.Infof("Failed to send key [%v:%x] to add to CentrifugeID [%x]: %v", keyPurpose, bKey, identity.CentrifugeId, err)
		return err
	}

	log.Infof("Sent off key [%v:%x] to add to CentrifugeID [%s]. Ethereum transaction hash [%x]", keyPurpose, bKey, identity, tx.Hash())
	return
}

// sendIdentityCreationTransaction sends the actual transaction to create identity on Ethereum registry contract
func sendIdentityCreationTransaction(identityFactory IdentityFactory, opts *bind.TransactOpts, identityToBeCreated Identity) (err error) {
	//preparation of data in specific types for the call to Ethereum
	tx, err := ethereum.SubmitTransactionWithRetries(identityFactory.CreateIdentity, opts, identityToBeCreated.GetCentrifugeID().BigInt())

	if err != nil {
		log.Infof("Failed to send identity for creation [CentrifugeID: %s] : %v", identityToBeCreated, err)
		return err
	} else {
		log.Infof("Sent off identity creation [%s]. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", identityToBeCreated, tx.Hash(), tx.Nonce(), tx.CheckNonce())
	}

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())

	return
}

func setUpKeyRegisteredEventListener(ethCreatedContract WatchKeyAdded, identity *EthereumIdentity, keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error) {
	//listen to this particular key being mined/event is triggered
	ctx, cancelFunc := ethereum.DefaultWaitForTransactionMiningContext()
	watchOpts := &bind.WatchOpts{Context: ctx}

	// there should always be only one notification coming for this
	// single key being registered
	keyAddedEvents := make(chan *EthereumIdentityContractKeyAdded)
	confirmations = make(chan *WatchIdentity)

	b32Key, err := tools.SliceToByte32(key)
	if err != nil {
		return confirmations, err
	}
	keyPurposeInt := big.NewInt(int64(keyPurpose))

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	subscription, err := ethCreatedContract.WatchKeyAdded(watchOpts, keyAddedEvents, [][32]byte{b32Key}, []*big.Int{keyPurposeInt})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Errorf(wError.Error())
		cancelFunc()
		return confirmations, wError
	}
	go waitAndRouteKeyRegistrationEvent(keyAddedEvents, subscription, watchOpts.Context, confirmations, identity)
	return
}

// setUpRegistrationEventListener sets up the listened for the "IdentityCreated" event to notify the upstream code about successful mining/creation
// of the identity.
func setUpRegistrationEventListener(identityToBeCreated Identity) (confirmations chan *WatchIdentity, err error) {
	confirmations = make(chan *WatchIdentity)
	bCentId := identityToBeCreated.GetCentrifugeID()
	if err != nil {
		return nil, err
	}
	asyncRes, err := queue.Queue.DelayKwargs(IdRegistrationConfirmationTaskName, map[string]interface{}{CentIdParam: bCentId})
	if err != nil {
		return nil, err
	}
	go waitAndRouteIdentityRegistrationEvent(asyncRes, confirmations, identityToBeCreated)
	return confirmations, nil
}

// waitAndRouteKeyRegistrationEvent notifies the confirmations channel whenever the key has been added to the identity and has been noted as Ethereum event
func waitAndRouteKeyRegistrationEvent(conf <-chan *EthereumIdentityContractKeyAdded, subscription event.Subscription, ctx context.Context, confirmations chan<- *WatchIdentity, pushThisIdentity Identity) {
	for {
		select {
		case err := <-subscription.Err():
			log.Errorf("Subscription error %s", err.Error())
			return
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving KeyRegistered event for Identity ID: %x\n", ctx, pushThisIdentity)
			confirmations <- &WatchIdentity{pushThisIdentity, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received KeyRegistered event from [%s] for keyPurpose: %x and value: %x\n", pushThisIdentity, res.Purpose, res.Key)
			confirmations <- &WatchIdentity{pushThisIdentity, nil}
			return
		}
	}
}

// waitAndRouteIdentityRegistrationEvent notifies the confirmations channel whenever the identity creation is being noted as Ethereum event
func waitAndRouteIdentityRegistrationEvent(asyncRes *gocelery.AsyncResult, confirmations chan<- *WatchIdentity, pushThisIdentity Identity) {
	_, err := asyncRes.Get(ethereum.GetDefaultContextTimeout())
	confirmations <- &WatchIdentity{pushThisIdentity, err}
}

func NewEthereumIdentityService() Service {
	return &EthereumIdentityService{}
}

// EthereumidentityService implements `Service`
type EthereumIdentityService struct{}

func (ids *EthereumIdentityService) CheckIdentityExists(centrifugeID CentID) (exists bool, err error) {
	id := new(EthereumIdentity)
	id.CentrifugeId = centrifugeID
	exists, err = id.CheckIdentityExists()
	return
}

func (ids *EthereumIdentityService) CreateIdentity(centrifugeID CentID) (id Identity, confirmations chan *WatchIdentity, err error) {
	log.Infof("Creating Identity [%x]", centrifugeID.ByteArray())

	id = new(EthereumIdentity)
	id.CentrifugeID(centrifugeID)

	ethIdentityFactoryContract, err := getIdentityFactoryContract()
	if err != nil {
		return
	}
	opts, err := ethereum.GetConnection().GetTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return nil, confirmations, err
	}

	confirmations, err = setUpRegistrationEventListener(id)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to set up event listener for identity [mockID: %s]: %v", id, wError)
		return nil, confirmations, wError
	}

	err = sendIdentityCreationTransaction(ethIdentityFactoryContract, opts, id)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to create transaction for identity [mockID: %s]: %v", id, wError)
		return nil, confirmations, wError
	}
	return id, confirmations, nil
}

func (ids *EthereumIdentityService) LookupIdentityForID(centrifugeID CentID) (Identity, error) {
	id := new(EthereumIdentity)
	id.CentrifugeID(centrifugeID)
	exists, err := id.CheckIdentityExists()
	if !exists {
		return id, fmt.Errorf("identity [%s] does not exist", id.CentrifugeId)
	}

	if err != nil {
		return id, err
	}

	return id, nil
}
