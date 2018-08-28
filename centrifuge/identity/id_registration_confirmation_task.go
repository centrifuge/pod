package identity

import (
	"fmt"

	"context"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
)

const IdRegistrationConfirmationTaskName string = "IdRegistrationConfirmationTaskName"
const CentIdParam string = "CentId"

type IdentityCreatedWatcher interface {
	WatchIdentityCreated(opts *bind.WatchOpts, sink chan<- *EthereumIdentityFactoryContractIdentityCreated, centrifugeId [][32]byte) (event.Subscription, error)
}

type IdRegistrationConfirmationTask struct {
	CentId                 [32]byte
	EthContextInitializer  func() (ctx context.Context, cancelFunc context.CancelFunc)
	IdentityCreatedEvents  chan *EthereumIdentityFactoryContractIdentityCreated
	EthContext             context.Context
	IdentityCreatedWatcher IdentityCreatedWatcher
}

func NewIdRegistrationConfirmationTask(
	identityCreatedWatcher IdentityCreatedWatcher,
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc),
) *IdRegistrationConfirmationTask {
	return &IdRegistrationConfirmationTask{
		IdentityCreatedWatcher: identityCreatedWatcher,
		EthContextInitializer:  ethContextInitializer,
	}
}

func (rct *IdRegistrationConfirmationTask) Name() string {
	return IdRegistrationConfirmationTaskName
}

func (rct *IdRegistrationConfirmationTask) Init() error {
	queue.Queue.Register(IdRegistrationConfirmationTaskName, rct)
	return nil
}

func (m *IdRegistrationConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &IdRegistrationConfirmationTask{
		m.CentId,
		m.EthContextInitializer,
		m.IdentityCreatedEvents,
		m.EthContext,
		m.IdentityCreatedWatcher}, nil
}

// ParseKwargs - define a method to parse CentId
func (rct *IdRegistrationConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
	CentId, ok := kwargs[CentIdParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + CentIdParam)
	}
	CentIdTyped, err := getBytes(CentId)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", CentIdParam, err.Error())
	}
	rct.CentId = CentIdTyped
	return nil
}

// RunTask calls listens to events from geth related to IdRegistrationConfirmationTask#CentId and records result.
// Currently covered by TestCreateAndLookupIdentity_Integration test.
func (rct *IdRegistrationConfirmationTask) RunTask() (interface{}, error) {
	rct.EthContext, _ = rct.EthContextInitializer()
	watchOpts := &bind.WatchOpts{Context: rct.EthContext}
	rct.IdentityCreatedEvents = make(chan *EthereumIdentityFactoryContractIdentityCreated)

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err := rct.IdentityCreatedWatcher.WatchIdentityCreated(watchOpts, rct.IdentityCreatedEvents, [][32]byte{rct.CentId})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Errorf(wError.Error())
		return nil, wError
	}
	for {
		select {
		case <-rct.EthContext.Done():
			log.Errorf("Context [%v] closed before receiving KeyRegistered event for Identity ID: %x\n", rct.EthContext, rct.CentId)
			return nil, rct.EthContext.Err()
		case res := <-rct.IdentityCreatedEvents:
			log.Infof("Received IdentityCreated event from: %x, identifier: %x\n", res.CentrifugeId, res.Identity)
			return res, nil
		}
	}
}

func getBytes(key interface{}) ([32]byte, error) {
	var fixed [32]byte
	b, ok := key.([]interface{})
	if !ok {
		return fixed, errors.New("Could not parse interface to []byte")
	}
	// convert and copy b byte values
	for i, v := range b {
		fv := v.(float64)
		fixed[i] = byte(fv)
	}
	return fixed, nil
}
