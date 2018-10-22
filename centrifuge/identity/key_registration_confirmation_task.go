package identity

import (
	"context"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/centrifuge/utils"

	"time"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/queue"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

const (
	KeyRegistrationConfirmationTaskName string = "KeyRegistrationConfirmationTaskName"
	KeyParam                            string = "KeyParam"
	KeyPurposeParam                            = "KeyPurposeParam"
)

type KeyRegisteredFilterer interface {
	FilterKeyAdded(opts *bind.FilterOpts, key [][32]byte, purpose []*big.Int) (*EthereumIdentityContractKeyAddedIterator, error)
}

// KeyRegistrationConfirmationTask is a queued task to filter key registration events on Ethereum using EthereumIdentityContract.
// To see how it gets registered see bootstrapper.go and to see how it gets used see setUpKeyRegisteredEventListener method
type KeyRegistrationConfirmationTask struct {
	CentID                CentID
	Key                   [32]byte
	KeyPurpose            int
	BlockHeight           uint64
	EthContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc)
	EthContext            context.Context
	KeyRegisteredWatcher  KeyRegisteredFilterer
	RegistryContract      *EthereumIdentityRegistryContract
	Config                Config
}

func NewKeyRegistrationConfirmationTask(
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc),
	registryContract *EthereumIdentityRegistryContract,
	config Config,
) *KeyRegistrationConfirmationTask {
	return &KeyRegistrationConfirmationTask{
		EthContextInitializer: ethContextInitializer,
		RegistryContract:      registryContract,
		Config:                config,
	}
}

func (krct *KeyRegistrationConfirmationTask) Name() string {
	return KeyRegistrationConfirmationTaskName
}

func (krct *KeyRegistrationConfirmationTask) Init() error {
	queue.Queue.Register(KeyRegistrationConfirmationTaskName, krct)
	return nil
}

func (krct *KeyRegistrationConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &KeyRegistrationConfirmationTask{
		krct.CentID,
		krct.Key,
		krct.KeyPurpose,
		krct.BlockHeight,
		krct.EthContextInitializer,
		krct.EthContext,
		krct.KeyRegisteredWatcher,
		krct.RegistryContract,
		krct.Config}, nil
}

// ParseKwargs - define a method to parse params
func (krct *KeyRegistrationConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
	centId, ok := kwargs[CentIdParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + CentIdParam)
	}
	centIdTyped, err := getCentID(centId)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", CentIdParam, err.Error())
	}
	krct.CentID = centIdTyped

	// key parsing
	key, ok := kwargs[KeyParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + KeyParam)
	}
	keyTyped, err := getBytes32(key)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", KeyParam, err.Error())
	}
	krct.Key = keyTyped

	// key purpose parsing
	keyPurpose, ok := kwargs[KeyPurposeParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + KeyPurposeParam)
	}
	keyPurposeF, ok := keyPurpose.(float64)
	if ok {
		krct.KeyPurpose = int(keyPurposeF)
	} else {
		return fmt.Errorf("can not parse " + KeyPurposeParam)
	}

	// block height parsing
	krct.BlockHeight, err = parseBlockHeight(kwargs)
	if err != nil {
		return err
	}
	return nil
}

// RunTask calls listens to events from geth related to KeyRegistrationConfirmationTask#Key and records result.
func (krct *KeyRegistrationConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the Key [%x]", krct.Key)
	if krct.EthContext == nil {
		krct.EthContext, _ = krct.EthContextInitializer()
	}

	id := EthereumIdentity{CentrifugeId: krct.CentID, RegistryContract: krct.RegistryContract, Config: krct.Config}
	contract, err := id.getContract()
	if err != nil {
		return nil, err
	}
	krct.KeyRegisteredWatcher = contract

	fOpts := &bind.FilterOpts{
		Context: krct.EthContext,
		Start:   krct.BlockHeight,
	}

	for {
		iter, err := krct.KeyRegisteredWatcher.FilterKeyAdded(
			fOpts,
			[][32]byte{krct.Key},
			[]*big.Int{big.NewInt(int64(krct.KeyPurpose))},
		)
		if err != nil {
			return nil, centerrors.Wrap(err, "failed to start filtering key event logs")
		}

		err = utils.LookForEvent(iter)
		if err == nil {
			log.Infof("Received filtered event Key Registration Confirmation for CentrifugeId [%s] and key [%x] with purpose [%d]\n", krct.CentID.String(), krct.Key, krct.KeyPurpose)
			return iter.Event, nil
		}

		if err != utils.EventNotFound {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to filter key events")
}
