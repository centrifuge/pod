package identity

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
	logging "github.com/ipfs/go-log"
	"math/big"
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

type EthereumIdentity struct {
	CentrifugeId []byte
	Keys         map[int][]EthereumIdentityKey
	Contract     *EthereumIdentityContract
}

func NewEthereumIdentity() (id *EthereumIdentity) {
	id = new(EthereumIdentity)
	id.Keys = make(map[int][]EthereumIdentityKey)
	return
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
	joinedKeys := ""
	for k, v := range id.Keys {
		for i, _ := range v {
			joinedKeys += fmt.Sprintf("[%v]%s ", k, v[i].String())
		}
	}
	return fmt.Sprintf("CentrifugeId [%s], Keys [%s]", id.CentrifugeIdString(), joinedKeys)
}

func (id *EthereumIdentity) GetCentrifugeId() []byte {
	return id.CentrifugeId
}

func (id *EthereumIdentity) GetLastB58KeyForType(keyType int) (ret string, err error) {
	if len(id.Keys[keyType]) == 0 {
		return
	}
	switch keyType {
	case 0:
		log.Infof("Error not authorized type")
	case 1:
		p2pId, err1 := keytools.PublicKeyToP2PKey(id.Keys[keyType][len(id.Keys[keyType])-1].Key)
		if err1 != nil {
			err = err1
			return
		}
		ret = p2pId.Pretty()
	default:
		log.Infof("keyType not found")
	}
	return
}

func (id *EthereumIdentity) findContract() (exists bool, err error) {
	if id.Contract != nil {
		return true, nil
	}

	ethIdentityRegistryContract, err := getIdentityRegistryContract()
	if err != nil {
		return
	}
	opts := ethereum.GetGethCallOpts()
	idAddress, err := ethIdentityRegistryContract.GetIdentityByCentrifugeId(opts, id.CentrifugeIdB32())
	if err != nil {
		return false, err
	}

	client := ethereum.GetConnection()
	id.Contract, err = NewEthereumIdentityContract(idAddress, client.GetClient())
	if err == bind.ErrNoCode {
		return false, err
	}
	if err != nil {
		log.Errorf("Failed to instantiate the identity contract: %v", err)
		return false, err
	}
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

func (id *EthereumIdentity) AddKeyToIdentity(keyType int, confirmations chan<- *WatchIdentity) (err error) {
	ethIdentityContract, err := id.getContract()
	if err != nil {
		return
	}

	err = setUpKeyRegisteredEventListener(ethIdentityContract, id, keyType, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to set up event listener for identity [id: %x]: %v", id.CentrifugeIdString(), wError)
		return
	}

	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return err
	}

	err = sendKeyRegistrationTransaction(ethIdentityContract, opts, id, keyType)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to create transaction for identity [id: %x]: %v", id.CentrifugeIdString(), wError)
		return
	}
	return
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

func CreateEthereumIdentity(identity *EthereumIdentity, confirmations chan<- *WatchIdentity) (err error) {
	err = tools.CheckBytesLen32(identity.CentrifugeId, "centrifugeId needs to be length of 32. Got value [%v]")
	if err != nil {
		return
	}
	ethIdentityFactoryContract, err := getIdentityFactoryContract()
	if err != nil {
		return
	}
	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return err
	}

	err = setUpRegistrationEventListener(ethIdentityFactoryContract, identity, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to set up event listener for identity [id: %x]: %v", identity.CentrifugeIdString(), wError)
		return
	}

	err = sendIdentityCreationTransaction(ethIdentityFactoryContract, opts, identity)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to create transaction for identity [id: %x]: %v", identity.CentrifugeIdString(), wError)
		return
	}
	return
}

// sendRegistrationTransaction sends the actual transaction to add a Key on Ethereum registry contract
func sendKeyRegistrationTransaction(identityContract IdentityContract, opts *bind.TransactOpts, identity *EthereumIdentity, keyType int) (err error) {

	//preparation of data in specific types for the call to Ethereum
	lastKey := len(identity.Keys[keyType]) - 1
	bigInt := big.NewInt(int64(keyType))
	bKey := identity.Keys[keyType][lastKey].Key
	tx, err := ethereum.SubmitTransactionWithRetries(identityContract.AddKey, opts, bKey, bigInt)

	if err != nil {
		log.Infof("Failed to send key [%v:%x] to add to CentrifugeID [%x]: %v", keyType, bKey, identity.CentrifugeId, err)
		return err
	} else {
		log.Infof("Sent off key [%v:%x] to add to CentrifugeID [%x]. Ethereum transaction hash [%x]", keyType, bKey, identity.CentrifugeId, tx.Hash())
	}

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())

	return
}

