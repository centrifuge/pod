package invoice

import (
	"encoding/json"
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
)

const prefix string = "invoice"

// tree prefixes for specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 1, 0, 0} }

// Invoice implements the documents.Model keeps track of invoice related fields and state
type Invoice struct {
	*documents.CoreDocument

	Number                   string // invoice number or reference number
	Status                   string // invoice status
	SenderInvoiceID          string
	RecipientInvoiceID       string
	SenderCompanyName        string
	SenderContactPersonName  string
	SenderStreet1            string // street and address details of the sender company
	SenderStreet2            string
	SenderCity               string
	SenderZipcode            string
	SenderState              string
	SenderCountry            string // country ISO code of the sender of this invoice
	BillToCompanyName        string
	BillToContactPersonName  string
	BillToStreet1            string
	BillToStreet2            string
	BillToCity               string
	BillToZipcode            string
	BillToState              string
	BillToCountry            string
	BillToVatNumber          string
	BillToLocalTaxID         string
	RemitToCompanyName       string
	RemitToContactPersonName string
	RemitToStreet1           string
	RemitToStreet2           string
	RemitToCity              string
	RemitToZipcode           string
	RemitToState             string
	RemitToCountry           string
	RemitToVatNumber         string
	RemitToLocalTaxID        string
	RemitToTaxCountry        string
	ShipToCompanyName        string
	ShipToContactPersonName  string
	ShipToStreet1            string
	ShipToStreet2            string
	ShipToCity               string
	ShipToZipcode            string
	ShipToState              string
	ShipToCountry            string
	Currency                 string             // ISO currency code
	GrossAmount              *documents.Decimal // invoice amount including tax
	NetAmount                *documents.Decimal // invoice amount excluding tax
	TaxAmount                *documents.Decimal
	TaxRate                  *documents.Decimal
	TaxOnLineLevel           bool
	Recipient                *identity.DID // centrifuge ID of the recipient
	Sender                   *identity.DID // centrifuge ID of the sender
	Payee                    *identity.DID // centrifuge ID of the payee
	Comment                  string
	ShippingTerms            string
	RequesterEmail           string
	RequesterName            string
	DeliveryNumber           string // number of the delivery note
	IsCreditNote             bool
	CreditNoteInvoiceNumber  string
	CreditForInvoiceDate     *timestamp.Timestamp
	DateDue                  *timestamp.Timestamp
	DatePaid                 *timestamp.Timestamp
	DateUpdated              *timestamp.Timestamp
	DateCreated              *timestamp.Timestamp
	Attachments              []*BinaryAttachment
	LineItems                []*LineItem
	PaymentDetails           []*PaymentDetails
	TaxItems                 []*TaxItem
}

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
	Amount                *documents.Decimal
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

// LineItem represents a single invoice line item.
type LineItem struct {
	ItemNumber              string
	Description             string
	SenderPartNo            string
	PricePerUnit            *documents.Decimal
	Quantity                *documents.Decimal
	UnitOfMeasure           string
	NetWeight               *documents.Decimal
	TaxAmount               *documents.Decimal
	TaxRate                 *documents.Decimal
	TaxCode                 *documents.Decimal
	TotalAmount             *documents.Decimal // the total amount of the line item
	PurchaseOrderNumber     string
	PurchaseOrderItemNumber string
	DeliveryNoteNumber      string
}

// TaxItem represents a single invoice tax item.
type TaxItem struct {
	ItemNumber        string
	InvoiceItemNumber string
	TaxAmount         *documents.Decimal
	TaxRate           *documents.Decimal
	TaxCode           *documents.Decimal
	TaxBaseAmount     *documents.Decimal
}

