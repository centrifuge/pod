package documents

import (
	"strings"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/common"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-centrifuge/utils/timeutils"
)

// BinaryAttachment represent a single file attached to invoice.
type BinaryAttachment struct {
	Name     string             `json:"name"`
	FileType string             `json:"file_type"` // mime type of attached file
	Size     int                `json:"size"`      // in bytes
	Data     byteutils.HexBytes `json:"data" swaggertype:"primitive,string"`
	Checksum byteutils.HexBytes `json:"checksum" swaggertype:"primitive,string"` // the md5 checksum of the original file for easier verification
}

// PaymentDetails holds the payment related details for invoice.
type PaymentDetails struct {
	ID                    string        `json:"id"` // identifying this payment. could be a sequential number, could be a transaction hash of the crypto payment
	DateExecuted          *time.Time    `json:"date_executed" swaggertype:"primitive,string"`
	Payee                 *identity.DID `json:"payee" swaggertype:"primitive,string"` // centrifuge id of payee
	Payer                 *identity.DID `json:"payer" swaggertype:"primitive,string"` // centrifuge id of payer
	Amount                *Decimal      `json:"amount" swaggertype:"primitive,string"`
	Currency              string        `json:"currency"`
	Reference             string        `json:"reference"` // payment reference (e.g. reference field on bank transfer)
	BankName              string        `json:"bank_name"`
	BankAddress           string        `json:"bank_address"`
	BankCountry           string        `json:"bank_country"`
	BankAccountNumber     string        `json:"bank_account_number"`
	BankAccountCurrency   string        `json:"bank_account_currency"`
	BankAccountHolderName string        `json:"bank_account_holder_name"`
	BankKey               string        `json:"bank_key"`

	CryptoChainURI      string `json:"crypto_chain_uri"`      // the ID of the chain to use in URI format. e.g. "ethereum://42/<tokenaddress>"
	CryptoTransactionID string `json:"crypto_transaction_id"` // the transaction in which the payment happened
	CryptoFrom          string `json:"crypto_from"`           // from address
	CryptoTo            string `json:"crypto_to"`             // to address
}

// ToProtocolAttachments converts Binary Attchments to protocol attachments.
func ToProtocolAttachments(atts []*BinaryAttachment) []*commonpb.BinaryAttachment {
	var patts []*commonpb.BinaryAttachment
	for _, att := range atts {
		patts = append(patts, &commonpb.BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     uint64(att.Size),
			Data:     att.Data,
			Checksum: att.Checksum,
		})
	}

	return patts
}

// FromProtocolAttachments converts Protocol attachments to Binary Attachments
func FromProtocolAttachments(patts []*commonpb.BinaryAttachment) []*BinaryAttachment {
	var atts []*BinaryAttachment
	for _, att := range patts {
		atts = append(atts, &BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     int(att.Size),
			Data:     att.Data,
			Checksum: att.Checksum,
		})
	}

	return atts
}

// ToProtocolPaymentDetails converts payment details to protocol payment details
func ToProtocolPaymentDetails(details []*PaymentDetails) ([]*commonpb.PaymentDetails, error) {
	var pdetails []*commonpb.PaymentDetails
	for _, detail := range details {
		decs, err := DecimalsToBytes(detail.Amount)
		if err != nil {
			return nil, err
		}

		tms, err := timeutils.ToProtoTimestamps(detail.DateExecuted)
		if err != nil {
			return nil, err
		}

		dids := identity.DIDsToBytes(detail.Payee, detail.Payer)
		pdetails = append(pdetails, &commonpb.PaymentDetails{
			Id:                    detail.ID,
			DateExecuted:          tms[0],
			Payee:                 dids[0],
			Payer:                 dids[1],
			Amount:                decs[0],
			Currency:              detail.Currency,
			Reference:             detail.Reference,
			BankName:              detail.BankName,
			BankAddress:           detail.BankAddress,
			BankAccountCurrency:   detail.BankAccountCurrency,
			BankAccountHolderName: detail.BankAccountHolderName,
			BankAccountNumber:     detail.BankAccountNumber,
			BankCountry:           detail.BankCountry,
			BankKey:               detail.BankKey,
			CryptoChainUri:        detail.CryptoChainURI,
			CryptoFrom:            detail.CryptoFrom,
			CryptoTo:              detail.CryptoTo,
			CryptoTransactionId:   detail.CryptoTransactionID,
		})
	}

	return pdetails, nil
}

// FromProtocolPaymentDetails converts protocol payment details to PaymentDetails
func FromProtocolPaymentDetails(pdetails []*commonpb.PaymentDetails) ([]*PaymentDetails, error) {
	var details []*PaymentDetails
	for _, detail := range pdetails {
		decs, err := BytesToDecimals(detail.Amount)
		if err != nil {
			return nil, err
		}
		dids, err := identity.BytesToDIDs(detail.Payee, detail.Payer)
		if err != nil {
			return nil, err
		}

		pts, err := timeutils.FromProtoTimestamps(detail.DateExecuted)
		if err != nil {
			return nil, err
		}

		details = append(details, &PaymentDetails{
			ID:                    detail.Id,
			DateExecuted:          pts[0],
			Payee:                 dids[0],
			Payer:                 dids[1],
			Amount:                decs[0],
			Currency:              detail.Currency,
			Reference:             detail.Reference,
			BankName:              detail.BankName,
			BankAddress:           detail.BankAddress,
			BankAccountCurrency:   detail.BankAccountCurrency,
			BankAccountHolderName: detail.BankAccountHolderName,
			BankAccountNumber:     detail.BankAccountNumber,
			BankCountry:           detail.BankCountry,
			BankKey:               detail.BankKey,
			CryptoChainURI:        detail.CryptoChainUri,
			CryptoFrom:            detail.CryptoFrom,
			CryptoTo:              detail.CryptoTo,
			CryptoTransactionID:   detail.CryptoTransactionId,
		})
	}

	return details, nil
}

