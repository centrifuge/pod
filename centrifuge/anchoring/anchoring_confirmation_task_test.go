// +build unit

package anchoring

import (
	"math/big"
	"testing"

	"context"
	"errors"
	"time"

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
	from []common.Address, anchorId []*big.Int, centrifugeId []*big.Int) (event.Subscription, error) {
	if m.shouldFail {
		return nil, errors.New("Just a dummy error")
	}

	return m.Subscription, nil
}

func TestAnchoringConfirmationTask_ParseKwargsHappy(t *testing.T) {
	act := AnchoringConfirmationTask{}
	tmp := [32]byte{1, 2, 3}
	anchorBigInt := new(big.Int).SetBytes(tmp[:])
	address := common.BytesToAddress([]byte{1, 2, 3, 4})

	//convert big int to byte 32
	var byte32anchorId [32]byte
	copy(byte32anchorId[:],anchorBigInt.Bytes()[:32])

	var centrifugeIdBytes [6]byte

	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIdParam: byte32anchorId,
		AddressParam:  address,
		CentrifugeIdParam: centrifugeIdBytes,
	})
	err := act.ParseKwargs(kwargs)
	if err != nil {
		t.Fatalf("Could not parse %s or %s", AnchorIdParam, AddressParam)
	}

	//convert byte 32 to big int
	actBigInt := tools.ByteFixedToBigInt(act.AnchorId[:], 32)
	assert.Equal(t, anchorBigInt, actBigInt, "Resulting anchor Id should have the same ID as the input")
	assert.Equal(t, address, act.From, "Resulting address should have the same ID as the input")
}

func TestAnchoringConfirmationTask_ParseKwargsAnchorNotPassed(t *testing.T) {
	act := AnchoringConfirmationTask{}
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	var centrifugeIdBytes [6]byte

	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AddressParam: address,
		CentrifugeIdParam: centrifugeIdBytes,

	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "Anchor id should not have been parsed")
}

func TestAnchoringConfirmationTask_ParseKwargsInvalidAnchor(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorId := 123
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIdParam: anchorId,
		AddressParam:  address,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "Anchor id should not have been parsed because it was of incorrect type")
}

func TestAnchoringConfirmationTask_ParseKwargsAddressNotPassed(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorId := [32]byte{1, 2, 3}
	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIdParam: anchorId,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "address should not have been parsed")
}

func TestAnchoringConfirmationTask_ParseKwargsInvalidAddress(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorId := [32]byte{1, 2, 3}
	address := 123
	kwargs, _ := tools.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIdParam: anchorId,
		AddressParam:  address,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "address should not have been parsed because it was of incorrect type")
}

func TestAnchoringConfirmationTask_RunTaskContextClose(t *testing.T) {
	toBeDone := time.Now().Add(time.Duration(1 * time.Millisecond))
	ctx, _ := context.WithDeadline(context.TODO(), toBeDone)
	earcar := make(chan *EthereumAnchorRepositoryContractAnchorCommitted)
	anchorId := [32]byte{1, 2, 3}

	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	act := AnchoringConfirmationTask{
		AnchorId:                anchorId,
		From:                    address,
		AnchorCommittedWatcher: &MockAnchorCommittedWatcher{Subscription: &testingutils.MockSubscription{}},
		EthContext:              ctx,
		AnchorRegisteredEvents:  earcar,
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
	anchorId := [32]byte{1, 2, 3}
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	act := AnchoringConfirmationTask{
		AnchorId:                anchorId,
		From:                    address,
		AnchorCommittedWatcher: &MockAnchorCommittedWatcher{shouldFail: true},
		EthContext:              ctx,
		AnchorRegisteredEvents:  earcar,
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
	anchorId := [32]byte{1, 2, 3}
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	errChan := make(chan error)
	watcher := &MockAnchorCommittedWatcher{Subscription: &testingutils.MockSubscription{ErrChan: errChan}}
	act := AnchoringConfirmationTask{
		AnchorId:                anchorId,
		From:                    address,
		AnchorCommittedWatcher: watcher,
		EthContext:              ctx,
		AnchorRegisteredEvents:  earcar,
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
	anchorId := [32]byte{1, 2, 3}

	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	act := AnchoringConfirmationTask{
		AnchorId:                anchorId,
		From:                    address,
		AnchorCommittedWatcher: &MockAnchorCommittedWatcher{Subscription: &testingutils.MockSubscription{}},
		EthContext:              ctx,
		AnchorRegisteredEvents:  earcar,
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