// getClientData returns the client data from the invoice model
func (i *Invoice) getClientData() *clientinvoicepb.InvoiceData {
	decs := documents.DecimalsToStrings(i.GrossAmount, i.NetAmount, i.TaxAmount, i.TaxRate)
	dids := identity.DIDsToStrings(i.Recipient, i.Sender, i.Payee)
	return &clientinvoicepb.InvoiceData{
		Number:                   i.Number,
		Status:                   i.Status,
		SenderInvoiceId:          i.SenderInvoiceID,
		RecipientInvoiceId:       i.RecipientInvoiceID,
		SenderCompanyName:        i.SenderCompanyName,
		SenderContactPersonName:  i.SenderContactPersonName,
		SenderStreet1:            i.SenderStreet1,
		SenderStreet2:            i.SenderStreet2,
		SenderCity:               i.SenderCity,
		SenderZipcode:            i.SenderZipcode,
		SenderState:              i.SenderState,
		SenderCountry:            i.SenderCountry,
		BillToCompanyName:        i.BillToCompanyName,
		BillToContactPersonName:  i.BillToContactPersonName,
		BillToStreet1:            i.BillToStreet1,
		BillToStreet2:            i.BillToStreet2,
		BillToCity:               i.BillToCity,
		BillToZipcode:            i.BillToZipcode,
		BillToState:              i.BillToState,
		BillToCountry:            i.BillToCountry,
		BillToLocalTaxId:         i.BillToLocalTaxID,
		BillToVatNumber:          i.BillToVatNumber,
		RemitToCompanyName:       i.RemitToCompanyName,
		RemitToContactPersonName: i.RemitToContactPersonName,
		RemitToStreet1:           i.RemitToStreet1,
		RemitToStreet2:           i.RemitToStreet2,
		RemitToCity:              i.RemitToCity,
		RemitToCountry:           i.RemitToCountry,
		RemitToState:             i.RemitToState,
		RemitToZipcode:           i.RemitToZipcode,
		RemitToLocalTaxId:        i.RemitToLocalTaxID,
		RemitToTaxCountry:        i.RemitToTaxCountry,
		RemitToVatNumber:         i.RemitToVatNumber,
		ShipToCompanyName:        i.ShipToCompanyName,
		ShipToContactPersonName:  i.ShipToContactPersonName,
		ShipToStreet1:            i.ShipToStreet1,
		ShipToStreet2:            i.ShipToStreet2,
		ShipToCity:               i.ShipToCity,
		ShipToState:              i.ShipToState,
		ShipToCountry:            i.ShipToCountry,
		ShipToZipcode:            i.ShipToZipcode,
		Currency:                 i.Currency,
		GrossAmount:              decs[0],
		NetAmount:                decs[1],
		TaxAmount:                decs[2],
		TaxRate:                  decs[3],
		TaxOnLineLevel:           i.TaxOnLineLevel,
		Recipient:                dids[0],
		Sender:                   dids[1],
		Payee:                    dids[2],
		Comment:                  i.Comment,
		ShippingTerms:            i.ShippingTerms,
		RequesterEmail:           i.RequesterEmail,
		RequesterName:            i.RequesterName,
		DeliveryNumber:           i.DeliveryNumber,
		IsCreditNote:             i.IsCreditNote,
		CreditNoteInvoiceNumber:  i.CreditNoteInvoiceNumber,
		CreditForInvoiceDate:     i.CreditForInvoiceDate,
		DateDue:                  i.DateDue,
		DatePaid:                 i.DatePaid,
		DateCreated:              i.DateCreated,
		DateUpdated:              i.DateUpdated,
		Attachments:              toClientAttachments(i.Attachments),
		LineItems:                toClientLineItems(i.LineItems),
		PaymentDetails:           toClientPaymentDetails(i.PaymentDetails),
		TaxItems:                 toClientTaxItems(i.TaxItems),
	}

}