// sendIdentityCreationTransaction sends the actual transaction to create identity on Ethereum registry contract
func sendIdentityCreationTransaction(identityFactory IdentityFactory, opts *bind.TransactOpts, identityToBeCreated *EthereumIdentity) (err error) {
	//preparation of data in specific types for the call to Ethereum
	tx, err := ethereum.SubmitTransactionWithRetries(identityFactory.CreateIdentity, opts, identityToBeCreated.CentrifugeIdB32())

	if err != nil {
		log.Infof("Failed to send identity for creation [CentrifugeID: %x] : %v", identityToBeCreated.CentrifugeIdString(), err)
		return err
	} else {
		log.Infof("Sent off identity creation [CentrifugeID: %x]. Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", identityToBeCreated.CentrifugeIdString(), tx.Hash(), tx.Nonce(), tx.CheckNonce())
	}

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())

	return
}

func setUpKeyRegisteredEventListener(ethCreatedContract WatchKeyRegistered, identity *EthereumIdentity, keyType int, confirmations chan<- *WatchIdentity) (err error) {
	//listen to this particular key being mined/event is triggered
	watchOpts := &bind.WatchOpts{}
	watchOpts.Context = ethereum.DefaultWaitForTransactionMiningContext()

	//only setting up a channel of 1 notification as there should always be only one notification coming for this
	//single key being registered
	keyAddedEvents := make(chan *EthereumIdentityContractKeyRegistered, 1)
	go waitAndRouteKeyRegistrationEvent(keyAddedEvents, watchOpts.Context, confirmations, identity)

	lastKey := len(identity.Keys[keyType]) - 1
	bKey := identity.Keys[keyType][lastKey].Key
	bigInt := big.NewInt(int64(keyType))

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.

	_, err = ethCreatedContract.WatchKeyRegistered(watchOpts, keyAddedEvents, []*big.Int{bigInt}, [][32]byte{bKey})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Errorf(wError.Error())
	}
	return
}

// setUpRegistrationEventListener sets up the listened for the "IdentityCreated" event to notify the upstream code about successful mining/creation
// of the identity.
func setUpRegistrationEventListener(ethCreatedContract WatchIdentityCreated, identityToBeCreated *EthereumIdentity, confirmations chan<- *WatchIdentity) (err error) {

	//listen to this particular identity being mined/event is triggered
	watchOpts := &bind.WatchOpts{}
	watchOpts.Context = ethereum.DefaultWaitForTransactionMiningContext()

	//only setting up a channel of 1 notification as there should always be only one notification coming for this
	//single identity being registered
	identityCreatedEvents := make(chan *EthereumIdentityFactoryContractIdentityCreated, 1)
	go waitAndRouteIdentityRegistrationEvent(identityCreatedEvents, watchOpts.Context, confirmations, identityToBeCreated)

	bCentId := identityToBeCreated.CentrifugeIdB32()

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = ethCreatedContract.WatchIdentityCreated(watchOpts, identityCreatedEvents, [][32]byte{bCentId})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Errorf(wError.Error())
	}
	return
}

// waitAndRouteKeyRegistrationEvent notifies the confirmations channel whenever the key has been added to the identity and has been noted as Ethereum event
func waitAndRouteKeyRegistrationEvent(conf <-chan *EthereumIdentityContractKeyRegistered, ctx context.Context, confirmations chan<- *WatchIdentity, pushThisIdentity *EthereumIdentity) {
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving KeyRegistered event for Identity ID: %x\n", ctx, pushThisIdentity)
			confirmations <- &WatchIdentity{pushThisIdentity, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received KeyRegistered event from [%x] for keyType: %x and value: %x\n", pushThisIdentity.CentrifugeId, res.KType, res.Key)
			confirmations <- &WatchIdentity{pushThisIdentity, nil}
			return
		}
	}
}

