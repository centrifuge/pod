package identity

import (
	"context"
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
	CentrifugeId string
	Keys         map[int][]EthereumIdentityKey
}

func NewEthereumIdentity() (id *EthereumIdentity) {
	id = new(EthereumIdentity)
	id.Keys = make(map[int][]EthereumIdentityKey)
	return
}

func (id *EthereumIdentity) String() string {
	joinedKeys := ""
	for k, v := range id.Keys {
		for i, _ := range v {
			joinedKeys += fmt.Sprintf("[%v]%s ", k, v[i].String())
		}
	}
	return fmt.Sprintf("CentrifugeId [%s], Keys [%s]", id.CentrifugeId, joinedKeys)
}

func (id *EthereumIdentity) GetCentrifugeId() string {
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
		log.Infof("keyType Not found")
	}
	return
}

func (id *EthereumIdentity) CheckIdentityExists() (exists bool, err error) {
	idContract, err := findIdentity(id.GetCentrifugeId())
	if err != nil {
		return false, err
	}
	if idContract != nil {
		opts := ethereum.GetGethCallOpts()
		centId, err := idContract.CentrifugeId(opts)
		if err == bind.ErrNoCode { //no contract in specified address, meaning Identity was not created
			log.Infof("Identity contract does not exist!")
			err = nil
		} else if len(centId) != 0 {
			log.Infof("Identity exists!")
			exists = true
		}
	}
	return
}

func (id *EthereumIdentity) AddKeyToIdentity(keyType int, confirmations chan<- *EthereumIdentity) (err error) {
	ethIdentityContract, err := findIdentity(id.CentrifugeId)
	if err != nil {
		return
	}

	err = setUpKeyRegisteredEventListener(ethIdentityContract, id, keyType, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to set up event listener for identity [id: %x]: %v", id.CentrifugeId, wError)
		return
	}

	opts, err := ethereum.GetGethTxOpts()
	if err != nil {
		return err
	}

	err = sendKeyRegistrationTransaction(ethIdentityContract, opts, id, keyType)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to create transaction for identity [id: %x]: %v", id.CentrifugeId, wError)
		return
	}
	return
}

func getIdentityFactoryContract() (identityFactoryContract *EthereumIdentityFactoryContract, err error) {
	client := ethereum.GetConnection()

	identityFactoryContract, err = NewEthereumIdentityFactoryContract(common.HexToAddress(config.Config.GetContractAddress("identityFactory")), client.GetClient())
	if err != nil {
		log.Infof("Failed to instantiate the identity factory contract: %v", err)
	}
	return
}

func getIdentityRegistryContract() (identityRegistryContract *EthereumIdentityRegistryContract, err error) {
	client := ethereum.GetConnection()

	identityRegistryContract, err = NewEthereumIdentityRegistryContract(common.HexToAddress(config.Config.GetContractAddress("identityRegistry")), client.GetClient())
	if err != nil {
		log.Infof("Failed to instantiate the identity registry contract: %v", err)
	}
	return
}

func getIdentityContract(identityContractAddress string) (identityContract *EthereumIdentityContract, err error) {
	client := ethereum.GetConnection()

	identityContract, err = NewEthereumIdentityContract(common.HexToAddress(identityContractAddress), client.GetClient())
	if err != nil {
		log.Infof("Failed to instantiate the identity contract: %v", err)
	}
	return
}

func CreateEthereumIdentity(identity *EthereumIdentity, confirmations chan<- *EthereumIdentity) (err error) {
	err = tools.CheckLen32(identity.CentrifugeId, "centrifugeId needs to be length of 32. Got value [%v]")
	if err != nil {
		return
	}
	ethIdentityFactoryContract, err := getIdentityFactoryContract()
	if err != nil {
		return
	}
	opts, err := ethereum.GetGethTxOpts()
	if err != nil {
		return err
	}

	err = setUpRegistrationEventListener(ethIdentityFactoryContract, identity, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to set up event listener for identity [id: %x]: %v", identity.CentrifugeId, wError)
		return
	}

	err = sendIdentityCreationTransaction(ethIdentityFactoryContract, opts, identity)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to create transaction for identity [id: %x]: %v", identity.CentrifugeId, wError)
		return
	}
	return
}

