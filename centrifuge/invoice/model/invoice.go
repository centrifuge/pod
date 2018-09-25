package model

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/golang/protobuf/ptypes/timestamp"
	"reflect"
)

// example of an implementation
type Invoice struct {
	// invoice number or reference number
	InvoiceNumber string `protobuf:"bytes,1,opt,name=invoice_number,json=invoiceNumber" json:"invoice_number,omitempty"`
	// name of the sender company
	SenderName string `protobuf:"bytes,3,opt,name=sender_name,json=senderName" json:"sender_name,omitempty"`
	// street and address details of the sender company
	SenderStreet  string `protobuf:"bytes,4,opt,name=sender_street,json=senderStreet" json:"sender_street,omitempty"`
	SenderCity    string `protobuf:"bytes,5,opt,name=sender_city,json=senderCity" json:"sender_city,omitempty"`
	SenderZipcode string `protobuf:"bytes,6,opt,name=sender_zipcode,json=senderZipcode" json:"sender_zipcode,omitempty"`
	// country ISO code of the sender of this invoice
	SenderCountry string `protobuf:"bytes,7,opt,name=sender_country,json=senderCountry" json:"sender_country,omitempty"`
	// name of the recipient company
	RecipientName    string `protobuf:"bytes,8,opt,name=recipient_name,json=recipientName" json:"recipient_name,omitempty"`
	RecipientStreet  string `protobuf:"bytes,9,opt,name=recipient_street,json=recipientStreet" json:"recipient_street,omitempty"`
	RecipientCity    string `protobuf:"bytes,10,opt,name=recipient_city,json=recipientCity" json:"recipient_city,omitempty"`
	RecipientZipcode string `protobuf:"bytes,11,opt,name=recipient_zipcode,json=recipientZipcode" json:"recipient_zipcode,omitempty"`
	// country ISO code of the receipient of this invoice
	RecipientCountry string `protobuf:"bytes,12,opt,name=recipient_country,json=recipientCountry" json:"recipient_country,omitempty"`
	// ISO currency code
	Currency string `protobuf:"bytes,13,opt,name=currency" json:"currency,omitempty"`
	// invoice amount including tax
	GrossAmount int64 `protobuf:"varint,14,opt,name=gross_amount,json=grossAmount" json:"gross_amount,omitempty"`
	// invoice amount excluding tax
	NetAmount            int64                `protobuf:"varint,15,opt,name=net_amount,json=netAmount" json:"net_amount,omitempty"`
	TaxAmount            int64                `protobuf:"varint,16,opt,name=tax_amount,json=taxAmount" json:"tax_amount,omitempty"`
	TaxRate              int64                `protobuf:"varint,17,opt,name=tax_rate,json=taxRate" json:"tax_rate,omitempty"`
	Recipient            []byte               `protobuf:"bytes,18,opt,name=recipient,proto3" json:"recipient,omitempty"`
	Sender               []byte               `protobuf:"bytes,19,opt,name=sender,proto3" json:"sender,omitempty"`
	Payee                []byte               `protobuf:"bytes,20,opt,name=payee,proto3" json:"payee,omitempty"`
	Comment              string               `protobuf:"bytes,21,opt,name=comment" json:"comment,omitempty"`
	DueDate              *timestamp.Timestamp `protobuf:"bytes,22,opt,name=due_date,json=dueDate" json:"due_date,omitempty"`
	DateCreated          *timestamp.Timestamp `protobuf:"bytes,23,opt,name=date_created,json=dateCreated" json:"date_created,omitempty"`
	ExtraData            []byte               `protobuf:"bytes,24,opt,name=extra_data,json=extraData,proto3" json:"extra_data,omitempty"`
}


func (i *Invoice) CoreDocument() *coredocumentpb.CoreDocument {
	panic("implement me")
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