// createP2PProtobuf returns centrifuge protobuf specific invoiceData
func (i *Invoice) createP2PProtobuf() (data *invoicepb.InvoiceData, err error) {
	decs, err := documents.DecimalsToBytes(i.GrossAmount, i.NetAmount, i.TaxAmount, i.TaxRate)
	if err != nil {
		return nil, err
	}

	li, err := toP2PLineItems(i.LineItems)
	if err != nil {
		return nil, err
	}

	pd, err := toP2PPaymentDetails(i.PaymentDetails)
	if err != nil {
		return nil, err
	}

	ti, err := toP2PTaxItems(i.TaxItems)
	if err != nil {
		return nil, err
	}

	dids := identity.DIDsToBytes(i.Recipient, i.Sender, i.Payee)
	return &invoicepb.InvoiceData{
		Number:                   i.Number,
		Status:                   i.Status,
		SenderInvoiceId:          i.SenderInvoiceID,
		RecipientInvoiceId:       i.RecipientInvoiceID,
		SenderCompanyName:        i.SenderCompanyName,
		SenderContactPersonName:  i.SenderContactPersonName,
		SenderStreet1:            i.SenderStreet1,
		SenderStreet2:            i.SenderStreet2,
		SenderCity:               i.SenderCity,
		SenderZipcode:            i.SenderZipcode,
		SenderState:              i.SenderState,
		SenderCountry:            i.SenderCountry,
		BillToCompanyName:        i.BillToCompanyName,
		BillToContactPersonName:  i.BillToContactPersonName,
		BillToStreet1:            i.BillToStreet1,
		BillToStreet2:            i.BillToStreet2,
		BillToCity:               i.BillToCity,
		BillToZipcode:            i.BillToZipcode,
		BillToState:              i.BillToState,
		BillToCountry:            i.BillToCountry,
		BillToLocalTaxId:         i.BillToLocalTaxID,
		BillToVatNumber:          i.BillToVatNumber,
		RemitToCompanyName:       i.RemitToCompanyName,
		RemitToContactPersonName: i.RemitToContactPersonName,
		RemitToStreet1:           i.RemitToStreet1,
		RemitToStreet2:           i.RemitToStreet2,
		RemitToCity:              i.RemitToCity,
		RemitToCountry:           i.RemitToCountry,
		RemitToState:             i.RemitToState,
		RemitToZipcode:           i.RemitToZipcode,
		RemitToLocalTaxId:        i.RemitToLocalTaxID,
		RemitToTaxCountry:        i.RemitToTaxCountry,
		RemitToVatNumber:         i.RemitToVatNumber,
		ShipToCompanyName:        i.ShipToCompanyName,
		ShipToContactPersonName:  i.ShipToContactPersonName,
		ShipToStreet1:            i.ShipToStreet1,
		ShipToStreet2:            i.ShipToStreet2,
		ShipToCity:               i.ShipToCity,
		ShipToState:              i.ShipToState,
		ShipToCountry:            i.ShipToCountry,
		ShipToZipcode:            i.ShipToZipcode,
		Currency:                 i.Currency,
		GrossAmount:              decs[0],
		NetAmount:                decs[1],
		TaxAmount:                decs[2],
		TaxRate:                  decs[3],
		//TaxOnLineLevel:           i.TaxOnLineLevel,
		Recipient:      dids[0],
		Sender:         dids[1],
		Payee:          dids[2],
		Comment:        i.Comment,
		ShippingTerms:  i.ShippingTerms,
		RequesterEmail: i.RequesterEmail,
		RequesterName:  i.RequesterName,
		DeliveryNumber: i.DeliveryNumber,
		//IsCreditNote:             i.IsCreditNote,
		CreditNoteInvoiceNumber: i.CreditNoteInvoiceNumber,
		CreditForInvoiceDate:    i.CreditForInvoiceDate,
		DateDue:                 i.DateDue,
		DatePaid:                i.DatePaid,
		DateCreated:             i.DateCreated,
		DateUpdated:             i.DateUpdated,
		Attachments:             toP2PAttachments(i.Attachments),
		LineItems:               li,
		PaymentDetails:          pd,
		TaxItems:                ti,
	}, nil

}

// InitInvoiceInput initialize the model based on the received parameters from the rest api call
func (i *Invoice) InitInvoiceInput(payload *clientinvoicepb.InvoiceCreatePayload, self string) error {
	err := i.initInvoiceFromData(payload.Data)
	if err != nil {
		return err
	}

	collaborators := append([]string{self}, payload.Collaborators...)
	cd, err := documents.NewCoreDocumentWithCollaborators(collaborators, compactPrefix())
	if err != nil {
		return errors.New("failed to init core document: %v", err)
	}

	i.CoreDocument = cd
	return nil
}

