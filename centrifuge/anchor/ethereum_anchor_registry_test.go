package anchor

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"math/big"
	"github.com/ethereum/go-ethereum/core/types"
	"errors"
	"reflect"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/common"
)

func TestGenerateAnchor(t *testing.T) {

	anchor := generateAnchor("ABCD", "DCBA")
	assert.Equal(t, anchor.AnchorID, "ABCD", "Anchor should have the passed ID")
	assert.Equal(t, anchor.RootHash, "DCBA", "Anchor should have the passed root hash")
	assert.Equal(t, anchor.SchemaVersion, SupportedSchemaVersion(), "Anchor should have the supported schema version")
}

type MockRegisterAnchor struct{}

func (mra *MockRegisterAnchor) RegisterAnchor(opts *bind.TransactOpts, identifier [32]byte, merkleRoot [32]byte, anchorSchemaVersion *big.Int) (*types.Transaction, error) {
	if reflect.DeepEqual(identifier, merkleRoot) {
		return nil, errors.New("for testing - error if identified == merkleRoot")
	}
	hashableTransaction := types.NewTransaction(1, common.StringToAddress("0x0000000000000000001"),big.NewInt(1000),1000,big.NewInt(1000),nil)

	return hashableTransaction, nil
}



func TestSendRegistrationTransaction(t *testing.T) {
	anchor := Anchor{"someId", "someRootHash", 1}

	err := sendRegistrationTransaction(&MockRegisterAnchor{}, nil, &anchor)
	assert.Contains(t, err.Error(), "32")
	anchor.AnchorID = tools.RandomString32()

	err = sendRegistrationTransaction(&MockRegisterAnchor{}, nil, &anchor)
	assert.Contains(t, err.Error(), "32")
	anchor.RootHash = tools.RandomString32()

	err = sendRegistrationTransaction(&MockRegisterAnchor{}, nil, &anchor)
	assert.Nil(t, err, "All inputs should validate now")

	//sameString := tools.RandomString32()
	//anchor2 := Anchor{"sameString", "sameString",1}
	//err2 := sendRegistrationTransaction(&MockRegisterAnchor{}, nil, &anchor2 )
	//assert.Error(t, err2, "Should have an error if registerAnchor returns error")
}
