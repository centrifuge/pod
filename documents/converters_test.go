//go:build unit

package documents

import (
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
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
	payee, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	payer, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	dec := new(Decimal)
	err = dec.SetString("0.99")
	assert.NoError(t, err)

	details := []*PaymentDetails{
		{
			ID:     "some id",
			Payee:  payee,
			Payer:  payer,
			Amount: dec,
		},
	}

	pdetails, err := ToProtocolPaymentDetails(details)
	assert.NoError(t, err)

	fpdetails, err := FromProtocolPaymentDetails(pdetails)
	assert.NoError(t, err)

	assert.Equal(t, details[0].Amount.String(), fpdetails[0].Amount.String())
	assert.Equal(t, details[0].ID, fpdetails[0].ID)
	assert.Equal(t, details[0].Payee, fpdetails[0].Payee)

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
		attr, err := NewStringAttribute(k, AttributeType(at.Type), at.Value)
		assert.NoError(t, err)

		m[attr.Key] = attr
	}

	return m
}

func TestP2PAttributes(t *testing.T) {
	cattrs := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339Nano),
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
			Value: time.Now().UTC().Format(time.RFC3339Nano),
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
	}

	attrs := toAttrsMap(t, cattrs)
	label := "signed_label"
	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	id := utils.RandomSlice(32)
	version := utils.RandomSlice(32)
	value := utils.RandomSlice(50)

	var epayload []byte
	epayload = append(epayload, identity.ToBytes()...)
	epayload = append(epayload, id...)
	epayload = append(epayload, version...)
	epayload = append(epayload, value...)

	signature := utils.RandomSlice(32)

	acc := config.NewAccountMock(t)
	acc.On("SignMsg", epayload).
		Once().
		Return(&coredocumentpb.Signature{Signature: signature}, nil)

	attr, err := NewSignedAttribute(label, identity, acc, id, version, value, AttrBytes)
	assert.NoError(t, err)

	attrs[attr.Key] = attr

	pattrs, err := toProtocolAttributes(attrs)
	assert.NoError(t, err)
	assert.Equal(t, "int256_test", string(pattrs[3].KeyLabel))
	assert.Len(t, pattrs[3].GetByteVal(), maxDecimalByteLength) //decimal length padded to 32 bytes
	assert.Equal(t, "time_test", string(pattrs[0].KeyLabel))
	assert.Len(t, pattrs[0].GetByteVal(), maxTimeByteLength) //timestamp length padded to 12 bytes

	gattrs, err := fromProtocolAttributes(pattrs)
	assert.NoError(t, err)
	assert.Equal(t, attrs, gattrs)

	// wrong id
	signed := pattrs[len(pattrs)-1].GetSignedVal()
	signed.Identity = nil
	_, err = fromProtocolAttributes(pattrs)
	assert.Error(t, err)
}

func TestAttributes_monetary(t *testing.T) {
	dec, err := NewDecimal("1001.1001")
	assert.NoError(t, err)

	tests_monetary := []struct {
		label   string
		dec     *Decimal
		chainID []byte
		id      string
	}{
		{
			label:   "invoice_amount",
			dec:     dec,
			chainID: []byte{1},
			id:      "USD",
		},
		{
			label:   "invoice_amount_erc20",
			dec:     dec,
			chainID: []byte{1},
			id:      "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2",
		},
	}

	for _, v := range tests_monetary {
		mon, err := NewMonetaryAttribute(v.label, v.dec, v.chainID, v.id)
		assert.NoError(t, err)
		m := map[AttrKey]Attribute{
			mon.Key: mon,
		}
		// convert to protocol
		pattrs, err := toProtocolAttributes(m)
		assert.NoError(t, err)
		assert.Len(t, pattrs, 1)
		assert.Equal(t, []byte(v.label), pattrs[0].KeyLabel)
		assert.Equal(t, coredocumentpb.AttributeType_ATTRIBUTE_TYPE_MONETARY, pattrs[0].Type)
		assert.Len(t, pattrs[0].GetMonetaryVal().Value, 32)
		assert.Len(t, pattrs[0].GetMonetaryVal().Id, 32)
		assert.Len(t, pattrs[0].GetMonetaryVal().Chain, 4)

		// convert from protocol
		cAttr, err := fromProtocolAttributes(pattrs)
		assert.NoError(t, err)
		assert.Equal(t, mon.Value.Monetary.Value.String(), cAttr[mon.Key].Value.Monetary.Value.String())
		assert.Equal(t, mon.Value.Monetary.ChainID, cAttr[mon.Key].Value.Monetary.ChainID)
		assert.Equal(t, mon.Value.Monetary.ID, cAttr[mon.Key].Value.Monetary.ID)
	}

}
