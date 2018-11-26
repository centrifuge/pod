package identity

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

const (
	keyRegistrationConfirmationTaskName = "keyRegistrationConfirmationTaskName"
	keyParam                            = "keyParam"
	keyPurposeParam                     = "keyPurposeParam"
)

type keyRegisteredFilterer interface {
	FilterKeyAdded(opts *bind.FilterOpts, key [][32]byte, purpose []*big.Int) (*EthereumIdentityContractKeyAddedIterator, error)
}

// keyRegistrationConfirmationTask is a queued task to filter key registration events on Ethereum using EthereumIdentityContract.
// To see how it gets registered see bootstrapper.go and to see how it gets used see setUpKeyRegisteredEventListener method
type keyRegistrationConfirmationTask struct {
	centID             CentID
	key                [32]byte
	keyPurpose         int
	blockHeight        uint64
	timeout            time.Duration
	contextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc)
	ctx                context.Context
	filterer           keyRegisteredFilterer
	contract           *EthereumIdentityRegistryContract
	config             Config
	gethClientFinder   func() ethereum.Client
	contractProvider   func(address common.Address, backend bind.ContractBackend) (contract, error)
	queue              *queue.Server
}

func newKeyRegistrationConfirmationTask(
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc),
	registryContract *EthereumIdentityRegistryContract,
	config Config,
	queue *queue.Server,
	gethClientFinder func() ethereum.Client,
	contractProvider func(address common.Address, backend bind.ContractBackend) (contract, error),
) *keyRegistrationConfirmationTask {
	return &keyRegistrationConfirmationTask{
		contextInitializer: ethContextInitializer,
		contract:           registryContract,
		config:             config,
		queue:              queue,
		timeout:            config.GetEthereumContextWaitTimeout(),
		gethClientFinder:   gethClientFinder,
		contractProvider:   contractProvider,
	}
}

// TaskTypeName returns keyRegistrationConfirmationTaskName
func (krct *keyRegistrationConfirmationTask) TaskTypeName() string {
	return keyRegistrationConfirmationTaskName
}

// Copy returns a new copy of the task
func (krct *keyRegistrationConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &keyRegistrationConfirmationTask{
		krct.centID,
		krct.key,
		krct.keyPurpose,
		krct.blockHeight,
		krct.timeout,
		krct.contextInitializer,
		krct.ctx,
		krct.filterer,
		krct.contract,
		krct.config,
		krct.gethClientFinder,
		krct.contractProvider,
		krct.queue,
	}, nil
}

// ParseKwargs parses the args into the task
func (krct *keyRegistrationConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
	id, ok := kwargs[centIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + centIDParam)
	}
	centID, err := getCentID(id)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", centIDParam, err.Error())
	}
	krct.centID = centID

	// key parsing
	key, ok := kwargs[keyParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + keyParam)
	}
	keyTyped, err := getBytes32(key)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", keyParam, err.Error())
	}
	krct.key = keyTyped

	// key purpose parsing
	keyPurpose, ok := kwargs[keyPurposeParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + keyPurposeParam)
	}
	keyPurposeF, ok := keyPurpose.(float64)
	if ok {
		krct.keyPurpose = int(keyPurposeF)
	} else {
		return fmt.Errorf("can not parse " + keyPurposeParam)
	}

	// block height parsing
	krct.blockHeight, err = queue.ParseBlockHeight(kwargs)
	if err != nil {
		return err
	}

	tdRaw, ok := kwargs[queue.TimeoutParam]
	if ok {
		td, err := queue.GetDuration(tdRaw)
		if err != nil {
			return fmt.Errorf("malformed kwarg [%s] because [%s]", queue.TimeoutParam, err.Error())
		}
		krct.timeout = td
	}

	return nil
}

// RunTask calls listens to events from geth related to keyRegistrationConfirmationTask#Key and records result.
func (krct *keyRegistrationConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the Key [%x]", krct.key)
	if krct.ctx == nil {
		krct.ctx, _ = krct.contextInitializer(krct.timeout)
	}

	id := newEthereumIdentity(krct.centID, krct.contract, krct.config, krct.queue, krct.gethClientFinder, krct.contractProvider)
	contract, err := id.getContract()
	if err != nil {
		return nil, err
	}

	krct.filterer = contract
	fOpts := &bind.FilterOpts{
		Context: krct.ctx,
		Start:   krct.blockHeight,
	}

	for {
		iter, err := krct.filterer.FilterKeyAdded(fOpts, [][32]byte{krct.key}, []*big.Int{big.NewInt(int64(krct.keyPurpose))})
		if err != nil {
			return nil, centerrors.Wrap(err, "failed to start filtering key event logs")
		}

		err = utils.LookForEvent(iter)
		if err == nil {
			log.Infof("Received filtered event Key Registration Confirmation for CentrifugeID [%s] and key [%x] with purpose [%d]\n", krct.centID.String(), krct.key, krct.keyPurpose)
			return iter.Event, nil
		}

		if err != utils.ErrEventNotFound {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}
}
