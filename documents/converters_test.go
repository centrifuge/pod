// +build unit

package documents

import (
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestBinaryAttachments(t *testing.T) {
	atts := []*BinaryAttachment{
		{
			Name:     "some name",
			FileType: "pdf",
			Size:     1024,
			Data:     utils.RandomSlice(32),
			Checksum: utils.RandomSlice(32),
		},

		{
			Name:     "some name 1",
			FileType: "jpeg",
			Size:     4096,
			Data:     utils.RandomSlice(32),
			Checksum: utils.RandomSlice(32),
		},
	}

	patts := ToProtocolAttachments(atts)

	fpatts := FromProtocolAttachments(patts)
	assert.Equal(t, atts, fpatts)
}

func TestPaymentDetails(t *testing.T) {
	did := testingidentity.GenerateRandomDID()
	dec := new(Decimal)
	err := dec.SetString("0.99")
	assert.NoError(t, err)
	details := []*PaymentDetails{
		{
			ID:     "some id",
			Payee:  &did,
			Amount: dec,
		},
	}

	pdetails, err := ToProtocolPaymentDetails(details)
	assert.NoError(t, err)

	fpdetails, err := FromProtocolPaymentDetails(pdetails)
	assert.NoError(t, err)

	assert.Equal(t, details, fpdetails)

	pdetails[0].Amount = utils.RandomSlice(40)
	_, err = FromProtocolPaymentDetails(pdetails)
	assert.Error(t, err)
}

type attribute struct {
	Type, Value string
}

func toAttrsMap(t *testing.T, cattrs map[string]attribute) map[AttrKey]Attribute {
	m := make(map[AttrKey]Attribute)
	for k, at := range cattrs {
		attr, err := NewAttribute(k, AttributeType(at.Type), at.Value)
		assert.NoError(t, err)

		m[attr.Key] = attr
	}

	return m
}

func TestP2PAttributes(t *testing.T) {
	cattrs := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "some string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("some bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000001",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000001",
		},
	}

	attrs := toAttrsMap(t, cattrs)
	pattrs, err := toProtocolAttributes(attrs)
	assert.NoError(t, err)

	attrs1, err := fromProtocolAttributes(pattrs)
	assert.NoError(t, err)
	assert.Equal(t, attrs, attrs1)

	pattrs1, err := toProtocolAttributes(attrs1)
	assert.NoError(t, err)
	assert.Equal(t, pattrs, pattrs1)

	attrKey, err := AttrKeyFromBytes(pattrs1[0].Key)
	assert.NoError(t, err)
	val := attrs1[attrKey]
	val.Value.Type = AttributeType("some type")
	attrs1[attrKey] = val
	_, err = toProtocolAttributes(attrs1)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrNotValidAttrType, err))

	pattrs[0].Type = coredocumentpb.AttributeType(0)
	_, err = fromProtocolAttributes(pattrs)
	assert.Error(t, err)

	pattrs[0].Key = utils.RandomSlice(31)
	_, err = fromProtocolAttributes(pattrs)
	assert.Error(t, err)
}

func TestAttributes_signed(t *testing.T) {
	cattrs := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "some string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("some bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000001",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000001",
		},
	}

	attrs := toAttrsMap(t, cattrs)
	label := "signed_label"
	did := testingidentity.GenerateRandomDID()
	id := utils.RandomSlice(32)
	version := utils.RandomSlice(32)
	value := utils.RandomSlice(50)

	var epayload []byte
	epayload = append(epayload, did[:]...)
	epayload = append(epayload, id...)
	epayload = append(epayload, version...)
	epayload = append(epayload, value...)

	signature := utils.RandomSlice(32)
	acc := new(mockAccount)
	acc.On("SignMsg", epayload).Return(&coredocumentpb.Signature{Signature: signature}, nil).Once()
	model := new(mockModel)
	model.On("ID").Return(id).Once()
	model.On("NextVersion").Return(version).Twice()
	attr, err := NewSignedAttribute(label, did, acc, model, value)
	assert.NoError(t, err)
	acc.AssertExpectations(t)
	model.AssertExpectations(t)
	attrs[attr.Key] = attr

	pattrs, err := toProtocolAttributes(attrs)
	assert.NoError(t, err)
	assert.Equal(t, "decimal_test", string(pattrs[3].KeyLabel))
	assert.Len(t, pattrs[3].GetByteVal(), maxDecimalByteLength) //decimal length padded to 32 bytes
	assert.Equal(t, "time_test", string(pattrs[0].KeyLabel))
	assert.Len(t, pattrs[0].GetByteVal(), maxTimeByteLength) //decimal length padded to 8 bytes
	gattrs, err := fromProtocolAttributes(pattrs)
	assert.NoError(t, err)
	assert.Equal(t, attrs, gattrs)

	// wrong id
	signed := pattrs[len(pattrs)-1].GetSignedVal()
	signed.Identity = nil
	_, err = fromProtocolAttributes(pattrs)
	assert.Error(t, err)
}
