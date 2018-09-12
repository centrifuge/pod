// +build unit

package anchors

import (
	"math/big"
	"testing"

	"context"
	"errors"
	"time"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/stretchr/testify/assert"
)

type MockAnchorCommittedWatcher struct {
	shouldFail   bool
	sink         chan<- *EthereumAnchorRepositoryContractAnchorCommitted
	Subscription event.Subscription
}

func (m *MockAnchorCommittedWatcher) WatchAnchorCommitted(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRepositoryContractAnchorCommitted,
	from []common.Address, anchorID []*big.Int, centrifugeId []*big.Int) (event.Subscription, error) {
	if m.shouldFail {
		return nil, errors.New("Just a dummy error")
	}

	return m.Subscription, nil
}

func TestAnchoringConfirmationTask_ParseKwargsHappy(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorID, _ := NewAnchorID(tools.RandomSlice(AnchorIDLength))
	address := common.BytesToAddress([]byte{1, 2, 3, 4})

	centId, _ := identity.NewCentID(tools.RandomSlice(identity.CentIDByteLength))

	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIDParam:     anchorID,
		AddressParam:      address,
		CentrifugeIDParam: centId,
	})
	err := act.ParseKwargs(kwargs)
	if err != nil {
		t.Fatalf("Could not parse %s or %s", AnchorIDParam, AddressParam)
	}

	//convert byte 32 to big int
	assert.Equal(t, anchorID, anchorID, "Resulting anchor Id should have the same ID as the input")
	assert.Equal(t, address, act.From, "Resulting address should have the same ID as the input")
	assert.Equal(t, centId, act.CentrifugeID, "Resulting centId should have the same centId as the input")
}

func TestAnchoringConfirmationTask_ParseKwargsAnchorNotPassed(t *testing.T) {
	act := AnchoringConfirmationTask{}
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	var centrifugeIdBytes [6]byte

	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AddressParam:      address,
		CentrifugeIDParam: centrifugeIdBytes,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "Anchor id should not have been parsed")
}

func TestAnchoringConfirmationTask_ParseKwargsInvalidAnchor(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorID := 123
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIDParam: anchorID,
		AddressParam:  address,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "Anchor id should not have been parsed because it was of incorrect type")
}

func TestAnchoringConfirmationTask_ParseKwargsAddressNotPassed(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorID := [32]byte{1, 2, 3}
	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIDParam: anchorID,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "address should not have been parsed")
}

func TestAnchoringConfirmationTask_ParseKwargsInvalidAddress(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorID := [32]byte{1, 2, 3}
	address := 123
	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIDParam: anchorID,
		AddressParam:  address,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "address should not have been parsed because it was of incorrect type")
}

func TestAnchoringConfirmationTask_RunTaskContextClose(t *testing.T) {
	toBeDone := time.Now().Add(time.Duration(1 * time.Millisecond))
	ctx, _ := context.WithDeadline(context.TODO(), toBeDone)
	earcar := make(chan *EthereumAnchorRepositoryContractAnchorCommitted)
	anchorID := [32]byte{1, 2, 3}

	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	act := AnchoringConfirmationTask{
		AnchorID: anchorID,
		From:     address,
		AnchorCommittedWatcher: &MockAnchorCommittedWatcher{Subscription: &testingutils.MockSubscription{}},
		EthContext:             ctx,
		AnchorRegisteredEvents: earcar,
	}
	exit := make(chan bool)
	go func() {
		_, err := act.RunTask()
		assert.NotNil(t, err)
		exit <- true
	}()
	time.Sleep(1 * time.Millisecond)
	// this would cause an error exit in the task
	ctx.Done()
	<-exit
}

func TestAnchoringConfirmationTask_RunTaskWatchError(t *testing.T) {
	toBeDone := time.Now().Add(time.Duration(1 * time.Second))
	ctx, _ := context.WithDeadline(context.TODO(), toBeDone)
	earcar := make(chan *EthereumAnchorRepositoryContractAnchorCommitted)
	anchorID := [32]byte{1, 2, 3}
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	act := AnchoringConfirmationTask{
		AnchorID: anchorID,
		From:     address,
		AnchorCommittedWatcher: &MockAnchorCommittedWatcher{shouldFail: true},
		EthContext:             ctx,
		AnchorRegisteredEvents: earcar,
	}
	exit := make(chan bool)
	go func() {
		_, err := act.RunTask()
		assert.NotNil(t, err)
		exit <- true
	}()
	time.Sleep(1 * time.Millisecond)
	<-exit
}

func TestAnchoringConfirmationTask_RunTaskSubscriptionError(t *testing.T) {
	toBeDone := time.Now().Add(time.Duration(1 * time.Second))
	ctx, _ := context.WithDeadline(context.TODO(), toBeDone)
	earcar := make(chan *EthereumAnchorRepositoryContractAnchorCommitted)
	anchorID := [32]byte{1, 2, 3}
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	errChan := make(chan error)
	watcher := &MockAnchorCommittedWatcher{Subscription: &testingutils.MockSubscription{ErrChan: errChan}}
	act := AnchoringConfirmationTask{
		AnchorID: anchorID,
		From:     address,
		AnchorCommittedWatcher: watcher,
		EthContext:             ctx,
		AnchorRegisteredEvents: earcar,
	}
	exit := make(chan bool)
	go func() {
		_, err := act.RunTask()
		assert.NotNil(t, err)
		exit <- true
	}()
	time.Sleep(1 * time.Millisecond)
	errChan <- errors.New("Dummy subscription error")
	<-exit
}

func TestAnchoringConfirmationTask_RunTaskSuccess(t *testing.T) {
	toBeDone := time.Now().Add(time.Duration(1 * time.Second))
	ctx, _ := context.WithDeadline(context.TODO(), toBeDone)
	earcar := make(chan *EthereumAnchorRepositoryContractAnchorCommitted)
	anchorID := [32]byte{1, 2, 3}

	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	act := AnchoringConfirmationTask{
		AnchorID: anchorID,
		From:     address,
		AnchorCommittedWatcher: &MockAnchorCommittedWatcher{Subscription: &testingutils.MockSubscription{}},
		EthContext:             ctx,
		AnchorRegisteredEvents: earcar,
	}
	exit := make(chan bool)
	go func() {
		res, err := act.RunTask()
		assert.Nil(t, err)
		assert.NotNil(t, res)
		exit <- true
	}()
	time.Sleep(1 * time.Millisecond)
	act.AnchorRegisteredEvents <- &EthereumAnchorRepositoryContractAnchorCommitted{}
	<-exit
}