// +build unit

package anchor

import (
	"errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cc.TestBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

func TestGenerateAnchor(t *testing.T) {
	anchor, err := generateAnchor("ABCD", "DCBA")
	assert.Nil(t, anchor)
	assert.Error(t, err, "Should fail with too short input values")

	anchor, err = generateAnchor("01234567890123456789012345678932", "DCBA")
	assert.Nil(t, anchor)
	assert.Error(t, err, "Should fail with too short input values")

	anchor, err = generateAnchor("012345678901234567890123456789333", "012345678901234567890123456789333")
	assert.Nil(t, anchor)
	assert.Error(t, err, "Should fail with too long input values")

	anchor, err = generateAnchor("01234567890123456789012345678932", "012345678901234567890123456789333")
	assert.Nil(t, anchor)
	assert.Error(t, err, "Should fail with too long input values")

	anchor, err = generateAnchor("0123456789012345678901234567893A", "0123456789012345678901234567893B")
	assert.Nil(t, err)
	assert.Equal(t, anchor.AnchorID, "0123456789012345678901234567893A", "Anchor should have the passed ID")
	assert.Equal(t, anchor.RootHash, "0123456789012345678901234567893B", "Anchor should have the passed root hash")
	assert.Equal(t, anchor.SchemaVersion, SupportedSchemaVersion(), "Anchor should have the supported schema version")
}

type MockRegisterAnchor struct {
	shouldFail bool
}

func (mra *MockRegisterAnchor) RegisterAnchor(opts *bind.TransactOpts, identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error) {
	if mra.shouldFail == true {
		return nil, errors.New("for testing - error if identifier == merkleRoot")
	}
	hashableTransaction := types.NewTransaction(1, common.StringToAddress("0x0000000000000000001"), big.NewInt(1000), 1000, big.NewInt(1000), nil)

	return hashableTransaction, nil
}

type MockWatchAnchorRegistered struct {
	shouldFail bool
	sink       chan<- *EthereumAnchorRegistryContractAnchorRegistered
}

func (mwar *MockWatchAnchorRegistered) WatchAnchorRegistered(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRegistryContractAnchorRegistered, from []common.Address, identifier [][32]byte, rootHash [][32]byte) (event.Subscription, error) {
	if mwar.shouldFail == true {
		return nil, errors.New("forced error during test")
	} else {
		if sink != nil {
			mwar.sink = sink
		}
		return nil, nil
	}
}

func TestSendRegistrationTransaction_ErrorPassThrough(t *testing.T) {
	anchor := Anchor{tools.RandomString32(), tools.RandomString32(), 1}
	failingAnchorRegistrar := &MockRegisterAnchor{shouldFail: true}

	err := sendRegistrationTransaction(failingAnchorRegistrar, nil, &anchor)
	assert.Error(t, err, "Should have an error if registerAnchor returns error")
}

func TestSendRegistrationTransaction_InputParams(t *testing.T) {
	anchor := Anchor{"someId", "someRootHash", 1}

	err := sendRegistrationTransaction(&MockRegisterAnchor{}, nil, &anchor)
	assert.Contains(t, err.Error(), "32")
	anchor.AnchorID = tools.RandomString32()

	err = sendRegistrationTransaction(&MockRegisterAnchor{}, nil, &anchor)
	assert.Contains(t, err.Error(), "32")
	anchor.RootHash = tools.RandomString32()

	err = sendRegistrationTransaction(&MockRegisterAnchor{}, nil, &anchor)
	assert.Nil(t, err, "All inputs should validate now")
}

func TestSetUpRegistrationEventListener_ErrorPassThrough(t *testing.T) {
	resetMock := testingutils.MockConfigOption("ethereum.contextWaitTimeout", "30s")

	failingWatchAnchorRegistered := &MockWatchAnchorRegistered{shouldFail: true}
	anchor := Anchor{tools.RandomString32(), tools.RandomString32(), 1}
	confirmations := make(chan *Anchor)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Code should have paniced if the subscription to confirmation channel failed")
		}
	}()

	err := setUpRegistrationEventListener(failingWatchAnchorRegistered, common.Address{}, &anchor, confirmations)

	defer resetMock()
	assert.Error(t, err, "Should fail if the anchor registration watcher failed")
}

func TestSetUpRegistrationEventListener_ChannelSubscriptionCreated(t *testing.T) {
	config.Config.V.Set("ethereum.contextWaitTimeout", "30s")
	mockWatchAnchorRegistered := &MockWatchAnchorRegistered{}
	anchor := Anchor{tools.RandomString32(), tools.RandomString32(), 1}
	confirmations := make(chan *Anchor, 1)
	err := setUpRegistrationEventListener(mockWatchAnchorRegistered, common.Address{}, &anchor, confirmations)
	assert.Nil(t, err, "Should not fail")
	//sending one "event" into the registered sink should result in the confirmations channel to receive the anchor
	//that has been created and passed through initially
	b32Id, _ := tools.StringToByte32(anchor.AnchorID)
	mockWatchAnchorRegistered.sink <- &EthereumAnchorRegistryContractAnchorRegistered{From: common.StringToAddress("0x0000000000000000001"), Identifier: b32Id}
	receivedAnchor := <-confirmations
	assert.Equal(t, anchor.AnchorID, receivedAnchor.AnchorID, "Received anchor should have the same data as the originally submitted anchor")
}