// initInvoiceFromData initialises invoice from invoiceData
func (i *Invoice) initInvoiceFromData(data *clientinvoicepb.InvoiceData) error {
	decs, err := documents.StringsToDecimals(data.GrossAmount, data.NetAmount, data.TaxAmount, data.TaxRate)
	if err != nil {
		return err
	}

	dids, err := identity.StringsToDIDs(data.Recipient, data.Sender, data.Payee)
	if err != nil {
		return err
	}

	atts, err := fromClientAttachments(data.Attachments)
	if err != nil {
		return err
	}

	li, err := fromClientLineItems(data.LineItems)
	if err != nil {
		return err
	}

	pd, err := fromClientPaymentDetails(data.PaymentDetails)
	if err != nil {
		return err
	}

	ti, err := fromClientTaxItems(data.TaxItems)
	if err != nil {
		return err
	}

	i.Number = data.Number
	i.Status = data.Status
	i.SenderInvoiceID = data.SenderInvoiceId
	i.RecipientInvoiceID = data.RecipientInvoiceId
	i.SenderCompanyName = data.SenderCompanyName
	i.SenderContactPersonName = data.SenderContactPersonName
	i.SenderStreet1 = data.SenderStreet1
	i.SenderStreet2 = data.SenderStreet2
	i.SenderCity = data.SenderCity
	i.SenderZipcode = data.SenderZipcode
	i.SenderState = data.SenderState
	i.SenderCountry = data.SenderCountry
	i.BillToCompanyName = data.BillToCompanyName
	i.BillToContactPersonName = data.BillToContactPersonName
	i.BillToStreet1 = data.BillToStreet1
	i.BillToStreet2 = data.BillToStreet2
	i.BillToCity = data.BillToCity
	i.BillToZipcode = data.BillToZipcode
	i.BillToState = data.BillToState
	i.BillToCountry = data.BillToCountry
	i.BillToVatNumber = data.BillToVatNumber
	i.BillToLocalTaxID = data.BillToLocalTaxId
	i.RemitToCompanyName = data.RemitToCompanyName
	i.RemitToContactPersonName = data.RemitToContactPersonName
	i.RemitToStreet1 = data.RemitToStreet1
	i.RemitToStreet2 = data.RemitToStreet2
	i.RemitToCity = data.RemitToCity
	i.RemitToZipcode = data.RemitToZipcode
	i.RemitToState = data.RemitToState
	i.RemitToCountry = data.RemitToCountry
	i.RemitToVatNumber = data.RemitToVatNumber
	i.RemitToLocalTaxID = data.RemitToLocalTaxId
	i.RemitToTaxCountry = data.RemitToTaxCountry
	i.ShipToCompanyName = data.ShipToCompanyName
	i.ShipToContactPersonName = data.ShipToContactPersonName
	i.ShipToStreet1 = data.ShipToStreet1
	i.ShipToStreet2 = data.ShipToStreet2
	i.ShipToCity = data.ShipToCity
	i.ShipToZipcode = data.ShipToZipcode
	i.ShipToState = data.ShipToState
	i.ShipToCountry = data.ShipToCountry
	i.Currency = data.Currency
	i.GrossAmount = decs[0]
	i.NetAmount = decs[1]
	i.TaxAmount = decs[2]
	i.TaxRate = decs[3]
	i.TaxOnLineLevel = data.TaxOnLineLevel
	i.Recipient = dids[0]
	i.Sender = dids[1]
	i.Payee = dids[2]
	i.Comment = data.Comment
	i.ShippingTerms = data.ShippingTerms
	i.RequesterEmail = data.RequesterEmail
	i.RequesterName = data.RequesterName
	i.DeliveryNumber = data.DeliveryNumber
	i.IsCreditNote = data.IsCreditNote
	i.CreditNoteInvoiceNumber = data.CreditNoteInvoiceNumber
	i.CreditForInvoiceDate = data.CreditForInvoiceDate
	i.DateDue = data.DateDue
	i.DatePaid = data.DatePaid
	i.DateUpdated = data.DateUpdated
	i.DateCreated = data.DateCreated
	i.Attachments = atts
	i.LineItems = li
	i.PaymentDetails = pd
	i.TaxItems = ti
	return nil
}