func findIdentity(centrifugeId string) (identityContract *EthereumIdentityContract, err error) {
	err = tools.CheckLen32(centrifugeId, "centrifugeId needs to be length of 32. Got value [%v]")
	if err != nil {
		return
	}
	ethIdentityRegistryContract, err := getIdentityRegistryContract()
	if err != nil {
		return
	}
	opts := ethereum.GetGethCallOpts()
	var b32CentId [32]byte
	copy(b32CentId[:], centrifugeId[:32])
	idAddress, err := ethIdentityRegistryContract.GetIdentityByCentrifugeId(opts, b32CentId)
	if err != nil {
		return
	}
	identityContract, err = getIdentityContract(idAddress.String())
	if err != nil {
		return
	}
	return
}

func ResolveP2PEthereumIdentityForId(centrifugeId string) (id *EthereumIdentity, err error) {
	id, err = resolveEthereumIdentityForKeyType(centrifugeId, 1)
	return
}

func resolveEthereumIdentityForKeyType(centrifugeId string, keyType int) (id *EthereumIdentity, err error) {
	ethIdentityContract, err := findIdentity(centrifugeId)
	if err != nil {
		return
	}
	opts := ethereum.GetGethCallOpts()
	bigInt := big.NewInt(int64(keyType))
	keys, err := ethIdentityContract.GetKeysByType(opts, bigInt)
	if err != nil {
		return
	}
	id = NewEthereumIdentity()
	id.CentrifugeId = centrifugeId
	for _, key := range keys {
		id.Keys[keyType] = append(id.Keys[keyType], EthereumIdentityKey{key})
	}
	return
}

// sendRegistrationTransaction sends the actual transaction to add a Key on Ethereum registry contract
func sendKeyRegistrationTransaction(identityContract IdentityContract, opts *bind.TransactOpts, identity *EthereumIdentity, keyType int) (err error) {

	//preparation of data in specific types for the call to Ethereum
	var bKey [32]byte
	lastKey := len(identity.Keys[keyType]) - 1
	copy(bKey[:], identity.Keys[keyType][lastKey].Key[:32])
	bigInt := big.NewInt(int64(keyType))

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
	err = tools.CheckLen32(identityToBeCreated.CentrifugeId, "CentrifugeId needs to be length of 32. Got value [%x]")
	if err != nil {
		return err
	}

	//preparation of data in specific types for the call to Ethereum
	var bCentId [32]byte
	copy(bCentId[:], identityToBeCreated.CentrifugeId[:32])

	tx, err := ethereum.SubmitTransactionWithRetries(identityFactory.CreateIdentity, opts, bCentId)

	if err != nil {
		log.Infof("Failed to send identity for creation [CentrifugeID: %x] : %v", bCentId, err)
		return err
	} else {
		log.Infof("Sent off identity creation [CentrifugeID: %x]. Ethereum transaction hash [%x]", bCentId, tx.Hash())
	}

	log.Infof("Transfer pending: 0x%x\n", tx.Hash())

	return
}

func setUpKeyRegisteredEventListener(ethCreatedContract WatchKeyRegistered, identity *EthereumIdentity, keyType int, confirmations chan<- *EthereumIdentity) (err error) {
	//listen to this particular key being mined/event is triggered
	watchOpts := &bind.WatchOpts{}
	watchOpts.Context = ethereum.DefaultWaitForTransactionMiningContext()

	//only setting up a channel of 1 notification as there should always be only one notification coming for this
	//single key being registered
	keyAddedEvents := make(chan *EthereumIdentityContractKeyRegistered, 1)
	go waitAndRouteKeyRegistrationEvent(keyAddedEvents, watchOpts.Context, confirmations, identity)

	var bKey [32]byte
	lastKey := len(identity.Keys[keyType]) - 1
	copy(bKey[:], identity.Keys[keyType][lastKey].Key[:32])
	bigInt := big.NewInt(int64(keyType))

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.

	_, err = ethCreatedContract.WatchKeyRegistered(watchOpts, keyAddedEvents, []*big.Int{bigInt}, [][32]byte{bKey})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Infof(wError.Error())
		panic(wError)
	}
	return
}