// toProtocolAttributes convert model attributes to p2p attributes
// since the protocol representation of attributes is a list, we will always sort the keys and then insert to the list.
func toProtocolAttributes(attrs map[AttrKey]Attribute) (pattrs []*coredocumentpb.Attribute, err error) {
	var keys [][32]byte
	for k := range attrs {
		k := k
		keys = append(keys, k)
	}

	keys = byteutils.SortByte32Slice(keys)
	for _, k := range keys {
		attr := attrs[k]
		if !isAttrTypeAllowed(attr.Value.Type) {
			return nil, ErrNotValidAttrType
		}

		pattr := &coredocumentpb.Attribute{
			Key:      attr.Key[:],
			KeyLabel: []byte(attr.KeyLabel),
			Type:     getProtocolAttributeType(attr.Value.Type),
		}

		switch attr.Value.Type {
		case AttrInt256:
			b := attr.Value.Int256.Bytes()
			pattr.Value = &coredocumentpb.Attribute_ByteVal{ByteVal: b[:]}
		case AttrDecimal:
			b, err := attr.Value.Decimal.Bytes()
			if err != nil {
				return nil, err
			}
			pattr.Value = &coredocumentpb.Attribute_ByteVal{ByteVal: b}
		case AttrString:
			pattr.Value = &coredocumentpb.Attribute_StrVal{StrVal: attr.Value.Str}
		case AttrBytes:
			pattr.Value = &coredocumentpb.Attribute_ByteVal{ByteVal: attr.Value.Bytes}
		case AttrTimestamp:
			pattr.Value = &coredocumentpb.Attribute_TimeVal{TimeVal: attr.Value.Timestamp}
		case AttrSigned:
			signed := attr.Value.Signed
			pattr.Value = &coredocumentpb.Attribute_SignedVal{
				SignedVal: &coredocumentpb.Signed{
					DocVersion: signed.DocumentVersion,
					Value:      signed.Value,
					Signature:  signed.Signature,
					PublicKey:  signed.PublicKey,
					Identity:   signed.Identity[:],
				},
			}
		}

		pattrs = append(pattrs, pattr)
	}

	return pattrs, nil
}

const attributeProtocolPrefix = "ATTRIBUTE_TYPE_"

func getProtocolAttributeType(attrType AttributeType) coredocumentpb.AttributeType {
	str := attributeProtocolPrefix + strings.ToUpper(attrType.String())
	return coredocumentpb.AttributeType(coredocumentpb.AttributeType_value[str])
}

func getAttributeTypeFromProtocolType(attrType coredocumentpb.AttributeType) AttributeType {
	str := coredocumentpb.AttributeType_name[int32(attrType)]
	return AttributeType(strings.ToLower(strings.TrimPrefix(str, attributeProtocolPrefix)))
}

// fromProtocolAttributes converts protocol attribute list to model attribute map
func fromProtocolAttributes(pattrs []*coredocumentpb.Attribute) (map[AttrKey]Attribute, error) {
	m := make(map[AttrKey]Attribute)
	for _, pattr := range pattrs {
		attrKey, err := AttrKeyFromBytes(pattr.Key)
		if err != nil {
			return nil, err
		}

		attrType := getAttributeTypeFromProtocolType(pattr.Type)
		attr := Attribute{
			Key:      attrKey,
			KeyLabel: string(pattr.KeyLabel),
		}

		attr.Value, err = attrValFromProtocolAttribute(attrType, pattr)
		if err != nil {
			return nil, err
		}

		m[attrKey] = attr
	}

	return m, nil
}

func attrValFromProtocolAttribute(attrType AttributeType, attribute *coredocumentpb.Attribute) (attrVal AttrVal, err error) {
	if !isAttrTypeAllowed(attrType) {
		return attrVal, ErrNotValidAttrType
	}

	attrVal.Type = attrType
	switch attrType {
	case AttrInt256:
		attrVal.Int256, err = Int256FromBytes(attribute.GetByteVal())
	case AttrDecimal:
		attrVal.Decimal, err = DecimalFromBytes(attribute.GetByteVal())
	case AttrString:
		attrVal.Str = attribute.GetStrVal()
	case AttrBytes:
		attrVal.Bytes = attribute.GetByteVal()
	case AttrTimestamp:
		attrVal.Timestamp = attribute.GetTimeVal()
	case AttrSigned:
		val := attribute.GetSignedVal()
		did, err := identity.NewDIDFromBytes(val.Identity)
		if err != nil {
			return attrVal, err
		}

		attrVal.Signed = Signed{
			Identity:        did,
			DocumentVersion: val.DocVersion,
			PublicKey:       val.PublicKey,
			Value:           val.Value,
			Signature:       val.Signature,
		}
	}

	return attrVal, err
}