// loadFromP2PProtobuf  loads the invoice from centrifuge protobuf invoice data
func (i *Invoice) loadFromP2PProtobuf(data *invoicepb.InvoiceData) error {
	decs, err := documents.BytesToDecimals(data.GrossAmount, data.NetAmount, data.TaxAmount, data.TaxRate)
	if err != nil {
		return err
	}

	dids := identity.BytesToDIDs(data.Recipient, data.Sender, data.Payee)
	atts := fromP2PAttachments(data.Attachments)
	li, err := fromP2PLineItems(data.LineItems)
	if err != nil {
		return err
	}

	pd, err := fromP2PPaymentDetails(data.PaymentDetails)
	if err != nil {
		return err
	}

	ti, err := fromP2PTaxItems(data.TaxItems)
	if err != nil {
		return err
	}

	i.Number = data.Number
	i.Status = data.Status
	i.SenderInvoiceID = data.SenderInvoiceId
	i.RecipientInvoiceID = data.RecipientInvoiceId
	i.SenderCompanyName = data.SenderCompanyName
	i.SenderContactPersonName = data.SenderContactPersonName
	i.SenderStreet1 = data.SenderStreet1
	i.SenderStreet2 = data.SenderStreet2
	i.SenderCity = data.SenderCity
	i.SenderZipcode = data.SenderZipcode
	i.SenderState = data.SenderState
	i.SenderCountry = data.SenderCountry
	i.BillToCompanyName = data.BillToCompanyName
	i.BillToContactPersonName = data.BillToContactPersonName
	i.BillToStreet1 = data.BillToStreet1
	i.BillToStreet2 = data.BillToStreet2
	i.BillToCity = data.BillToCity
	i.BillToZipcode = data.BillToZipcode
	i.BillToState = data.BillToState
	i.BillToCountry = data.BillToCountry
	i.BillToVatNumber = data.BillToVatNumber
	i.BillToLocalTaxID = data.BillToLocalTaxId
	i.RemitToCompanyName = data.RemitToCompanyName
	i.RemitToContactPersonName = data.RemitToContactPersonName
	i.RemitToStreet1 = data.RemitToStreet1
	i.RemitToStreet2 = data.RemitToStreet2
	i.RemitToCity = data.RemitToCity
	i.RemitToZipcode = data.RemitToZipcode
	i.RemitToState = data.RemitToState
	i.RemitToCountry = data.RemitToCountry
	i.RemitToVatNumber = data.RemitToVatNumber
	i.RemitToLocalTaxID = data.RemitToLocalTaxId
	i.RemitToTaxCountry = data.RemitToTaxCountry
	i.ShipToCompanyName = data.ShipToCompanyName
	i.ShipToContactPersonName = data.ShipToContactPersonName
	i.ShipToStreet1 = data.ShipToStreet1
	i.ShipToStreet2 = data.ShipToStreet2
	i.ShipToCity = data.ShipToCity
	i.ShipToZipcode = data.ShipToZipcode
	i.ShipToState = data.ShipToState
	i.ShipToCountry = data.ShipToCountry
	i.Currency = data.Currency
	i.GrossAmount = decs[0]
	i.NetAmount = decs[1]
	i.TaxAmount = decs[2]
	i.TaxRate = decs[3]
	// TODO(ved): enable these after precise proofs are integrated
	//i.TaxOnLineLevel = data.TaxOnLineLevel
	i.Recipient = dids[0]
	i.Sender = dids[1]
	i.Payee = dids[2]
	i.Comment = data.Comment
	i.ShippingTerms = data.ShippingTerms
	i.RequesterEmail = data.RequesterEmail
	i.RequesterName = data.RequesterName
	i.DeliveryNumber = data.DeliveryNumber
	//i.IsCreditNote = data.IsCreditNote
	i.CreditNoteInvoiceNumber = data.CreditNoteInvoiceNumber
	i.CreditForInvoiceDate = data.CreditForInvoiceDate
	i.DateDue = data.DateDue
	i.DatePaid = data.DatePaid
	i.DateUpdated = data.DateUpdated
	i.DateCreated = data.DateCreated
	i.Attachments = atts
	i.LineItems = li
	i.PaymentDetails = pd
	i.TaxItems = ti
	return nil
}

// PackCoreDocument packs the Invoice into a CoreDocument.
func (i *Invoice) PackCoreDocument() (cd coredocumentpb.CoreDocument, err error) {
	invData, err := i.createP2PProtobuf()
	if err != nil {
		return cd, err
	}

	data, err := proto.Marshal(invData)
	if err != nil {
		return cd, errors.New("couldn't serialise InvoiceData: %v", err)
	}

	embedData := &any.Any{
		TypeUrl: i.DocumentType(),
		Value:   data,
	}
	return i.CoreDocument.PackCoreDocument(embedData), nil
}

// UnpackCoreDocument unpacks the core document into Invoice.
func (i *Invoice) UnpackCoreDocument(cd coredocumentpb.CoreDocument) error {
	if cd.EmbeddedData == nil ||
		cd.EmbeddedData.TypeUrl != i.DocumentType() {
		return errors.New("trying to convert document with incorrect schema")
	}

	invoiceData := new(invoicepb.InvoiceData)
	err := proto.Unmarshal(cd.EmbeddedData.Value, invoiceData)
	if err != nil {
		return err
	}

	if err := i.loadFromP2PProtobuf(invoiceData); err != nil {
		return err
	}

	i.CoreDocument = documents.NewCoreDocumentFromProtobuf(cd)
	return nil
}

