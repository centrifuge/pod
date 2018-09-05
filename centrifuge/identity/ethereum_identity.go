package identity

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
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
	peerdId, _ := ed25519.PublicKeyToP2PKey(idk.Key)
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
	if len(b) != CentIdByteLength {
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

func (id *EthereumIdentity) CentrifugeIdBytes() [CentIdByteLength]byte {
	var idBytes [CentIdByteLength]byte
	copy(idBytes[:], id.CentrifugeId[:CentIdByteLength])
	return idBytes
}

// Solidity works with bigendian format, this function returns a bigendian from a CentIdByteLength byte cent id
func (id *EthereumIdentity) CentrifugeIdBigInt() *big.Int {
	bi := tools.ByteSliceToBigInt(id.CentrifugeId)
	return bi
}

func (id *EthereumIdentity) String() string {
	return fmt.Sprintf("CentrifugeId [%s]", id.CentrifugeIdString())
}

func (id *EthereumIdentity) GetCentrifugeId() []byte {
	return id.CentrifugeId
}

func (id *EthereumIdentity) GetLastKeyForPurpose(keyPurpose int) (key []byte, err error) {
	err = id.fetchKeysByPurpose(keyPurpose)
	if err != nil {
		return
	}

	if len(id.cachedKeys[keyPurpose]) == 0 {
		return []byte{}, fmt.Errorf("No key found for type [%d] in id [%s]", keyPurpose, id.CentrifugeIdString())
	}

	return id.cachedKeys[keyPurpose][len(id.cachedKeys[keyPurpose])-1].Key[:32], nil
}

func (id *EthereumIdentity) FetchKey(key []byte) (IdentityKey, error) {
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
	idAddress, err := ethIdentityRegistryContract.GetIdentityByCentrifugeId(opts, id.CentrifugeIdBigInt())
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
	if tools.IsEmptyByteSlice(key) || len(key) != 32 {
		return confirmations, errors.New("Can't add key to identity: Inavlid key")
	}

	ethIdentityContract, err := id.getContract()
	if err != nil {
		return
	}

	confirmations, err = setUpKeyRegisteredEventListener(ethIdentityContract, id, keyPurpose, key)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to set up event listener for identity [id: %s]: %v", id, wError)
		return
	}

	opts, err := ethereum.GetGethTxOpts(config.Config.GetEthereumDefaultAccountName())
	if err != nil {
		return confirmations, err
	}

	err = sendKeyRegistrationTransaction(ethIdentityContract, opts, id, keyPurpose, key)
	if err != nil {
		wError := errors.Wrap(err, 1)
		log.Infof("Failed to create transaction for identity [id: %s]: %v", id, wError)
		return confirmations, wError
	}
	return confirmations, nil
}

func (id *EthereumIdentity) fetchKeysByPurpose(keyPurpose int) error {
	contract, err := id.getContract()
	if err != nil {
		return err
	}
	opts := ethereum.GetGethCallOpts()
	bigInt := big.NewInt(int64(keyPurpose))
	keys, err := contract.GetKeysByPurpose(opts, bigInt)
	if err != nil {
		return err
	}

	for _, key := range keys {
		id.cachedKeys[keyPurpose] = append(id.cachedKeys[keyPurpose], EthereumIdentityKey{key, nil, nil})
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
	tx, err := ethereum.SubmitTransactionWithRetries(identityFactory.CreateIdentity, opts, identityToBeCreated.CentrifugeIdBigInt())

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
	go waitAndRouteKeyRegistrationEvent(keyAddedEvents, watchOpts.Context, confirmations, identity)

	b32Key, err := tools.SliceToByte32(key)
	if err != nil {
		return confirmations, err
	}
	bigInt := big.NewInt(int64(keyPurpose))

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = ethCreatedContract.WatchKeyAdded(watchOpts, keyAddedEvents, [][32]byte{b32Key}, []*big.Int{bigInt})
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
func setUpRegistrationEventListener(identityToBeCreated Identity) (confirmations chan *WatchIdentity, err error) {
	confirmations = make(chan *WatchIdentity)
	bCentId := identityToBeCreated.CentrifugeIdBytes()
	asyncRes, err := queue.Queue.DelayKwargs(IdRegistrationConfirmationTaskName, map[string]interface{}{CentIdParam: bCentId})
	if err != nil {
		return nil, err
	}
	go waitAndRouteIdentityRegistrationEvent(asyncRes, confirmations, identityToBeCreated)
	return confirmations, nil
}

// waitAndRouteKeyRegistrationEvent notifies the confirmations channel whenever the key has been added to the identity and has been noted as Ethereum event
func waitAndRouteKeyRegistrationEvent(conf <-chan *EthereumIdentityContractKeyAdded, ctx context.Context, confirmations chan<- *WatchIdentity, pushThisIdentity Identity) {
	for {
		select {
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

func NewEthereumIdentityService() IdentityService {
	return &EthereumIdentityService{}
}

// EthereumidentityService implements `IdentityService`
type EthereumIdentityService struct {
}

func (ids *EthereumIdentityService) CheckIdentityExists(centrifugeId []byte) (exists bool, err error) {
	if tools.IsEmptyByteSlice(centrifugeId) || len(centrifugeId) != CentIdByteLength {
		return false, errors.New("centrifugeId empty or of incorrect length")
	}
	id := NewEthereumIdentity()
	id.CentrifugeId = centrifugeId
	exists, err = id.CheckIdentityExists()
	return
}

func (ids *EthereumIdentityService) CreateIdentity(centrifugeId []byte) (id Identity, confirmations chan *WatchIdentity, err error) {
	log.Infof("Creating Identity [%x]", centrifugeId)

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

	confirmations, err = setUpRegistrationEventListener(id)
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
		return instanceId, fmt.Errorf("identity [%s] does not exist", instanceId.CentrifugeIdString())
	}

	if err != nil {
		return instanceId, err
	}

	return instanceId, nil
}
