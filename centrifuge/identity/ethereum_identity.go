package identity

import (
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math/big"
	"github.com/go-errors/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/event"
	"context"
	"github.com/ethereum/go-ethereum/core/types"
)

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

func getIdentityFactoryContract() (identityFactoryContract *EthereumIdentityFactoryContract, err error) {
	client := ethereum.GetConnection()

	identityFactoryContract, err = NewEthereumIdentityFactoryContract(common.HexToAddress(viper.GetString("identity.ethereum.identityFactoryAddress")), client.GetClient())
	if err != nil {
		log.Printf("Failed to instantiate the identity factory contract: %v", err)
	}
	return
}

func getIdentityRegistryContract() (identityRegistryContract *EthereumIdentityRegistryContract, err error) {
	client := ethereum.GetConnection()

	identityRegistryContract, err = NewEthereumIdentityRegistryContract(common.HexToAddress(viper.GetString("identity.ethereum.identityRegistryAddress")), client.GetClient())
	if err != nil {
		log.Printf("Failed to instantiate the identity registry contract: %v", err)
	}
	return
}

func getIdentityContract(identityContractAddress string) (identityContract *EthereumIdentityContract, err error) {
	client := ethereum.GetConnection()

	identityContract, err = NewEthereumIdentityContract(common.HexToAddress(identityContractAddress), client.GetClient())
	if err != nil {
		log.Printf("Failed to instantiate the identity contract: %v", err)
	}
	return
}

func doAddKeyToIdentity(identity Identity, keyType int, confirmations chan<- *Identity) (err error) {
	ethIdentityContract, err := doFindIdentity(identity.CentrifugeId)
	if err != nil {
		return
	}

	err = setUpKeyRegisteredEventListener(ethIdentityContract, &identity, keyType, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Printf("Failed to set up event listener for identity [id: %x]: %v", identity.CentrifugeId, wError)
		return
	}

	opts, err := ethereum.GetGethTxOpts()
	if err != nil {
		return err
	}

	err = sendKeyRegistrationTransaction(ethIdentityContract, opts, &identity, keyType)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Printf("Failed to create transaction for identity [id: %x]: %v", identity.CentrifugeId, wError)
		return
	}
	return
}

func doCreateIdentity(identity Identity, confirmations chan<- *Identity) (err error) {
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

	err = setUpRegistrationEventListener(ethIdentityFactoryContract, &identity, confirmations)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Printf("Failed to set up event listener for identity [id: %x]: %v", identity.CentrifugeId, wError)
		return
	}

	err = sendIdentityCreationTransaction(ethIdentityFactoryContract, opts, &identity)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Printf("Failed to create transaction for identity [id: %x]: %v", identity.CentrifugeId, wError)
		return
	}
	return
}

