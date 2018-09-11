// +build unit

package identity

import (
	"testing"

	"context"
	"time"

	"errors"

	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/event"
	"github.com/stretchr/testify/assert"
)

type MockIdentityCreatedWatcher struct {
	shouldFail bool
	sink       chan<- *EthereumIdentityFactoryContractIdentityCreated
}

func (mcw *MockIdentityCreatedWatcher) WatchIdentityCreated(opts *bind.WatchOpts, sink chan<- *EthereumIdentityFactoryContractIdentityCreated, centrifugeId []*big.Int) (event.Subscription, error) {
	if mcw.shouldFail {
		return nil, errors.New("Identity watching could not be started")
	}
	mcw.sink = sink
	return &testingutils.MockSubscription{}, nil
}

func TestRegistrationConfirmationTask_ParseKwargsHappyPath(t *testing.T) {
	rct := IdRegistrationConfirmationTask{}
	id := tools.RandomSlice(CentIdByteLength)
	idBytes := NewCentId(id)
	kwargs := map[string]interface{}{CentIdParam: idBytes}
	decoded, err := tools.SimulateJsonDecodeForGocelery(kwargs)
	err = rct.ParseKwargs(decoded)
	if err != nil {
		t.Errorf("Could not parse %s for [%x]", CentIdParam, id)
	}
	assert.Equal(t, idBytes, rct.CentId, "Resulting mockID should have the same ID as the input")
}

func TestRegistrationConfirmationTask_ParseKwargsDoesNotExist(t *testing.T) {
	rct := IdRegistrationConfirmationTask{}
	id := tools.RandomSlice(CentIdByteLength)
	err := rct.ParseKwargs(map[string]interface{}{"notId": id})
	assert.NotNil(t, err, "Should not allow parsing without centId")
}

func TestRegistrationConfirmationTask_ParseKwargsInvalidType(t *testing.T) {
	rct := IdRegistrationConfirmationTask{}
	id := tools.RandomSlice(CentIdByteLength)
	err := rct.ParseKwargs(map[string]interface{}{CentIdParam: id})
	assert.NotNil(t, err, "Should not parse without the correct type of centId")
}

func TestIdRegistrationConfirmationTask_RunTaskContextError(t *testing.T) {
	toBeDone := time.Now().Add(time.Duration(1 * time.Millisecond))
	ctx, _ := context.WithDeadline(context.TODO(), toBeDone)
	eifc := make(chan *EthereumIdentityFactoryContractIdentityCreated)
	rct := IdRegistrationConfirmationTask{
		CentId:                 NewCentId(tools.RandomSlice(CentIdByteLength)),
		IdentityCreatedWatcher: &MockIdentityCreatedWatcher{},
		EthContext:             ctx,
		IdentityCreatedEvents:  eifc,
	}
	exit := make(chan bool)
	go func() {
		_, err := rct.RunTask()
		assert.NotNil(t, err)
		exit <- true
	}()
	time.Sleep(1 * time.Millisecond)
	// this would cause an error exit in the task
	ctx.Done()
	<-exit
}

func TestIdRegistrationConfirmationTask_RunTaskCallError(t *testing.T) {
	identityCreatedWatcher := &MockIdentityCreatedWatcher{shouldFail: true}
	rct := IdRegistrationConfirmationTask{
		CentId: NewCentId(tools.RandomSlice(CentIdByteLength)),
		EthContextInitializer: func() (ctx context.Context, cancelFunc context.CancelFunc) {
			toBeDone := time.Now().Add(time.Duration(1 * time.Millisecond))
			return context.WithDeadline(context.TODO(), toBeDone)
		},
		IdentityCreatedWatcher: identityCreatedWatcher,
	}
	exit := make(chan bool)
	go func() {
		_, err := rct.RunTask()
		assert.NotNil(t, err)
		exit <- true
	}()
	time.Sleep(1 * time.Millisecond)
	<-exit
}

func TestIdRegistrationConfirmationTask_RunTaskSuccess(t *testing.T) {
	toBeDone := time.Now().Add(time.Duration(1 * time.Second))
	ctx, _ := context.WithDeadline(context.TODO(), toBeDone)
	eifc := make(chan *EthereumIdentityFactoryContractIdentityCreated)
	rct := IdRegistrationConfirmationTask{
		CentId:                 NewCentId(tools.RandomSlice(CentIdByteLength)),
		IdentityCreatedWatcher: &MockIdentityCreatedWatcher{},
		EthContext:             ctx,
		IdentityCreatedEvents:  eifc,
	}
	exit := make(chan bool)
	go func() {
		res, err := rct.RunTask()
		assert.Nil(t, err)
		assert.NotNil(t, res)
		exit <- true
	}()
	time.Sleep(1 * time.Millisecond)
	// this would cause a successful exit in the task
	eifc <- &EthereumIdentityFactoryContractIdentityCreated{}
	<-exit
}
