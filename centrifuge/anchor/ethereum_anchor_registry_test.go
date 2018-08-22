// +build unit

package anchor

import (
	"math/big"
	"os"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	config.Config.V.Set("ethereum.txPoolAccessEnabled", false)
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

// ----- MOCKING -----
type MockRegisterAnchor struct {
	shouldFail bool
}

func (mra *MockRegisterAnchor) RegisterAnchor(opts *bind.TransactOpts, identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error) {
	if mra.shouldFail == true {
		return nil, errors.New("for testing - error if identifier == merkleRoot")
	}
	hashableTransaction := types.NewTransaction(1, common.HexToAddress("0x0000000000000000001"), big.NewInt(1000), 1000, big.NewInt(1000), nil)

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

// END ----- MOCKING -----

func TestGenerateAnchor(t *testing.T) {
	anchorID := tools.RandomByte32()
	rootHash := tools.RandomByte32()

	anchor, err := generateAnchor(anchorID, rootHash)
	assert.Nil(t, err)
	assert.Equal(t, anchor.AnchorID, anchorID, "Anchor should have the passed ID")
	assert.Equal(t, anchor.RootHash, rootHash, "Anchor should have the passed root hash")
	assert.Equal(t, anchor.SchemaVersion, SupportedSchemaVersion(), "Anchor should have the supported schema version")
}

//started building this as table based test
//TODO build the rest of the suite like this and makre more unit-testable
var registerAsAnchorData = []struct {
	id       [32]byte
	hs       [32]byte
	chn      chan<- *WatchAnchor
	expected error // expected result
}{
	{[32]byte{}, [32]byte{'1'}, nil, errors.New("Can not work with empty anchor ID")},
	{[32]byte{'1'}, [32]byte{}, nil, errors.New("Can not work with empty root hash")},
}

func TestRegisterAsAnchor(t *testing.T) {
	for _, tt := range registerAsAnchorData {
		_, actual := new(EthereumAnchorRegistry).RegisterAsAnchor(tt.id, tt.hs)
		assert.Equal(t, tt.expected.Error(), actual.Error())
	}
}

func TestSendRegistrationTransaction_ErrorPassThrough(t *testing.T) {
	anchor := Anchor{tools.RandomByte32(), tools.RandomByte32(), 1}
	failingAnchorRegistrar := &MockRegisterAnchor{shouldFail: true}

	err := sendRegistrationTransaction(failingAnchorRegistrar, nil, &anchor)
	assert.Error(t, err, "Should have an error if registerAnchor returns error")
}

func TestSetUpRegistrationEventListener_ErrorPassThrough(t *testing.T) {
	failingWatchAnchorRegistered := &MockWatchAnchorRegistered{shouldFail: true}
	anchor := Anchor{tools.RandomByte32(), tools.RandomByte32(), 1}

	_, err := setUpRegistrationEventListener(failingWatchAnchorRegistered, common.Address{}, &anchor)
	assert.Error(t, err, "Should fail if the anchor registration watcher failed")
}

func TestSetUpRegistrationEventListener_ChannelSubscriptionCreated(t *testing.T) {
	config.Config.V.Set("ethereum.contextWaitTimeout", "30s")
	mockWatchAnchorRegistered := &MockWatchAnchorRegistered{}
	anchor := Anchor{tools.RandomByte32(), tools.RandomByte32(), 1}
	confirmations, err := setUpRegistrationEventListener(mockWatchAnchorRegistered, common.Address{}, &anchor)
	assert.Nil(t, err, "Should not fail")
	//sending one "event" into the registered sink should result in the confirmations channel to receive the anchor
	//that has been created and passed through initially
	mockWatchAnchorRegistered.sink <- &EthereumAnchorRegistryContractAnchorRegistered{From: common.HexToAddress("0x0000000000000000001"), Identifier: anchor.AnchorID}
	watchReceivedAnchor := <-confirmations
	assert.Equal(t, anchor.AnchorID, watchReceivedAnchor.Anchor.AnchorID, "Received anchor should have the same data as the originally submitted anchor")
}
