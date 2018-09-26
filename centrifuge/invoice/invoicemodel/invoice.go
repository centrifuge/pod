package invoicemodel

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/centrifuge/go-centrifuge/centrifuge/identity"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// example of an implementation
type Invoice struct {
	// invoice number or reference number
	InvoiceNumber string
	// name of the sender company
	SenderName string
	// street and address details of the sender company
	SenderStreet  string
	SenderCity    string
	SenderZipcode string
	// country ISO code of the sender of this invoice
	SenderCountry string
	// name of the recipient company
	RecipientName    string
	RecipientStreet  string
	RecipientCity    string
	RecipientZipcode string
	// country ISO code of the receipient of this invoice
	RecipientCountry string
	// ISO currency code
	Currency string
	// invoice amount including tax
	GrossAmount int64
	// invoice amount excluding tax
	NetAmount   int64
	TaxAmount   int64
	TaxRate     int64
	Recipient   identity.CentID
	Sender      identity.CentID
	Payee       identity.CentID
	Comment     string
	DueDate     *timestamp.Timestamp
	DateCreated *timestamp.Timestamp
	ExtraData   []byte

	InvoiceSalts *invoicepb.InvoiceDataSalts
}

func (i *Invoice) createInvoiceData() (*invoicepb.InvoiceData, error) {

	recipient, err := i.Recipient.MarshalBinary()
	if err != nil {
		return nil, err
	}

	sender, err := i.Sender.MarshalBinary()
	if err != nil {
		return nil, err
	}
	payee, err := i.Payee.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &invoicepb.InvoiceData{
		InvoiceNumber:    i.InvoiceNumber,
		SenderName:       i.SenderName,
		SenderStreet:     i.SenderStreet,
		SenderCity:       i.SenderCity,
		SenderZipcode:    i.SenderZipcode,
		SenderCountry:    i.SenderCountry,
		RecipientName:    i.RecipientName,
		RecipientStreet:  i.RecipientStreet,
		RecipientCity:    i.RecipientCity,
		RecipientZipcode: i.RecipientZipcode,
		RecipientCountry: i.RecipientCountry,
		Currency:         i.Currency,
		GrossAmount:      i.GrossAmount,
		NetAmount:        i.NetAmount,
		TaxAmount:        i.TaxAmount,
		TaxRate:          i.TaxRate,
		Recipient:        recipient,
		Sender:           sender,
		Payee:            payee,
		Comment:          i.Comment,
		DueDate:          i.DueDate,
		DateCreated:      i.DateCreated,
		ExtraData:        i.ExtraData,
	}, nil

}

func (i *Invoice) initInvoice(invoiceData *invoicepb.InvoiceData) error {

	i.InvoiceNumber = invoiceData.InvoiceNumber
	i.SenderName = invoiceData.SenderName
	i.SenderStreet = invoiceData.SenderStreet
	i.SenderCity = invoiceData.SenderCity
	i.SenderZipcode = invoiceData.SenderZipcode
	i.SenderCountry = invoiceData.SenderCountry
	i.RecipientName = invoiceData.RecipientName
	i.RecipientStreet = invoiceData.RecipientStreet
	i.RecipientCity = invoiceData.RecipientCity
	i.RecipientZipcode = invoiceData.RecipientZipcode
	i.RecipientCountry = invoiceData.RecipientCountry
	i.Currency = invoiceData.Currency
	i.GrossAmount = invoiceData.GrossAmount
	i.NetAmount = invoiceData.NetAmount
	i.TaxAmount = invoiceData.TaxAmount
	i.TaxRate = invoiceData.TaxRate

	recipientCentID, err := identity.NewCentID(invoiceData.Recipient)
	if err != nil {
		return err
	}
	i.Recipient = recipientCentID

	senderCentID, err := identity.NewCentID(invoiceData.Sender)
	if err != nil {
		return err
	}
	i.Sender = senderCentID

	payeeCentID, err := identity.NewCentID(invoiceData.Payee)
	if err != nil {
		return err
	}
	i.Payee = payeeCentID

	i.Comment = invoiceData.Comment
	i.DueDate = invoiceData.DueDate
	i.DateCreated = invoiceData.DateCreated
	i.ExtraData = invoiceData.ExtraData

	return nil

}

func (i *Invoice) getInvoiceSalts(invoiceData *invoicepb.InvoiceData) *invoicepb.InvoiceDataSalts {
	if i.InvoiceSalts == nil {
		invoiceSalts := &invoicepb.InvoiceDataSalts{}
		proofs.FillSalts(invoiceData, invoiceSalts)
		i.InvoiceSalts = invoiceSalts

	}
	return i.InvoiceSalts
}

//CoreDocument returns a CoreDocument with an embedded invoice
func (i *Invoice) CoreDocument() (*coredocumentpb.CoreDocument, error) {
	coreDocument := new(coredocumentpb.CoreDocument)

	invoiceData, err := i.createInvoiceData()
	if err != nil {
		return nil, err
	}
	serializedInvoice, err := proto.Marshal(invoiceData)
	if err != nil {
		return nil, centerrors.Wrap(err, "couldn't serialise InvoiceData")
	}

	invoiceAny := any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   serializedInvoice,
	}

	invoiceSalt := i.getInvoiceSalts(invoiceData)

	serializedSalts, err := proto.Marshal(invoiceSalt)
	if err != nil {
		return nil, centerrors.Wrap(err, "couldn't serialise InvoiceSalts")
	}

	invoiceSaltsAny := any.Any{
		TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
		Value:   serializedSalts,
	}

	coreDocument.EmbeddedData = &invoiceAny
	coreDocument.EmbeddedDataSalts = &invoiceSaltsAny
	return coreDocument, err
}

//InitWithCoreDocument initials the invoice model with a core document which embeds an invoice
func (i *Invoice) InitWithCoreDocument(coreDocument *coredocumentpb.CoreDocument) error {
	if coreDocument == nil {
		return centerrors.NilError(coreDocument)
	}
	if coreDocument.EmbeddedData == nil || coreDocument.EmbeddedData.TypeUrl != documenttypes.InvoiceDataTypeUrl ||
		coreDocument.EmbeddedDataSalts.TypeUrl != documenttypes.InvoiceSaltsTypeUrl {
		return fmt.Errorf("trying to convert document with incorrect schema")
	}

	invoiceData := &invoicepb.InvoiceData{}
	err := proto.Unmarshal(coreDocument.EmbeddedData.Value, invoiceData)

	if err != nil {
		return err
	}

	invoiceSalts := &invoicepb.InvoiceDataSalts{}
	err = proto.Unmarshal(coreDocument.EmbeddedDataSalts.Value, invoiceSalts)

	if err != nil {
		return err
	}

	err = i.initInvoice(invoiceData)
	i.InvoiceSalts = invoiceSalts

	return err
}

func (i *Invoice) JSON() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Invoice) InitWithJSON(jsonData []byte) error {

	if err := json.Unmarshal(jsonData, i); err != nil {
		return err
	}
	return nil

}

func (i *Invoice) Type() reflect.Type {
	return reflect.TypeOf(i)
}