func doFindIdentity(centrifugeId string) (identityContract *EthereumIdentityContract, err error) {
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

func doResolveIdentityForKeyType(centrifugeId string, keyType int) (id Identity, err error) {
	ethIdentityContract, err := doFindIdentity(centrifugeId)
	if err != nil {
		return
	}
	opts := ethereum.GetGethCallOpts()
	bigInt := big.NewInt(int64(keyType))
	keys, err := ethIdentityContract.GetKeysByType(opts, bigInt)
	if err != nil {
		return
	}
	var m = make(map[int][]IdentityKey)
	for _, key := range keys {
		m[keyType] = append(m[keyType], IdentityKey{key})
	}
	id = Identity{ CentrifugeId: centrifugeId , Keys: m }
	return
}

// sendRegistrationTransaction sends the actual transaction to add a Key on Ethereum registry contract
func sendKeyRegistrationTransaction(identityContract IdentityContract, opts *bind.TransactOpts, identity *Identity, keyType int) (err error) {

	//preparation of data in specific types for the call to Ethereum
	var bKey [32]byte
	lastKey := len(identity.Keys[keyType])-1
	copy(bKey[:], identity.Keys[keyType][lastKey].Key[:32])
	bigInt := big.NewInt(int64(keyType))

	tx, err := ethereum.SubmitTransactionWithRetries(identityContract.AddKey, opts, bKey, bigInt)

	if err != nil {
		log.Printf("Failed to send key [%v:%x] to add to CentrifugeID [%x]: %v", keyType, bKey, identity.CentrifugeId, err)
		return err
	} else {
		log.Printf("Sent off key [%v:%x] to add to CentrifugeID [%x]. Ethereum transaction hash [%x]", keyType, bKey, identity.CentrifugeId, tx.Hash())
	}

	log.Printf("Transfer pending: 0x%x\n", tx.Hash())

	return
}

// sendIdentityCreationTransaction sends the actual transaction to create identity on Ethereum registry contract
func sendIdentityCreationTransaction(identityFactory IdentityFactory, opts *bind.TransactOpts, identityToBeCreated *Identity) (err error) {
	err = tools.CheckLen32(identityToBeCreated.CentrifugeId, "CentrifugeId needs to be length of 32. Got value [%x]")
	if err != nil {
		return err
	}

	//preparation of data in specific types for the call to Ethereum
	var bCentId [32]byte
	copy(bCentId[:], identityToBeCreated.CentrifugeId[:32])

	tx, err := ethereum.SubmitTransactionWithRetries(identityFactory.CreateIdentity, opts, bCentId)

	if err != nil {
		log.Printf("Failed to send identity for creation [CentrifugeID: %x] : %v", bCentId, err)
		return err
	} else {
		log.Printf("Sent off identity creation [CentrifugeID: %x]. Ethereum transaction hash [%x]", bCentId, tx.Hash())
	}

	log.Printf("Transfer pending: 0x%x\n", tx.Hash())

	return
}

func setUpKeyRegisteredEventListener(ethCreatedContract WatchKeyRegistered, identity *Identity, keyType int, confirmations chan<- *Identity) (err error) {
	//listen to this particular key being mined/event is triggered
	watchOpts := &bind.WatchOpts{}
	watchOpts.Context = ethereum.DefaultWaitForTransactionMiningContext()

	//only setting up a channel of 1 notification as there should always be only one notification coming for this
	//single key being registered
	keyAddedEvents := make(chan *EthereumIdentityContractKeyRegistered, 1)
	go waitAndRouteKeyRegistrationEvent(keyAddedEvents, watchOpts.Context, confirmations, identity)

	var bKey [32]byte
	lastKey := len(identity.Keys[keyType])-1
	copy(bKey[:], identity.Keys[keyType][lastKey].Key[:32])
	bigInt := big.NewInt(int64(keyType))

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.

	_, err = ethCreatedContract.WatchKeyRegistered(watchOpts, keyAddedEvents, []*big.Int{bigInt}, [][32]byte{bKey})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Printf(wError.Error())
		panic(wError)
	}
	return
}

// setUpRegistrationEventListener sets up the listened for the "IdentityCreated" event to notify the upstream code about successful mining/creation
// of the identity.
func setUpRegistrationEventListener(ethCreatedContract WatchIdentityCreated, identityToBeCreated *Identity, confirmations chan<- *Identity) (err error) {

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
		log.Printf(wError.Error())
		panic(wError)
	}
	return
}

// waitAndRouteKeyRegistrationEvent notifies the confirmations channel whenever the key has been added to the identity and has been noted as Ethereum event
func waitAndRouteKeyRegistrationEvent(conf <-chan *EthereumIdentityContractKeyRegistered, ctx context.Context, confirmations chan<- *Identity, pushThisIdentity *Identity) {
	for {
		select {
		case <-ctx.Done():
			log.Fatalf("Context [%v] closed before receiving KeyRegistered event for Identity ID: %x\n", ctx, pushThisIdentity)
			return
		case res := <-conf:
			log.Printf("Received KeyRegistered event from [%x] for keyType: %x and value: %x\n", pushThisIdentity.CentrifugeId, res.KType, res.Key)
			confirmations <- pushThisIdentity
			return
		}
	}
}

// waitAndRouteIdentityRegistrationEvent notifies the confirmations channel whenever the identity creation is being noted as Ethereum event
func waitAndRouteIdentityRegistrationEvent(conf <-chan *EthereumIdentityFactoryContractIdentityCreated, ctx context.Context, confirmations chan<- *Identity, pushThisIdentity *Identity) {
	for {
		select {
		case <-ctx.Done():
			log.Fatalf("Context [%v] closed before receiving IdentityCreated event for Identity ID: %x\n", ctx, pushThisIdentity)
			return
		case res := <-conf:
			log.Printf("Received IdentityCreated event from: %x, identifier: %s\n", res.CentrifugeId, res.Identity.String())
			confirmations <- pushThisIdentity
			return
		}
	}
}