// JSON marshals Invoice into a json bytes
func (i *Invoice) JSON() ([]byte, error) {
	return json.Marshal(i)
}

// FromJSON unmarshals the json bytes into Invoice
func (i *Invoice) FromJSON(jsonData []byte) error {
	return json.Unmarshal(jsonData, i)
}

// Type gives the Invoice type
func (i *Invoice) Type() reflect.Type {
	return reflect.TypeOf(i)
}

// CalculateDataRoot calculates the data root and sets the root to core document.
func (i *Invoice) CalculateDataRoot() ([]byte, error) {
	t, err := i.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("failed to get data tree: %v", err)
	}

	dr := t.RootHash()
	i.CoreDocument.SetDataRoot(dr)
	return dr, nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (i *Invoice) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	invProto, err := i.createP2PProtobuf()
	if err != nil {
		return nil, err
	}
	if i.CoreDocument == nil {
		return nil, errors.New("getDocumentDataTree error CoreDocument not set")
	}
	t := i.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	err = t.AddLeavesFromDocument(invProto)
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}
	err = t.Generate()
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}

	i.SetDataModified(false)
	return t, nil
}

// CreateProofs generates proofs for given fields.
func (i *Invoice) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	tree, err := i.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	return i.CoreDocument.CreateProofs(i.DocumentType(), tree, fields)
}

// DocumentType returns the invoice document type.
func (*Invoice) DocumentType() string {
	return documenttypes.InvoiceDataTypeUrl
}

// PrepareNewVersion prepares new version from the old invoice.
func (i *Invoice) PrepareNewVersion(old documents.Model, data *clientinvoicepb.InvoiceData, collaborators []string) error {
	err := i.initInvoiceFromData(data)
	if err != nil {
		return err
	}

	oldCD := old.(*Invoice).CoreDocument
	i.CoreDocument, err = oldCD.PrepareNewVersion(collaborators, compactPrefix())
	if err != nil {
		return err
	}

	i.DataModified = true
	return nil
}

// AddNFT adds NFT to the Invoice.
func (i *Invoice) AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte) error {
	cd, err := i.CoreDocument.AddNFT(grantReadAccess, registry, tokenID)
	if err != nil {
		return err
	}

	i.CoreDocument = cd
	return nil
}

// CalculateSigningRoot calculates the signing root of the document.
func (i *Invoice) CalculateSigningRoot() ([]byte, error) {
	return i.CoreDocument.CalculateSigningRoot(i.DocumentType())
}

// CalculateDocumentRoot calculate the document root
// TODO: Should we add this
func (i *Invoice) CalculateDocumentRoot() ([]byte, error) {
	return i.CoreDocument.CalculateDocumentRoot()
}

// CreateNFTProofs creates proofs specific to NFT minting.
func (i *Invoice) CreateNFTProofs(
	account identity.DID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error) {

	tree, err := i.getDocumentDataTree()
	if err != nil {
		return nil, err
	}

	return i.CoreDocument.CreateNFTProofs(
		i.DocumentType(),
		tree,
		account, registry, tokenID, nftUniqueProof, readAccessProof)
}

// CollaboratorCanUpdate checks if the collaborator can update the document.
func (i *Invoice) CollaboratorCanUpdate(updated documents.Model, collaborator identity.DID) error {
	newInv, ok := updated.(*Invoice)
	if !ok {
		return errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("expecting an invoice but got %T", updated))
	}

	// check the core document changes
	err := i.CoreDocument.CollaboratorCanUpdate(newInv.CoreDocument, collaborator, i.DocumentType())
	if err != nil {
		return err
	}

	// check invoice specific changes
	oldTree, err := i.getDocumentDataTree()
	if err != nil {
		return err
	}

	newTree, err := newInv.getDocumentDataTree()
	if err != nil {
		return err
	}

	rules := i.CoreDocument.TransitionRulesFor(collaborator)
	cf := documents.GetChangedFields(oldTree, newTree)
	return documents.ValidateTransitions(rules, cf)
}
