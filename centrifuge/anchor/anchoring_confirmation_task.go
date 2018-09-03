package anchor

import (
	"fmt"

	"context"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
)

const (
	AnchoringConfirmationTaskName string = "AnchoringConfirmationTaskName"
	AnchorIdParam                 string = "AnchorIdParam"
	AddressParam                  string = "AddressParam"
	ThirtyTwo                            = 32
)

type AnchorRegisteredWatcher interface {
	WatchAnchorRegistered(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRegistryContractAnchorRegistered, from []common.Address, identifier [][32]byte, rootHash [][32]byte) (event.Subscription, error)
}

// AnchoringConfirmationTask is a queued task to watch ID registration events on Ethereum using EthereumIdentityFactoryContract.
// To see how it gets registered see bootstrapper.go and to see how it gets used see setUpRegistrationEventListener method
type AnchoringConfirmationTask struct {
	// task parameters
	From     common.Address
	AnchorId [ThirtyTwo]byte

	// state
	EthContextInitializer   func() (ctx context.Context, cancelFunc context.CancelFunc)
	AnchorRegisteredEvents  chan *EthereumAnchorRegistryContractAnchorRegistered
	EthContext              context.Context
	AnchorRegisteredWatcher AnchorRegisteredWatcher
}

func NewAnchoringConfirmationTask(
	anchorRegisteredWatcher AnchorRegisteredWatcher,
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc),
) *AnchoringConfirmationTask {
	return &AnchoringConfirmationTask{
		AnchorRegisteredWatcher: anchorRegisteredWatcher,
		EthContextInitializer:   ethContextInitializer,
	}
}

func (act *AnchoringConfirmationTask) Name() string {
	return AnchoringConfirmationTaskName
}

func (act *AnchoringConfirmationTask) Init() error {
	queue.Queue.Register(AnchoringConfirmationTaskName, act)
	return nil
}

func (act *AnchoringConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &AnchoringConfirmationTask{
		act.From,
		act.AnchorId,
		act.EthContextInitializer,
		act.AnchorRegisteredEvents,
		act.EthContext,
		act.AnchorRegisteredWatcher,
	}, nil
}

// ParseKwargs - define a method to parse AnchorId, Address and RootHash
func (act *AnchoringConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
	anchorId, ok := kwargs[AnchorIdParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + AnchorIdParam)
	}
	anchorIdTyped, err := get32Bytes(anchorId)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", AnchorIdParam, err.Error())
	}
	act.AnchorId = anchorIdTyped

	// parse the address
	address, ok := kwargs[AddressParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + AddressParam)
	}
	addressTyped, err := getAddress(address)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", AddressParam, err.Error())
	}
	act.From = addressTyped
	return nil
}

// RunTask calls listens to events from geth related to AnchoringConfirmationTask#AnchorId and records result.
func (act *AnchoringConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the anchorID [%x]", act.AnchorId)
	if act.EthContext == nil {
		act.EthContext, _ = act.EthContextInitializer()
	}
	watchOpts := &bind.WatchOpts{Context: act.EthContext}
	if act.AnchorRegisteredEvents == nil {
		act.AnchorRegisteredEvents = make(chan *EthereumAnchorRegistryContractAnchorRegistered)
	}

	subscription, err := act.AnchorRegisteredWatcher.WatchAnchorRegistered(watchOpts, act.AnchorRegisteredEvents, []common.Address{act.From}, [][32]byte{act.AnchorId}, nil)
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Errorf(wError.Error())
		return nil, wError
	}
	for {
		select {
		case err := <-subscription.Err():
			log.Errorf("Subscription error %s for anchor ID: %x\n", err.Error(), act.AnchorId)
			return nil, err
		case <-act.EthContext.Done():
			log.Errorf("Context closed before receiving AnchorRegistered event for anchor ID: %x\n", act.AnchorId)
			return nil, act.EthContext.Err()
		case res := <-act.AnchorRegisteredEvents:
			log.Infof("Received AnchorRegistered event from: %x, identifier: %x\n", res.From, res.Identifier)
			subscription.Unsubscribe()
			return res, nil
		}
	}
}

func get32Bytes(key interface{}) ([ThirtyTwo]byte, error) {
	var fixed [ThirtyTwo]byte
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

func getAddress(key interface{}) (common.Address, error) {
	// for safety always use length fixed bytes
	var fixed [common.AddressLength]byte
	b, ok := key.([]interface{})
	if !ok {
		return fixed, errors.New("Could not parse interface to []byte")
	}
	// convert and copy b byte values
	for i, v := range b {
		fv := v.(float64)
		fixed[i] = byte(fv)
	}
	return common.BytesToAddress(fixed[:]), nil
}
