package model

import (
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
	"reflect"
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
	NetAmount            int64
	TaxAmount            int64
	TaxRate              int64
	Recipient            []byte
	Sender               []byte
	Payee                []byte
	Comment              string
	DueDate              *timestamp.Timestamp
	DateCreated          *timestamp.Timestamp
	ExtraData            []byte
}


func (i *Invoice) createInvoiceData() *invoicepb.InvoiceData {
	invoiceData := &invoicepb.InvoiceData{
		InvoiceNumber:i.InvoiceNumber,
		SenderName:i.SenderName,
		SenderStreet:i.SenderStreet,
		SenderCity:i.SenderCity,
		SenderZipcode:i.SenderZipcode,
		SenderCountry:i.SenderCountry,
		RecipientName:i.RecipientName,
		RecipientStreet:i.RecipientStreet,
		RecipientCity:i.RecipientCity,
		RecipientZipcode:i.RecipientZipcode,
		RecipientCountry:i.RecipientCountry,
		Currency:i.Currency,
		GrossAmount:i.GrossAmount,
		NetAmount:i.NetAmount,
		TaxAmount:i.TaxAmount,
		TaxRate:i.TaxRate,
		Recipient:i.Recipient,
		Sender:i.Sender,
		Payee:i.Payee,
		Comment:i.Comment,
		DueDate:i.DueDate,
		DateCreated:i.DateCreated,
		ExtraData:i.ExtraData,
	}
	return invoiceData
}

func generateInvoiceSalts(invoiceData *invoicepb.InvoiceData) *invoicepb.InvoiceDataSalts{
	invoiceSalts := &invoicepb.InvoiceDataSalts{}
	proofs.FillSalts(invoiceData, invoiceSalts)
	return invoiceSalts
}


func (i *Invoice) CoreDocument() (*coredocumentpb.CoreDocument, error) {
	coreDocument := new(coredocumentpb.CoreDocument)
	//proto.Merge(coreDocument, inv.Document.CoreDocument)
	invoiceData := i.createInvoiceData()
	serializedInvoice, err := proto.Marshal(invoiceData)
	if err != nil {
		return nil, centerrors.Wrap(err, "couldn't serialise InvoiceData")
	}

	invoiceAny := any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   serializedInvoice,
	}

	invoiceSalt := generateInvoiceSalts(invoiceData)

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

func (i *Invoice) SetCoreDocument(cd *coredocumentpb.CoreDocument) error {
	panic("implement me")
}

func (i *Invoice) JSON() ([]byte, error) {
	panic("implement me")
}

func (i *Invoice) Type() reflect.Type {
	return reflect.TypeOf(i)
}
