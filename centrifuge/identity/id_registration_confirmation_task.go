package identity

import (
	"fmt"

	"context"

	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
)

const IdRegistrationConfirmationTaskName string = "IdRegistrationConfirmationTaskName"
const CentIdParam string = "CentID"

type IdentityCreatedWatcher interface {
	WatchIdentityCreated(opts *bind.WatchOpts, sink chan<- *EthereumIdentityFactoryContractIdentityCreated, centrifugeId []*big.Int) (event.Subscription, error)
}

// IdRegistrationConfirmationTask is a queued task to watch ID registration events on Ethereum using EthereumIdentityFactoryContract.
// To see how it gets registered see bootstrapper.go and to see how it gets used see setUpRegistrationEventListener method
type IdRegistrationConfirmationTask struct {
	CentId                 CentID
	EthContextInitializer  func() (ctx context.Context, cancelFunc context.CancelFunc)
	IdentityCreatedEvents  chan *EthereumIdentityFactoryContractIdentityCreated
	EthContext             context.Context
	IdentityCreatedWatcher IdentityCreatedWatcher
}

func NewIdRegistrationConfirmationTask(
	identityCreatedWatcher IdentityCreatedWatcher,
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc),
) queue.QueuedTask {
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

// ParseKwargs - define a method to parse CentID
func (rct *IdRegistrationConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
	centId, ok := kwargs[CentIdParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + CentIdParam)
	}
	centIdTyped, err := getBytes(centId)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", CentIdParam, err.Error())
	}
	rct.CentId = centIdTyped
	return nil
}

// RunTask calls listens to events from geth related to IdRegistrationConfirmationTask#CentID and records result.
func (rct *IdRegistrationConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the ID [%x]", rct.CentId)
	if rct.EthContext == nil {
		rct.EthContext, _ = rct.EthContextInitializer()
	}
	watchOpts := &bind.WatchOpts{Context: rct.EthContext}
	if rct.IdentityCreatedEvents == nil {
		rct.IdentityCreatedEvents = make(chan *EthereumIdentityFactoryContractIdentityCreated)
	}

	subscription, err := rct.IdentityCreatedWatcher.WatchIdentityCreated(watchOpts, rct.IdentityCreatedEvents, []*big.Int{tools.ByteFixedToBigInt(rct.CentId[:], CentIDByteLength)})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Errorf(wError.Error())
		return nil, wError
	}
	for {
		select {
		case err := <-subscription.Err():
			log.Errorf("Subscription error %s", err.Error())
			return nil, err
		case <-rct.EthContext.Done():
			log.Errorf("Context [%v] closed before receiving IdRegistered event for Identity ID: %x\n", rct.EthContext, rct.CentId)
			return nil, rct.EthContext.Err()
		case res := <-rct.IdentityCreatedEvents:
			log.Infof("Received IdentityCreated event from: %x, identifier: %x\n", res.CentrifugeId, res.Identity)
			subscription.Unsubscribe()
			return res, nil
		}
	}
}

func getBytes(key interface{}) (CentID, error) {
	var fixed [CentIDByteLength]byte
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
