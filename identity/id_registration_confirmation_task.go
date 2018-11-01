package identity

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

const (
	idRegistrationConfirmationTaskName string = "IDRegistrationConfirmationTaskName"
)

type identitiesCreatedFilterer interface {
	FilterIdentityCreated(opts *bind.FilterOpts, centrifugeId []*big.Int) (*EthereumIdentityFactoryContractIdentityCreatedIterator, error)
}

// idRegistrationConfirmationTask is a queued task to watch ID registration events on Ethereum using EthereumIdentityFactoryContract.
// To see how it gets registered see bootstrapper.go and to see how it gets used see setUpRegistrationEventListener method
type idRegistrationConfirmationTask struct {
	CentID                 CentID
	BlockHeight            uint64
	EthContextInitializer  func() (ctx context.Context, cancelFunc context.CancelFunc)
	EthContext             context.Context
	IdentityCreatedWatcher identitiesCreatedFilterer
}

func newIdRegistrationConfirmationTask(
	identityCreatedWatcher identitiesCreatedFilterer,
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc),
) *idRegistrationConfirmationTask {
	return &idRegistrationConfirmationTask{
		IdentityCreatedWatcher: identityCreatedWatcher,
		EthContextInitializer:  ethContextInitializer,
	}
}

// Name returns the name of the task
func (rct *idRegistrationConfirmationTask) Name() string {
	return idRegistrationConfirmationTaskName
}

// Init registers the task to queue
func (rct *idRegistrationConfirmationTask) Init() error {
	queue.Queue.Register(idRegistrationConfirmationTaskName, rct)
	return nil
}

// Copy returns a new copy of the the task
func (rct *idRegistrationConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &idRegistrationConfirmationTask{
		rct.CentID,
		rct.BlockHeight,
		rct.EthContextInitializer,
		rct.EthContext,
		rct.IdentityCreatedWatcher}, nil
}

// ParseKwargs parses the kwargs into the task
func (rct *idRegistrationConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
	centId, ok := kwargs[centIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + centIDParam)
	}
	centIdTyped, err := getCentID(centId)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", centIDParam, err.Error())
	}
	rct.CentID = centIdTyped

	rct.BlockHeight, err = parseBlockHeight(kwargs)
	if err != nil {
		return err
	}
	return nil
}

// RunTask calls listens to events from geth related to idRegistrationConfirmationTask#CentID and records result.
func (rct *idRegistrationConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the ID [%x]", rct.CentID)
	if rct.EthContext == nil {
		rct.EthContext, _ = rct.EthContextInitializer()
	}

	fOpts := &bind.FilterOpts{
		Context: rct.EthContext,
		Start:   rct.BlockHeight,
	}

	for {
		iter, err := rct.IdentityCreatedWatcher.FilterIdentityCreated(
			fOpts,
			[]*big.Int{rct.CentID.BigInt()},
		)
		if err != nil {
			return nil, centerrors.Wrap(err, "failed to start filtering identity event logs")
		}

		err = utils.LookForEvent(iter)
		if err == nil {
			log.Infof("Received filtered event Id Registration Confirmation for CentrifugeID [%s]\n", rct.CentID.String())
			return iter.Event, nil
		}

		if err != utils.EventNotFound {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to filter identity events")
}
