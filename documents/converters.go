package documents

import (
	"strings"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/common"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// BinaryAttachment represent a single file attached to invoice.
type BinaryAttachment struct {
	Name     string
	FileType string // mime type of attached file
	Size     uint64 // in bytes
	Data     []byte
	Checksum []byte // the md5 checksum of the original file for easier verification
}

// PaymentDetails holds the payment related details for invoice.
type PaymentDetails struct {
	ID                    string // identifying this payment. could be a sequential number, could be a transaction hash of the crypto payment
	DateExecuted          *timestamp.Timestamp
	Payee                 *identity.DID // centrifuge id of payee
	Payer                 *identity.DID // centrifuge id of payer
	Amount                *Decimal
	Currency              string
	Reference             string // payment reference (e.g. reference field on bank transfer)
	BankName              string
	BankAddress           string
	BankCountry           string
	BankAccountNumber     string
	BankAccountCurrency   string
	BankAccountHolderName string
	BankKey               string

	CryptoChainURI      string // the ID of the chain to use in URI format. e.g. "ethereum://42/<tokenaddress>"
	CryptoTransactionID string // the transaction in which the payment happened
	CryptoFrom          string // from address
	CryptoTo            string // to address
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
			Size:     att.Size,
			Data:     data,
			Checksum: checksum,
		})
	}

	return catts
}

// ToP2PAttachments converts Binary Attchments to P2P attachments.
func ToP2PAttachments(atts []*BinaryAttachment) []*commonpb.BinaryAttachment {
	var patts []*commonpb.BinaryAttachment
	for _, att := range atts {
		patts = append(patts, &commonpb.BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     att.Size,
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
			Size:     att.Size,
			Data:     data,
			Checksum: checksum,
		})
	}

	return atts, nil
}

// FromP2PAttachments converts P2P attachments to Binary Attchments
func FromP2PAttachments(patts []*commonpb.BinaryAttachment) []*BinaryAttachment {
	var atts []*BinaryAttachment
	for _, att := range patts {
		atts = append(atts, &BinaryAttachment{
			Name:     att.Name,
			FileType: att.FileType,
			Size:     att.Size,
			Data:     att.Data,
			Checksum: att.Checksum,
		})
	}

	return atts
}

// ToClientPaymentDetails converts PaymentDetails to client payment details.
func ToClientPaymentDetails(details []*PaymentDetails) []*documentpb.PaymentDetails {
	var cdetails []*documentpb.PaymentDetails
	for _, detail := range details {
		decs := DecimalsToStrings(detail.Amount)
		dids := identity.DIDsToStrings(detail.Payee, detail.Payer)
		cdetails = append(cdetails, &documentpb.PaymentDetails{
			Id:                    detail.ID,
			DateExecuted:          detail.DateExecuted,
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

	return cdetails
}

// ToP2PPaymentDetails converts payment details to P2P payment details
func ToP2PPaymentDetails(details []*PaymentDetails) ([]*commonpb.PaymentDetails, error) {
	var pdetails []*commonpb.PaymentDetails
	for _, detail := range details {
		decs, err := DecimalsToBytes(detail.Amount)
		if err != nil {
			return nil, err
		}
		dids := identity.DIDsToBytes(detail.Payee, detail.Payer)
		pdetails = append(pdetails, &commonpb.PaymentDetails{
			Id:                    detail.ID,
			DateExecuted:          detail.DateExecuted,
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

		details = append(details, &PaymentDetails{
			ID:                    detail.Id,
			DateExecuted:          detail.DateExecuted,
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

// FromP2PPaymentDetails converts P2P payment details to PaymentDetails
func FromP2PPaymentDetails(pdetails []*commonpb.PaymentDetails) ([]*PaymentDetails, error) {
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
		details = append(details, &PaymentDetails{
			ID:                    detail.Id,
			DateExecuted:          detail.DateExecuted,
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
func FromClientCollaboratorAccess(racess *documentpb.ReadAccess, waccess *documentpb.WriteAccess) (ca CollaboratorsAccess, err error) {
	wmap, rmap := make(map[string]struct{}), make(map[string]struct{})
	var wcs, rcs []string
	if waccess != nil {
		for _, c := range waccess.Collaborators {
			c = strings.ToLower(c)
			if _, ok := wmap[c]; ok {
				continue
			}

			wmap[c] = struct{}{}
			wcs = append(wcs, c)
		}
	}

	if racess != nil {
		for _, c := range racess.Collaborators {
			c = strings.ToLower(c)
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
func ToClientCollaboratorAccess(ca CollaboratorsAccess) (*documentpb.ReadAccess, *documentpb.WriteAccess) {
	rcs := identity.DIDsToStrings(identity.DIDsPointers(ca.ReadCollaborators...)...)
	wcs := identity.DIDsToStrings(identity.DIDsPointers(ca.ReadWriteCollaborators...)...)
	return &documentpb.ReadAccess{Collaborators: rcs}, &documentpb.WriteAccess{Collaborators: wcs}
}

// ToClientAttributes converts attribute map to the client api format
func ToClientAttributes(attributes map[AttrKey]Attribute) (map[string]*documentpb.Attribute, error) {
	if len(attributes) < 1 {
		return nil, nil
	}

	m := make(map[string]*documentpb.Attribute)
	for k, v := range attributes {
		val, err := v.Value.String()
		if err != nil {
			return nil, errors.NewTypedError(ErrCDAttribute, err)
		}

		m[v.KeyLabel] = &documentpb.Attribute{
			Key:   k.String(),
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
		attr, err := newAttribute(k, attributeType(at.Type), at.Value)
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
		Version:     hexutil.Encode(model.CurrentVersion()),
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
