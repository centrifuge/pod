package identity

import (
	"context"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/queue"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/go-errors/errors"
)

const (
	IdRegistrationConfirmationTaskName string = "IdRegistrationConfirmationTaskName"
	CentIdParam                        string = "CentID"
	BlockHeight                        string = "BlockHeight"
)

type IdentitiesCreatedFilterer interface {
	FilterIdentityCreated(opts *bind.FilterOpts, centrifugeId []*big.Int) (*EthereumIdentityFactoryContractIdentityCreatedIterator, error)
}

// IdRegistrationConfirmationTask is a queued task to watch ID registration events on Ethereum using EthereumIdentityFactoryContract.
// To see how it gets registered see bootstrapper.go and to see how it gets used see setUpRegistrationEventListener method
type IdRegistrationConfirmationTask struct {
	CentID                 CentID
	BlockHeight            uint64
	EthContextInitializer  func() (ctx context.Context, cancelFunc context.CancelFunc)
	EthContext             context.Context
	IdentityCreatedWatcher IdentitiesCreatedFilterer
}

func NewIdRegistrationConfirmationTask(
	identityCreatedWatcher IdentitiesCreatedFilterer,
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

func (rct *IdRegistrationConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &IdRegistrationConfirmationTask{
		rct.CentID,
		rct.BlockHeight,
		rct.EthContextInitializer,
		rct.EthContext,
		rct.IdentityCreatedWatcher}, nil
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

	rct.CentID = centIdTyped
	rct.BlockHeight = uint64(kwargs[BlockHeight].(float64))
	return nil
}

func getEvent(iter *EthereumIdentityFactoryContractIdentityCreatedIterator) (*EthereumIdentityFactoryContractIdentityCreated, bool, error) {
	defer iter.Close()
	if iter.Next() {
		return iter.Event, false, nil
	}

	err := iter.Error()
	if err != nil {
		return nil, false, err
	}

	return nil, true, fmt.Errorf("no matching events found")
}

// RunTask calls listens to events from geth related to IdRegistrationConfirmationTask#CentID and records result.
func (rct *IdRegistrationConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the ID [%x]", rct.CentID.ByteArray())
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

		res, proceed, err := getEvent(iter)
		if err == nil || !proceed {
			return res, err
		}
	}

	return nil, fmt.Errorf("failed to filter identity events")
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
