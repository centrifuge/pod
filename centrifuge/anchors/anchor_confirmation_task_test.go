// +build unit

package anchors

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

type MockAnchorCommittedFilter struct {
	iter *EthereumAnchorRepositoryContractAnchorCommittedIterator
	err  error
}

func (m *MockAnchorCommittedFilter) FilterAnchorCommitted(
	opts *bind.FilterOpts,
	from []common.Address,
	anchorID []*big.Int,
	centrifugeID []*big.Int) (*EthereumAnchorRepositoryContractAnchorCommittedIterator, error) {

	return m.iter, m.err
}

func TestAnchoringConfirmationTask_ParseKwargsHappy(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorID, _ := NewAnchorID(utils.RandomSlice(AnchorIDLength))
	address := common.BytesToAddress([]byte{1, 2, 3, 4})

	centId, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))

	kwargs, _ := utils.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIDParam:     anchorID,
		AddressParam:      address,
		CentrifugeIDParam: centId,
		BlockHeight:       float64(0),
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

	kwargs, _ := utils.SimulateJsonDecodeForGocelery(map[string]interface{}{
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
	kwargs, _ := utils.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIDParam: anchorID,
		AddressParam:  address,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "Anchor id should not have been parsed because it was of incorrect type")
}

func TestAnchoringConfirmationTask_ParseKwargsAddressNotPassed(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorID := [32]byte{1, 2, 3}
	kwargs, _ := utils.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIDParam: anchorID,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "address should not have been parsed")
}

func TestAnchoringConfirmationTask_ParseKwargsInvalidAddress(t *testing.T) {
	act := AnchoringConfirmationTask{}
	anchorID := [32]byte{1, 2, 3}
	address := 123
	kwargs, _ := utils.SimulateJsonDecodeForGocelery(map[string]interface{}{
		AnchorIDParam: anchorID,
		AddressParam:  address,
	})
	err := act.ParseKwargs(kwargs)
	assert.NotNil(t, err, "address should not have been parsed because it was of incorrect type")
}

func TestAnchoringConfirmationTask_RunTaskIterError(t *testing.T) {
	anchorID := [32]byte{1, 2, 3}
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	act := AnchoringConfirmationTask{
		AnchorID:                anchorID,
		From:                    address,
		AnchorCommittedFilterer: &MockAnchorCommittedFilter{err: fmt.Errorf("failed iterator")},
		EthContext:              context.Background(),
	}

	_, err := act.RunTask()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed iterator")
}

func TestAnchoringConfirmationTask_RunTaskWatchError(t *testing.T) {
	toBeDone := time.Now().Add(time.Duration(1 * time.Millisecond))
	ctx, _ := context.WithDeadline(context.Background(), toBeDone)
	anchorID := [32]byte{1, 2, 3}
	address := common.BytesToAddress([]byte{1, 2, 3, 4})
	act := AnchoringConfirmationTask{
		AnchorID: anchorID,
		From:     address,
		AnchorCommittedFilterer: &MockAnchorCommittedFilter{iter: &EthereumAnchorRepositoryContractAnchorCommittedIterator{
			fail: fmt.Errorf("watch error"),
			sub:  &testingutils.MockSubscription{},
		}},
		EthContext: ctx,
	}

	_, err := act.RunTask()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "watch error")
}