// waitAndRouteIdentityRegistrationEvent notifies the confirmations channel whenever the identity creation is being noted as Ethereum event
func waitAndRouteIdentityRegistrationEvent(conf <-chan *EthereumIdentityFactoryContractIdentityCreated, ctx context.Context, confirmations chan<- *WatchIdentity, pushThisIdentity *EthereumIdentity) {
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving IdentityCreated event for Identity ID: %x\n", ctx, pushThisIdentity)
			confirmations <- &WatchIdentity{pushThisIdentity, ctx.Err()}
			return
		case res := <-conf:
			log.Infof("Received IdentityCreated event from: %x, identifier: %s\n", res.CentrifugeId, res.Identity.String())
			confirmations <- &WatchIdentity{pushThisIdentity, nil}
			return
		}
	}
}

func CreateEthereumIdentityFromApi(centrifugeId []byte, idKey [32]byte) (err error) {
	return createOrAddKeyOnEthereumIdentity(centrifugeId, KEY_TYPE_PEERID, idKey, ACTION_CREATE)
}

func AddKeyToIdentityFromApi(centrifugeId []byte, keyType int, idKey [32]byte) (err error) {
	return createOrAddKeyOnEthereumIdentity(centrifugeId, keyType, idKey, ACTION_ADDKEY)
}

func createOrAddKeyOnEthereumIdentity(centrifugeId []byte, keyType int, idKey [32]byte, action string) (err error) {
	if tools.IsEmptyByteSlice(centrifugeId) || len(centrifugeId) != 32 {
		return errors.New("centrifugeId empty or of incorrect length")
	}

	id := NewEthereumIdentity()
	id.CentrifugeId = centrifugeId
	exists, errLocal := id.CheckIdentityExists()
	if errLocal != nil {
		err = errLocal
		return
	}
	if (action == ACTION_CREATE && exists) || (action == ACTION_ADDKEY && !exists) {
		err = errors.New(fmt.Sprintf("ACTION [%v] but identity exists [%v]", action, exists))
		return
	}

	pid, errLocal := keytools.PublicKeyToP2PKey(idKey)
	if errLocal != nil {
		err = errLocal
		return
	}
	if action == ACTION_ADDKEY {
		currentKey, errLocal := id.GetLastB58KeyForType(keyType)
		if errLocal != nil {
			err = errLocal
			return
		}
		if currentKey == pid.Pretty() {
			err = errors.New("Key trying to be added already exists as latest. Skipping Update.")
			return
		}
	}

	id.Keys[keyType] = append(id.Keys[keyType], EthereumIdentityKey{idKey})
	confirmations := make(chan *WatchIdentity, 1)

	if action == ACTION_CREATE {
		log.Infof("Creating Identity [%v] with PeerID [%v]\n", centrifugeId, pid.Pretty())
		err = CreateEthereumIdentity(id, confirmations)
		if err != nil {
			return
		}
		watchRegisteredIdentity := <-confirmations
		log.Infof("Identity [%v] Created", watchRegisteredIdentity.Identity.CentrifugeIdString())
	}

	log.Infof("Adding Key [%v] to Identity [%v]\n", pid.Pretty(), centrifugeId)
	err = id.AddKeyToIdentity(keyType, confirmations)
	if err != nil {
		return
	}
	watchAddedToIdentity := <-confirmations

	lastKey, errLocal := watchAddedToIdentity.Identity.GetLastB58KeyForType(keyType)
	if errLocal != nil {
		err = errLocal
		return
	}
	log.Infof("%v Key [%v] Added to Identity [%v]", action, lastKey, watchAddedToIdentity.Identity.CentrifugeIdString())
	return
}

func NewEthereumIdentityService() IdentityService {
	return &EthereumIdentityService{}
}

// EthereumidentityService implements `IdentityService`
type EthereumIdentityService struct {
}

func (ids *EthereumIdentityService) LookupIdentityForId(centrifugeId []byte) (id Identity, err error) {
	id = NewEthereumIdentity()
	err = id.SetCentrifugeId(centrifugeId)
	if err != nil {
		return id, err
	}

	exists, err := id.CheckIdentityExists()
	if !exists {
		return id, fmt.Errorf("Identity [%s] does not exist", id.CentrifugeIdString())
	}

	if err != nil {
		return id, err
	}

	return id, nil
}
