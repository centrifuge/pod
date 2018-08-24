package identity

import (
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/go-errors/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
)

const RegistrationConfirmationTaskName string = "RegistrationConfirmationTaskName"
const CentIdParam string = "CentId"

type RegistrationConfirmationTask struct {
	CentId [32]byte
}

func (rct *RegistrationConfirmationTask) Name() string {
	return RegistrationConfirmationTaskName
}

func (rct *RegistrationConfirmationTask) Init() error {
	queue.Queue.Register(RegistrationConfirmationTaskName, rct)
	return nil
}

// ParseKwargs - define a method to parse CentId
// TODO test this !!!!
func (rct *RegistrationConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
	CentId, ok := kwargs[CentIdParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + CentIdParam)
	}
	CentIdTyped, ok := CentId.([32]byte)
	if !ok {
		return fmt.Errorf("malformed kwarg " + CentIdParam)
	}
	rct.CentId = CentIdTyped
	return nil
}

// RunTask - define a method to run
func (rct *RegistrationConfirmationTask) RunTask() (interface{}, error) {
	ctx, cancelFunc := ethereum.DefaultWaitForTransactionMiningContext()
	watchOpts := &bind.WatchOpts{Context: ctx}
	contract, err := getIdentityFactoryContract()
	identityCreatedEvents := make(chan *EthereumIdentityFactoryContractIdentityCreated)

	//TODO do something with the returned Subscription that is currently simply discarded
	// Somehow there are some possible resource leakage situations with this handling but I have to understand
	// Subscriptions a bit better before writing this code.
	_, err = contract.WatchIdentityCreated(watchOpts, identityCreatedEvents, [][32]byte{rct.CentId})
	if err != nil {
		wError := errors.WrapPrefix(err, "Could not subscribe to event logs for identity registration", 1)
		log.Errorf(wError.Error())
		cancelFunc()
		return nil, wError
	}
	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context [%v] closed before receiving KeyRegistered event for Identity ID: %x\n", ctx, rct.CentId)
			return nil, ctx.Err()
		case res := <-identityCreatedEvents:
			log.Infof("Received IdentityCreated event from: %x, identifier: %x\n", res.CentrifugeId, res.Identity)
			return res, nil
		}
	}
}