// setUpRegistrationEventListener sets up the listened for the "IdentityCreated" event to notify the upstream code about successful mining/creation
// of the identity.
func setUpRegistrationEventListener(ethCreatedContract WatchIdentityCreated, identityToBeCreated *EthereumIdentity, confirmations chan<- *EthereumIdentity) (err error) {

	//listen to this particular identity being mined/event is triggered
	watchOpts := &bind.WatchOpts{}
	watchOpts.Context = ethereum.DefaultWaitForTransactionMiningContext()

	//only setting up a channel of 1 notification as there should always be only one notification coming for this
	//single identity being registered
	identityCreatedEvents := make(chan *EthereumIdentityFactoryContractIdentityCreated, 1)
	go waitAndRouteIdentityRegistrationEvent(identityCreatedEvents, watchOpts.Context, confirmations, identityToBeCreated)

	var bCentId [32]byte
	copy(bCentId[:], identityToBeCreated.CentrifugeId[:32])

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = ethCreatedContract.WatchIdentityCreated(watchOpts, identityCreatedEvents, [][32]byte{bCentId})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Infof(wError.Error())
		panic(wError)
	}
	return
}

// waitAndRouteKeyRegistrationEvent notifies the confirmations channel whenever the key has been added to the identity and has been noted as Ethereum event
func waitAndRouteKeyRegistrationEvent(conf <-chan *EthereumIdentityContractKeyRegistered, ctx context.Context, confirmations chan<- *EthereumIdentity, pushThisIdentity *EthereumIdentity) {
	for {
		select {
		case <-ctx.Done():
			log.Fatalf("Context [%v] closed before receiving KeyRegistered event for Identity ID: %x\n", ctx, pushThisIdentity)
			return
		case res := <-conf:
			log.Infof("Received KeyRegistered event from [%x] for keyType: %x and value: %x\n", pushThisIdentity.CentrifugeId, res.KType, res.Key)
			confirmations <- pushThisIdentity
			return
		}
	}
}

// waitAndRouteIdentityRegistrationEvent notifies the confirmations channel whenever the identity creation is being noted as Ethereum event
func waitAndRouteIdentityRegistrationEvent(conf <-chan *EthereumIdentityFactoryContractIdentityCreated, ctx context.Context, confirmations chan<- *EthereumIdentity, pushThisIdentity *EthereumIdentity) {
	for {
		select {
		case <-ctx.Done():
			log.Fatalf("Context [%v] closed before receiving IdentityCreated event for Identity ID: %x\n", ctx, pushThisIdentity)
			return
		case res := <-conf:
			log.Infof("Received IdentityCreated event from: %x, identifier: %s\n", res.CentrifugeId, res.Identity.String())
			confirmations <- pushThisIdentity
			return
		}
	}
}

func CreateEthereumIdentityFromApi(centrifugeId string, idKey [32]byte) (err error) {
	return createOrAddKeyOnEthereumIdentity(centrifugeId, KEY_TYPE_PEERID, idKey, ACTION_CREATE)
}

func AddKeyToIdentityFromApi(centrifugeId string, keyType int, idKey [32]byte) (err error) {
	return createOrAddKeyOnEthereumIdentity(centrifugeId, keyType, idKey, ACTION_ADDKEY)
}

func createOrAddKeyOnEthereumIdentity(centrifugeId string, keyType int, idKey [32]byte, action string) (err error) {
	if centrifugeId == "" {
		err = errors.New("Centrifuge ID not provided")
		return
	}
	_, err = tools.StringToByte32(centrifugeId)
	if err != nil {
		return
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
		currentId, errLocal := ResolveP2PEthereumIdentityForId(centrifugeId)
		if errLocal != nil {
			err = errLocal
			return
		}
		currentKey, errLocal := currentId.GetLastB58KeyForType(keyType)
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
	confirmations := make(chan *EthereumIdentity, 1)

	if action == ACTION_CREATE {
		log.Infof("Creating Identity [%v] with PeerID [%v]\n", centrifugeId, pid.Pretty())
		err = CreateEthereumIdentity(id, confirmations)
		if err != nil {
			return
		}
		registeredIdentity := <-confirmations
		log.Infof("Identity [%v] Created", registeredIdentity.CentrifugeId)
	}

	log.Infof("Adding Key [%v] to Identity [%v]\n", pid.Pretty(), centrifugeId)
	err = id.AddKeyToIdentity(keyType, confirmations)
	if err != nil {
		return
	}
	addedToIdentity := <-confirmations

	lastKey, errLocal := addedToIdentity.GetLastB58KeyForType(keyType)
	if errLocal != nil {
		err = errLocal
		return
	}
	log.Infof("%v Key [%v] Added to Identity [%v]", action, lastKey, addedToIdentity.CentrifugeId)
	return
}
