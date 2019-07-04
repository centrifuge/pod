package documents

import (
	"strings"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/common"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-centrifuge/utils/timeutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

// ToClientAttachments converts Attachments to Client Attachments.
func ToClientAttachments(atts []*BinaryAttachment) []*documentpb.BinaryAttachment {
	var catts []*documentpb.BinaryAttachment
	for _, att := range atts {
		var data, checksum string
		if len(att.Data) > 0 {
			data = hexutil.Encode(att.Data)
		}

		if len(att.Checksum) > 0 {
			checksum = hexutil.Encode(att.Checksum)
		}

		catts = append(catts, &documentpb.BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     uint64(att.Size),
			Data:     data,
			Checksum: checksum,
		})
	}

	return catts
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

// FromClientAttachments converts Client Attachments to Binary Attachments
func FromClientAttachments(catts []*documentpb.BinaryAttachment) ([]*BinaryAttachment, error) {
	var atts []*BinaryAttachment
	for _, att := range catts {
		var data, checksum []byte
		var err error
		if s := strings.TrimSpace(att.Data); s != "" {
			data, err = hexutil.Decode(s)
			if err != nil {
				return nil, err
			}
		}

		if s := strings.TrimSpace(att.Checksum); s != "" {
			checksum, err = hexutil.Decode(s)
			if err != nil {
				return nil, err
			}
		}

		atts = append(atts, &BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     int(att.Size),
			Data:     data,
			Checksum: checksum,
		})
	}

	return atts, nil
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

// ToClientPaymentDetails converts PaymentDetails to client payment details.
func ToClientPaymentDetails(details []*PaymentDetails) ([]*documentpb.PaymentDetails, error) {
	var cdetails []*documentpb.PaymentDetails
	for _, detail := range details {
		decs := DecimalsToStrings(detail.Amount)
		dids := identity.DIDsToStrings(detail.Payee, detail.Payer)
		tms, err := timeutils.ToProtoTimestamps(detail.DateExecuted)
		if err != nil {
			return nil, err
		}

		cdetails = append(cdetails, &documentpb.PaymentDetails{
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

	return cdetails, nil
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

// FromClientPaymentDetails converts Client PaymentDetails to PaymentDetails
func FromClientPaymentDetails(cdetails []*documentpb.PaymentDetails) ([]*PaymentDetails, error) {
	var details []*PaymentDetails
	for _, detail := range cdetails {
		decs, err := StringsToDecimals(detail.Amount)
		if err != nil {
			return nil, err
		}

		dids, err := identity.StringsToDIDs(detail.Payee, detail.Payer)
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

// FromClientCollaboratorAccess converts client collaborator access to CollaboratorsAccess
func FromClientCollaboratorAccess(racess, waccess []string) (ca CollaboratorsAccess, err error) {
	wmap, rmap := make(map[string]struct{}), make(map[string]struct{})
	var wcs, rcs []string
	if waccess != nil {
		for _, c := range waccess {
			c = strings.TrimSpace(strings.ToLower(c))
			if c == "" {
				continue
			}

			if _, ok := wmap[c]; ok {
				continue
			}

			wmap[c] = struct{}{}
			wcs = append(wcs, c)
		}
	}

	if racess != nil {
		for _, c := range racess {
			c = strings.TrimSpace(strings.ToLower(c))
			if c == "" {
				continue
			}

			if _, ok := wmap[c]; ok {
				continue
			}

			if _, ok := rmap[c]; ok {
				continue
			}

			rcs = append(rcs, c)
			rmap[c] = struct{}{}
		}
	}

	rdids, err := identity.StringsToDIDs(rcs...)
	if err != nil {
		return ca, err
	}

	wdids, err := identity.StringsToDIDs(wcs...)
	if err != nil {
		return ca, err
	}

	return CollaboratorsAccess{
		ReadCollaborators:      identity.FromPointerDIDs(rdids...),
		ReadWriteCollaborators: identity.FromPointerDIDs(wdids...),
	}, nil
}

// ToClientCollaboratorAccess converts CollaboratorAccess to client collaborator access
func ToClientCollaboratorAccess(ca CollaboratorsAccess) (readAccess, writeAccess []string) {
	rcs := identity.DIDsToStrings(identity.DIDsPointers(ca.ReadCollaborators...)...)
	wcs := identity.DIDsToStrings(identity.DIDsPointers(ca.ReadWriteCollaborators...)...)
	return rcs, wcs
}

// ToClientAttributes converts attribute map to the client api format
func ToClientAttributes(attributes []Attribute) (map[string]*documentpb.Attribute, error) {
	if len(attributes) < 1 {
		return nil, nil
	}

	m := make(map[string]*documentpb.Attribute)
	for _, v := range attributes {
		val, err := v.Value.String()
		if err != nil {
			return nil, errors.NewTypedError(ErrCDAttribute, err)
		}

		m[v.KeyLabel] = &documentpb.Attribute{
			Key:   v.Key.String(),
			Type:  v.Value.Type.String(),
			Value: val,
		}
	}

	return m, nil
}

// FromClientAttributes converts the api attributes type to local Attributes map.
func FromClientAttributes(attrs map[string]*documentpb.Attribute) (map[AttrKey]Attribute, error) {
	if len(attrs) < 1 {
		return nil, nil
	}

	m := make(map[AttrKey]Attribute)
	for k, at := range attrs {
		attr, err := NewAttribute(k, AttributeType(at.Type), at.Value)
		if err != nil {
			return nil, errors.NewTypedError(ErrCDAttribute, err)
		}

		m[attr.Key] = attr
	}

	return m, nil
}

// DeriveResponseHeader derives common response header for model
func DeriveResponseHeader(tokenRegistry TokenRegistry, model Model) (*documentpb.ResponseHeader, error) {
	cs, err := model.GetCollaborators()
	if err != nil {
		return nil, errors.NewTypedError(ErrCollaborators, err)
	}

	// we ignore error here because it can happen when a model is first created but its not anchored yet
	a, _ := model.Author()
	author := a.String()

	// we ignore error here because it can happen when a model is first created but its not anchored yet
	time := ""
	t, err := model.Timestamp()
	if err == nil {
		time = t.UTC().String()
	}

	nfts := model.NFTs()
	cnfts, err := convertNFTs(tokenRegistry, nfts)
	if err != nil {
		// this could be a temporary failure, so we ignore but warn about the error
		log.Warningf("errors encountered when trying to set nfts to the response: %v", errors.NewTypedError(ErrNftNotFound, err))
	}

	rcs, wcs := ToClientCollaboratorAccess(cs)
	return &documentpb.ResponseHeader{
		DocumentId:  hexutil.Encode(model.ID()),
		VersionId:   hexutil.Encode(model.CurrentVersion()),
		Author:      author,
		CreatedAt:   time,
		ReadAccess:  rcs,
		WriteAccess: wcs,
		Nfts:        cnfts,
	}, nil
}

func convertNFTs(tokenRegistry TokenRegistry, nfts []*coredocumentpb.NFT) (nnfts []*documentpb.NFT, err error) {
	for _, n := range nfts {
		regAddress := common.BytesToAddress(n.RegistryId[:common.AddressLength])
		i, errn := tokenRegistry.CurrentIndexOfToken(regAddress, n.TokenId)
		if errn != nil || i == nil {
			err = errors.AppendError(err, errors.New("token index received is nil or other error: %v", errn))
			continue
		}

		o, errn := tokenRegistry.OwnerOf(regAddress, n.TokenId)
		if errn != nil {
			err = errors.AppendError(err, errn)
			continue
		}

		nnfts = append(nnfts, &documentpb.NFT{
			Registry:   regAddress.Hex(),
			Owner:      o.Hex(),
			TokenId:    hexutil.Encode(n.TokenId),
			TokenIndex: hexutil.Encode(i.Bytes()),
		})
	}
	return nnfts, err
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
