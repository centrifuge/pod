// +build unit

package documents

import (
	"encoding/json"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAttribute_isAttrTypeAllowed(t *testing.T) {
	tests := []struct {
		attrType AttributeType
		result   bool
	}{
		{
			attrType: AttrDecimal,
			result:   true,
		},

		{
			attrType: AttributeType("some type"),
			result:   false,
		},
	}

	for _, c := range tests {
		assert.Equal(t, c.result, isAttrTypeAllowed(c.attrType))
	}
}

func TestNewAttribute(t *testing.T) {
	tests := []struct {
		name        string
		readableKey string
		attrType    AttributeType
		value       string
		errs        bool
		errStr      string
	}{
		{
			"readable key empty",
			"",
			AttrString,
			"",
			true,
			ErrEmptyAttrLabel.Error(),
		},

		{
			"type not allowed",
			"somekey",
			"some type",
			"",
			true,
			ErrNotValidAttrType.Error(),
		},

		{
			"string",
			"string",
			AttrString,
			"someval",
			false,
			"",
		},
		{
			"int256",
			"int256",
			AttrInt256,
			"123",
			false,
			"",
		},
		{
			"bigdecimal",
			"bigdecimal",
			AttrDecimal,
			"5.1321312",
			false,
			"",
		},
		{
			"bytes",
			"bytes",
			AttrBytes,
			hexutil.Encode([]byte{1}),
			false,
			"",
		},
		{
			"timestamp",
			"timestamp",
			AttrTimestamp,
			time.Now().UTC().Format(time.RFC3339),
			false,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attr, err := NewStringAttribute(test.readableKey, test.attrType, test.value)
			if test.errs {
				assert.Error(t, err)
				assert.Equal(t, test.errStr, err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.readableKey, attr.KeyLabel)
			assert.Equal(t, test.attrType, attr.Value.Type)
			attrKey, err := AttrKeyFromLabel(test.readableKey)
			assert.NoError(t, err)
			assert.Equal(t, attrKey, attr.Key)
			str, err := attr.Value.String()
			assert.NoError(t, err)
			assert.Equal(t, test.value, str)
		})
	}
}

func TestAttrKey(t *testing.T) {
	a, err := AttrKeyFromLabel("")
	assert.Error(t, err)
	a, err = AttrKeyFromLabel("somekey")
	assert.NoError(t, err)
	m := map[AttrKey]string{a: "dwefw"}
	mstr, err := json.Marshal(m)
	assert.NoError(t, err)
	m1 := make(map[AttrKey]string)
	err = json.Unmarshal(mstr, &m1)
	assert.NoError(t, err)
	assert.Equal(t, m[a], m1[a])
}

func TestAttrValFromString(t *testing.T) {
	tests := []struct {
		name  string
		tp    AttributeType
		value string
		error bool
	}{
		{
			"Int256",
			AttrInt256,
			"12343",
			false,
		},
		{
			"Decimal",
			AttrDecimal,
			"12343.2121",
			false,
		},
		{
			"string",
			AttrString,
			"123ewqewqer",
			false,
		},
		{
			"byte",
			AttrBytes,
			"0x12321abc",
			false,
		},
		{
			"timestamp",
			AttrTimestamp,
			time.Now().UTC().Format(time.RFC3339),
			false,
		},
		{
			"timestamp_nano",
			AttrTimestamp,
			time.Now().UTC().Format(time.RFC3339Nano),
			false,
		},
		{
			"unknown type",
			AttributeType("some type"),
			"",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := AttrValFromString(test.tp, test.value)
			if test.error {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			str, err := v.String()
			assert.NoError(t, err)
			assert.Equal(t, test.value, str)

			v.Type = AttributeType("some type")
			_, err = v.String()
			assert.Error(t, err)
		})
	}
}

type mockAccount struct {
	config.Account
	mock.Mock
}

func (m *mockAccount) SignMsg(msg []byte) (*coredocumentpb.Signature, error) {
	args := m.Called(msg)
	sig, _ := args.Get(0).(*coredocumentpb.Signature)
	return sig, args.Error(1)
}

func TestNewSignedAttribute(t *testing.T) {
	// empty label
	_, err := NewSignedAttribute("", testingidentity.GenerateRandomDID(), nil, nil, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrEmptyAttrLabel, err))

	// failed sign
	label := "signed_label"
	did := testingidentity.GenerateRandomDID()
	id := utils.RandomSlice(32)
	version := utils.RandomSlice(32)
	value := utils.RandomSlice(50)

	epayload := attributeSignaturePayload(did[:], id, version, value)
	acc := new(mockAccount)
	acc.On("SignMsg", epayload).Return(nil, errors.New("failed")).Once()
	model := new(mockModel)
	model.On("ID").Return(id).Once()
	model.On("NextVersion").Return(version).Once()
	_, err = NewSignedAttribute(label, did, acc, model, value)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
	acc.AssertExpectations(t)
	model.AssertExpectations(t)

	// success
	signature := utils.RandomSlice(32)
	acc = new(mockAccount)
	acc.On("SignMsg", epayload).Return(&coredocumentpb.Signature{Signature: signature}, nil).Once()
	model = new(mockModel)
	model.On("ID").Return(id).Once()
	model.On("NextVersion").Return(version).Twice()
	attr, err := NewSignedAttribute(label, did, acc, model, value)
	assert.NoError(t, err)
	attrKey, err := AttrKeyFromLabel(label)
	assert.NoError(t, err)
	assert.Equal(t, attrKey, attr.Key)
	assert.Equal(t, label, attr.KeyLabel)
	assert.Equal(t, AttrSigned, attr.Value.Type)
	assert.Equal(t, signature, attr.Value.Signed.Signature)
	acc.AssertExpectations(t)
	model.AssertExpectations(t)
}

func TestNewMonetaryAttribute(t *testing.T) {
	// empty label
	_, err := NewMonetaryAttribute("", "", nil, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrEmptyAttrLabel, err))

	// decimal error, empty string
	_, err = NewMonetaryAttribute("invoice_amount", "", nil, nil)
	assert.Error(t, err)

	// success
	label := "invoice_amount"
	chainID := []byte{1}
	id := []byte("USD")
	attr, err := NewMonetaryAttribute(label, "1001.1001", chainID, id)
	assert.NoError(t, err)
	assert.Equal(t, AttrMonetary, attr.Value.Type)
	attrKey, err := AttrKeyFromLabel(label)
	assert.NoError(t, err)
	assert.Equal(t, attrKey, attr.Key)
	assert.Equal(t, id, attr.Value.Monetary.ID)
	assert.Equal(t, chainID, attr.Value.Monetary.ChainID)
}
