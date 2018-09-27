package invoice

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// InvoiceModel specific deriver
type ModelDeriver interface {
	// Embedded ModelDeriver
	documents.ModelDeriver

	DeriveWithInvoiceInput(*InvoiceInput) (*documents.Model, error)
}

//InvoiceInput represents the parameter received from the rest api
type InvoiceInput struct {

	// TODO add new parameters according to new client API
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
}
