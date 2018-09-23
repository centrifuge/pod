// +build unit

package identity_test

import (
	"testing"

	"errors"

	"math/big"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/event"
	"github.com/stretchr/testify/assert"
)

type MockIdentityCreatedWatcher struct {
	shouldFail bool
	sink       chan<- *identity.EthereumIdentityFactoryContractIdentityCreated
}

func (mcw *MockIdentityCreatedWatcher) WatchIdentityCreated(opts *bind.WatchOpts, sink chan<- *identity.EthereumIdentityFactoryContractIdentityCreated, centrifugeId []*big.Int) (event.Subscription, error) {
	if mcw.shouldFail {
		return nil, errors.New("Identity watching could not be started")
	}
	mcw.sink = sink
	return &testingutils.MockSubscription{}, nil
}

func TestRegistrationConfirmationTask_ParseKwargsHappyPath(t *testing.T) {
	rct := identity.IdRegistrationConfirmationTask{}
	id := tools.RandomSlice(identity.CentIDByteLength)
	idBytes, _ := identity.NewCentID(id)
	kwargs := map[string]interface{}{identity.CentIdParam: idBytes}
	decoded, err := tools.SimulateJsonDecodeForGocelery(kwargs)
	err = rct.ParseKwargs(decoded)
	if err != nil {
		t.Errorf("Could not parse %s for [%x]", identity.CentIdParam, id)
	}
	assert.Equal(t, idBytes, rct.CentID, "Resulting mockID should have the same ID as the input")
}

func TestRegistrationConfirmationTask_ParseKwargsDoesNotExist(t *testing.T) {
	rct := identity.IdRegistrationConfirmationTask{}
	id := tools.RandomSlice(identity.CentIDByteLength)
	err := rct.ParseKwargs(map[string]interface{}{"notId": id})
	assert.NotNil(t, err, "Should not allow parsing without centId")
}

func TestRegistrationConfirmationTask_ParseKwargsInvalidType(t *testing.T) {
	rct := identity.IdRegistrationConfirmationTask{}
	id := tools.RandomSlice(identity.CentIDByteLength)
	err := rct.ParseKwargs(map[string]interface{}{identity.CentIdParam: id})
	assert.NotNil(t, err, "Should not parse without the correct type of centId")
}

//func TestIdRegistrationConfirmationTask_RunTaskContextError(t *testing.T) {
//	cenId, _ := identity.NewCentID(tools.RandomSlice(identity.CentIDByteLength))
//	toBeDone := time.Now().Add(time.Duration(1 * time.Millisecond))
//	ctx, _ := context.WithDeadline(context.TODO(), toBeDone)
//	eifc := make(chan *identity.EthereumIdentityFactoryContractIdentityCreated)
//	rct := identity.IdRegistrationConfirmationTask{
//		CentID:                 cenId,
//		IdentityCreatedWatcher: &MockIdentityCreatedWatcher{},
//		EthContext:             ctx,
//		IdentityCreatedEvents:  eifc,
//	}
//	exit := make(chan bool)
//	go func() {
//		_, err := rct.RunTask()
//		assert.NotNil(t, err)
//		exit <- true
//	}()
//	time.Sleep(1 * time.Millisecond)
//	// this would cause an error exit in the task
//	ctx.Done()
//	<-exit
//}
//
//func TestIdRegistrationConfirmationTask_RunTaskCallError(t *testing.T) {
//	cenId, _ := identity.NewCentID(tools.RandomSlice(identity.CentIDByteLength))
//	identityCreatedWatcher := &MockIdentityCreatedWatcher{shouldFail: true}
//	rct := identity.IdRegistrationConfirmationTask{
//		CentID: cenId,
//		EthContextInitializer: func() (ctx context.Context, cancelFunc context.CancelFunc) {
//			toBeDone := time.Now().Add(time.Duration(1 * time.Millisecond))
//			return context.WithDeadline(context.TODO(), toBeDone)
//		},
//		IdentityCreatedWatcher: identityCreatedWatcher,
//	}
//	exit := make(chan bool)
//	go func() {
//		_, err := rct.RunTask()
//		assert.NotNil(t, err)
//		exit <- true
//	}()
//	time.Sleep(1 * time.Millisecond)
//	<-exit
//}
//
//func TestIdRegistrationConfirmationTask_RunTaskSuccess(t *testing.T) {
//	cenId, _ := identity.NewCentID(tools.RandomSlice(identity.CentIDByteLength))
//	toBeDone := time.Now().Add(time.Duration(1 * time.Second))
//	ctx, _ := context.WithDeadline(context.TODO(), toBeDone)
//	eifc := make(chan *identity.EthereumIdentityFactoryContractIdentityCreated)
//	rct := identity.IdRegistrationConfirmationTask{
//		CentID:                 cenId,
//		IdentityCreatedWatcher: &MockIdentityCreatedWatcher{},
//		EthContext:             ctx,
//		IdentityCreatedEvents:  eifc,
//	}
//	exit := make(chan bool)
//	go func() {
//		res, err := rct.RunTask()
//		assert.Nil(t, err)
//		assert.NotNil(t, res)
//		exit <- true
//	}()
//	time.Sleep(1 * time.Millisecond)
//	// this would cause a successful exit in the task
//	eifc <- &identity.EthereumIdentityFactoryContractIdentityCreated{}
//	<-exit
//}
