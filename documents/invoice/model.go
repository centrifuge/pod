package invoice

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils/timeutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/jinzhu/copier"
)

const (
	prefix string = "invoice"

	// Scheme is the scheme used by invoice documents
	Scheme = prefix

	// ErrInvoiceInvalidData sentinel error when data unmarshal is failed.
	ErrInvoiceInvalidData = errors.Error("invalid invoice data")
)

// tree prefixes for specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 1, 0, 0} }

// Data holds the invoice specific fields.
type Data struct {
	Number                   string                        `json:"number"` // invoice number or reference number
	Status                   string                        `json:"status"` // invoice status
	SenderInvoiceID          string                        `json:"sender_invoice_id"`
	RecipientInvoiceID       string                        `json:"recipient_invoice_id"`
	SenderCompanyName        string                        `json:"sender_company_name"`
	SenderContactPersonName  string                        `json:"sender_contact_person_name"`
	SenderStreet1            string                        `json:"sender_street_1"` // street and address details of the sender company
	SenderStreet2            string                        `json:"sender_street_2"`
	SenderCity               string                        `json:"sender_city"`
	SenderZipcode            string                        `json:"sender_zipcode"`
	SenderState              string                        `json:"sender_state"`
	SenderCountry            string                        `json:"sender_country"` // country ISO code of the sender of this invoice
	BillToCompanyName        string                        `json:"bill_to_company_name"`
	BillToContactPersonName  string                        `json:"bill_to_contact_person_name"`
	BillToStreet1            string                        `json:"bill_to_street_1"`
	BillToStreet2            string                        `json:"bill_to_street_2"`
	BillToCity               string                        `json:"bill_to_city"`
	BillToZipcode            string                        `json:"bill_to_zipcode"`
	BillToState              string                        `json:"bill_to_state"`
	BillToCountry            string                        `json:"bill_to_country"`
	BillToVatNumber          string                        `json:"bill_to_vat_number"`
	BillToLocalTaxID         string                        `json:"bill_to_local_tax_id"`
	RemitToCompanyName       string                        `json:"remit_to_company_name"`
	RemitToContactPersonName string                        `json:"remit_to_contact_person_name"`
	RemitToStreet1           string                        `json:"remit_to_street_1"`
	RemitToStreet2           string                        `json:"remit_to_street_2"`
	RemitToCity              string                        `json:"remit_to_city"`
	RemitToZipcode           string                        `json:"remit_to_zipcode"`
	RemitToState             string                        `json:"remit_to_state"`
	RemitToCountry           string                        `json:"remit_to_country"`
	RemitToVatNumber         string                        `json:"remit_to_vat_number"`
	RemitToLocalTaxID        string                        `json:"remit_to_local_tax_id"`
	RemitToTaxCountry        string                        `json:"remit_to_tax_country"`
	ShipToCompanyName        string                        `json:"ship_to_company_name"`
	ShipToContactPersonName  string                        `json:"ship_to_contact_person_name"`
	ShipToStreet1            string                        `json:"ship_to_street_1"`
	ShipToStreet2            string                        `json:"ship_to_street_2"`
	ShipToCity               string                        `json:"ship_to_city"`
	ShipToZipcode            string                        `json:"ship_to_zipcode"`
	ShipToState              string                        `json:"ship_to_state"`
	ShipToCountry            string                        `json:"ship_to_country"`
	Currency                 string                        `json:"currency"`                                    // ISO currency code
	GrossAmount              *documents.Decimal            `json:"gross_amount" swaggertype:"primitive,string"` // invoice amount including tax
	NetAmount                *documents.Decimal            `json:"net_amount" swaggertype:"primitive,string"`   // invoice amount excluding tax
	TaxAmount                *documents.Decimal            `json:"tax_amount" swaggertype:"primitive,string"`
	TaxRate                  *documents.Decimal            `json:"tax_rate" swaggertype:"primitive,string"`
	TaxOnLineLevel           bool                          `json:"tax_on_line_level"`
	Recipient                *identity.DID                 `json:"recipient,string" swaggertype:"primitive,string"` // centrifuge ID of the recipient
	Sender                   *identity.DID                 `json:"sender,string" swaggertype:"primitive,string"`    // centrifuge ID of the sender
	Payee                    *identity.DID                 `json:"payee,string" swaggertype:"primitive,string"`     // centrifuge ID of the payee
	Comment                  string                        `json:"comment"`
	ShippingTerms            string                        `json:"shipping_terms"`
	RequesterEmail           string                        `json:"requester_email"`
	RequesterName            string                        `json:"requester_name"`
	DeliveryNumber           string                        `json:"delivery_number"` // number of the delivery note
	IsCreditNote             bool                          `json:"is_credit_note"`
	CreditNoteInvoiceNumber  string                        `json:"credit_note_invoice_number"`
	CreditForInvoiceDate     *time.Time                    `json:"credit_for_invoice_date"`
	DateDue                  *time.Time                    `json:"date_due" swaggertype:"primitive,string"`
	DatePaid                 *time.Time                    `json:"date_paid" swaggertype:"primitive,string"`
	DateUpdated              *time.Time                    `json:"date_updated" swaggertype:"primitive,string"`
	DateCreated              *time.Time                    `json:"date_created" swaggertype:"primitive,string"`
	Attachments              []*documents.BinaryAttachment `json:"attachments"`
	LineItems                []*LineItem                   `json:"line_items"`
	PaymentDetails           []*documents.PaymentDetails   `json:"payment_details"`
	TaxItems                 []*TaxItem                    `json:"tax_items"`
}

// Invoice implements the documents.Model keeps track of invoice related fields and state
type Invoice struct {
	*documents.CoreDocument
	Data Data
}

// LineItem represents a single invoice line item.
type LineItem struct {
	ItemNumber              string             `json:"item_number"`
	Description             string             `json:"description"`
	SenderPartNo            string             `json:"sender_part_no"`
	PricePerUnit            *documents.Decimal `json:"price_per_unit" swaggertype:"primitive,string"`
	Quantity                *documents.Decimal `json:"quantity" swaggertype:"primitive,string"`
	UnitOfMeasure           string             `json:"unit_of_measure"`
	NetWeight               *documents.Decimal `json:"net_weight" swaggertype:"primitive,string"`
	TaxAmount               *documents.Decimal `json:"tax_amount" swaggertype:"primitive,string"`
	TaxRate                 *documents.Decimal `json:"tax_rate" swaggertype:"primitive,string"`
	TaxCode                 *documents.Decimal `json:"tax_code" swaggertype:"primitive,string"`
	TotalAmount             *documents.Decimal `json:"total_amount" swaggertype:"primitive,string"` // the total amount of the line item
	PurchaseOrderNumber     string             `json:"purchase_order_number"`
	PurchaseOrderItemNumber string             `json:"purchase_order_item_number"`
	DeliveryNoteNumber      string             `json:"delivery_note_number"`
}

// TaxItem represents a single invoice tax item.
type TaxItem struct {
	ItemNumber        string             `json:"item_number"`
	InvoiceItemNumber string             `json:"invoice_item_number"`
	TaxAmount         *documents.Decimal `json:"tax_amount" swaggertype:"primitive,string"`
	TaxRate           *documents.Decimal `json:"tax_rate" swaggertype:"primitive,string"`
	TaxCode           *documents.Decimal `json:"tax_code" swaggertype:"primitive,string"`
	TaxBaseAmount     *documents.Decimal `json:"tax_base_amount" swaggertype:"primitive,string"`
}

// createP2PProtobuf returns centrifuge protobuf specific invoiceData
func (i *Invoice) createP2PProtobuf() (data *invoicepb.InvoiceData, err error) {
	d := i.Data
	decs, err := documents.DecimalsToBytes(d.GrossAmount, d.NetAmount, d.TaxAmount, d.TaxRate)
	if err != nil {
		return nil, err
	}

	li, err := toP2PLineItems(d.LineItems)
	if err != nil {
		return nil, err
	}

	pd, err := documents.ToProtocolPaymentDetails(d.PaymentDetails)
	if err != nil {
		return nil, err
	}

	ti, err := toP2PTaxItems(d.TaxItems)
	if err != nil {
		return nil, err
	}

	pts, err := timeutils.ToProtoTimestamps(d.CreditForInvoiceDate, d.DateDue, d.DatePaid, d.DateCreated, d.DateUpdated)
	if err != nil {
		return nil, err
	}

	dids := identity.DIDsToBytes(d.Recipient, d.Sender, d.Payee)
	return &invoicepb.InvoiceData{
		Number:                   d.Number,
		Status:                   d.Status,
		SenderInvoiceId:          d.SenderInvoiceID,
		RecipientInvoiceId:       d.RecipientInvoiceID,
		SenderCompanyName:        d.SenderCompanyName,
		SenderContactPersonName:  d.SenderContactPersonName,
		SenderStreet1:            d.SenderStreet1,
		SenderStreet2:            d.SenderStreet2,
		SenderCity:               d.SenderCity,
		SenderZipcode:            d.SenderZipcode,
		SenderState:              d.SenderState,
		SenderCountry:            d.SenderCountry,
		BillToCompanyName:        d.BillToCompanyName,
		BillToContactPersonName:  d.BillToContactPersonName,
		BillToStreet1:            d.BillToStreet1,
		BillToStreet2:            d.BillToStreet2,
		BillToCity:               d.BillToCity,
		BillToZipcode:            d.BillToZipcode,
		BillToState:              d.BillToState,
		BillToCountry:            d.BillToCountry,
		BillToLocalTaxId:         d.BillToLocalTaxID,
		BillToVatNumber:          d.BillToVatNumber,
		RemitToCompanyName:       d.RemitToCompanyName,
		RemitToContactPersonName: d.RemitToContactPersonName,
		RemitToStreet1:           d.RemitToStreet1,
		RemitToStreet2:           d.RemitToStreet2,
		RemitToCity:              d.RemitToCity,
		RemitToCountry:           d.RemitToCountry,
		RemitToState:             d.RemitToState,
		RemitToZipcode:           d.RemitToZipcode,
		RemitToLocalTaxId:        d.RemitToLocalTaxID,
		RemitToTaxCountry:        d.RemitToTaxCountry,
		RemitToVatNumber:         d.RemitToVatNumber,
		ShipToCompanyName:        d.ShipToCompanyName,
		ShipToContactPersonName:  d.ShipToContactPersonName,
		ShipToStreet1:            d.ShipToStreet1,
		ShipToStreet2:            d.ShipToStreet2,
		ShipToCity:               d.ShipToCity,
		ShipToState:              d.ShipToState,
		ShipToCountry:            d.ShipToCountry,
		ShipToZipcode:            d.ShipToZipcode,
		Currency:                 d.Currency,
		GrossAmount:              decs[0],
		NetAmount:                decs[1],
		TaxAmount:                decs[2],
		TaxRate:                  decs[3],
		TaxOnLineLevel:           d.TaxOnLineLevel,
		Recipient:                dids[0],
		Sender:                   dids[1],
		Payee:                    dids[2],
		Comment:                  d.Comment,
		ShippingTerms:            d.ShippingTerms,
		RequesterEmail:           d.RequesterEmail,
		RequesterName:            d.RequesterName,
		DeliveryNumber:           d.DeliveryNumber,
		IsCreditNote:             d.IsCreditNote,
		CreditNoteInvoiceNumber:  d.CreditNoteInvoiceNumber,
		CreditForInvoiceDate:     pts[0],
		DateDue:                  pts[1],
		DatePaid:                 pts[2],
		DateCreated:              pts[3],
		DateUpdated:              pts[4],
		Attachments:              documents.ToProtocolAttachments(d.Attachments),
		LineItems:                li,
		PaymentDetails:           pd,
		TaxItems:                 ti,
	}, nil

}

// loadFromP2PProtobuf  loads the invoice from centrifuge protobuf invoice data
func (i *Invoice) loadFromP2PProtobuf(data *invoicepb.InvoiceData) error {
	decs, err := documents.BytesToDecimals(data.GrossAmount, data.NetAmount, data.TaxAmount, data.TaxRate)
	if err != nil {
		return err
	}

	dids, err := identity.BytesToDIDs(data.Recipient, data.Sender, data.Payee)
	if err != nil {
		return err
	}

	atts := documents.FromProtocolAttachments(data.Attachments)
	li, err := fromP2PLineItems(data.LineItems)
	if err != nil {
		return err
	}

	pd, err := documents.FromProtocolPaymentDetails(data.PaymentDetails)
	if err != nil {
		return err
	}

	ti, err := fromP2PTaxItems(data.TaxItems)
	if err != nil {
		return err
	}

	tms, err := timeutils.FromProtoTimestamps(data.CreditForInvoiceDate, data.DateDue, data.DatePaid, data.DateCreated, data.DateUpdated)
	if err != nil {
		return err
	}

	var d Data
	d.Number = data.Number
	d.Status = data.Status
	d.SenderInvoiceID = data.SenderInvoiceId
	d.RecipientInvoiceID = data.RecipientInvoiceId
	d.SenderCompanyName = data.SenderCompanyName
	d.SenderContactPersonName = data.SenderContactPersonName
	d.SenderStreet1 = data.SenderStreet1
	d.SenderStreet2 = data.SenderStreet2
	d.SenderCity = data.SenderCity
	d.SenderZipcode = data.SenderZipcode
	d.SenderState = data.SenderState
	d.SenderCountry = data.SenderCountry
	d.BillToCompanyName = data.BillToCompanyName
	d.BillToContactPersonName = data.BillToContactPersonName
	d.BillToStreet1 = data.BillToStreet1
	d.BillToStreet2 = data.BillToStreet2
	d.BillToCity = data.BillToCity
	d.BillToZipcode = data.BillToZipcode
	d.BillToState = data.BillToState
	d.BillToCountry = data.BillToCountry
	d.BillToVatNumber = data.BillToVatNumber
	d.BillToLocalTaxID = data.BillToLocalTaxId
	d.RemitToCompanyName = data.RemitToCompanyName
	d.RemitToContactPersonName = data.RemitToContactPersonName
	d.RemitToStreet1 = data.RemitToStreet1
	d.RemitToStreet2 = data.RemitToStreet2
	d.RemitToCity = data.RemitToCity
	d.RemitToZipcode = data.RemitToZipcode
	d.RemitToState = data.RemitToState
	d.RemitToCountry = data.RemitToCountry
	d.RemitToVatNumber = data.RemitToVatNumber
	d.RemitToLocalTaxID = data.RemitToLocalTaxId
	d.RemitToTaxCountry = data.RemitToTaxCountry
	d.ShipToCompanyName = data.ShipToCompanyName
	d.ShipToContactPersonName = data.ShipToContactPersonName
	d.ShipToStreet1 = data.ShipToStreet1
	d.ShipToStreet2 = data.ShipToStreet2
	d.ShipToCity = data.ShipToCity
	d.ShipToZipcode = data.ShipToZipcode
	d.ShipToState = data.ShipToState
	d.ShipToCountry = data.ShipToCountry
	d.Currency = data.Currency
	d.GrossAmount = decs[0]
	d.NetAmount = decs[1]
	d.TaxAmount = decs[2]
	d.TaxRate = decs[3]
	d.TaxOnLineLevel = data.TaxOnLineLevel
	d.Recipient = dids[0]
	d.Sender = dids[1]
	d.Payee = dids[2]
	d.Comment = data.Comment
	d.ShippingTerms = data.ShippingTerms
	d.RequesterEmail = data.RequesterEmail
	d.RequesterName = data.RequesterName
	d.DeliveryNumber = data.DeliveryNumber
	d.IsCreditNote = data.IsCreditNote
	d.CreditNoteInvoiceNumber = data.CreditNoteInvoiceNumber
	d.CreditForInvoiceDate = tms[0]
	d.DateDue = tms[1]
	d.DatePaid = tms[2]
	d.DateCreated = tms[3]
	d.DateUpdated = tms[4]
	d.Attachments = atts
	d.LineItems = li
	d.PaymentDetails = pd
	d.TaxItems = ti
	i.Data = d
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

	i.CoreDocument, err = documents.NewCoreDocumentFromProtobuf(cd)
	return err
}

// JSON marshals Invoice into a json bytes
func (i *Invoice) JSON() ([]byte, error) {
	return i.CoreDocument.MarshalJSON(i)
}

// FromJSON unmarshals the json bytes into Invoice
func (i *Invoice) FromJSON(jsonData []byte) error {
	if i.CoreDocument == nil {
		i.CoreDocument = new(documents.CoreDocument)
	}

	return i.CoreDocument.UnmarshalJSON(jsonData, i)
}

// Type gives the Invoice type
func (i *Invoice) Type() reflect.Type {
	return reflect.TypeOf(i)
}

// CalculateDataRoot calculates the data root and sets the root to core document.
func (i *Invoice) CalculateDataRoot() ([]byte, error) {
	t, err := i.getDataTree()
	if err != nil {
		return nil, errors.New("failed to get data tree: %v", err)
	}

	return t.RootHash(), nil
}

func (i *Invoice) getDataLeaves() ([]proofs.LeafNode, error) {
	t, err := i.getRawDataTree()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return t.GetLeaves(), nil
}

func (i *Invoice) getRawDataTree() (*proofs.DocumentTree, error) {
	invProto, err := i.createP2PProtobuf()
	if err != nil {
		return nil, err
	}
	if i.CoreDocument == nil {
		return nil, errors.New("getDataTree error CoreDocument not set")
	}
	t, err := i.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	err = t.AddLeavesFromDocument(invProto)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return t, nil
}

// getDataTree creates precise-proofs data tree for the model
func (i *Invoice) getDataTree() (*proofs.DocumentTree, error) {
	tree, err := i.getRawDataTree()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	err = tree.Generate()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}

	return tree, nil
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
	t, err := i.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	if err != nil {
		return nil, err
	}

	err = t.AddLeavesFromDocument(invProto)
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}
	err = t.Generate()
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}

	return t, nil
}

// CreateProofs generates proofs for given fields.
func (i *Invoice) CreateProofs(fields []string) (prf *documents.DocumentProof, err error) {
	dataLeaves, err := i.getDataLeaves()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	return i.CoreDocument.CreateProofs(i.DocumentType(), dataLeaves, fields)
}

// DocumentType returns the invoice document type.
func (*Invoice) DocumentType() string {
	return documenttypes.InvoiceDataTypeUrl
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
	dataLeaves, err := i.getDataLeaves()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return i.CoreDocument.CalculateSigningRoot(i.DocumentType(), dataLeaves)
}

// CalculateDocumentRoot calculates the document root
func (i *Invoice) CalculateDocumentRoot() ([]byte, error) {
	dataLeaves, err := i.getDataLeaves()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return i.CoreDocument.CalculateDocumentRoot(i.DocumentType(), dataLeaves)
}

// DocumentRootTree creates and returns the document root tree
func (i *Invoice) DocumentRootTree() (tree *proofs.DocumentTree, err error) {
	dataLeaves, err := i.getDataLeaves()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return i.CoreDocument.DocumentRootTree(i.DocumentType(), dataLeaves)
}

// CreateNFTProofs creates proofs specific to NFT minting.
func (i *Invoice) CreateNFTProofs(
	account identity.DID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proof *documents.DocumentProof, err error) {

	dataLeaves, err := i.getDataLeaves()
	if err != nil {
		return nil, err
	}

	return i.CoreDocument.CreateNFTProofs(
		i.DocumentType(),
		dataLeaves,
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

// AddAttributes adds attributes to the Invoice model.
func (i *Invoice) AddAttributes(ca documents.CollaboratorsAccess, prepareNewVersion bool, attrs ...documents.Attribute) error {
	ncd, err := i.CoreDocument.AddAttributes(ca, prepareNewVersion, compactPrefix(), attrs...)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDAttribute, err)
	}

	i.CoreDocument = ncd
	return nil
}

// DeleteAttribute deletes the attribute from the model.
func (i *Invoice) DeleteAttribute(key documents.AttrKey, prepareNewVersion bool) error {
	ncd, err := i.CoreDocument.DeleteAttribute(key, prepareNewVersion, compactPrefix())
	if err != nil {
		return errors.NewTypedError(documents.ErrCDAttribute, err)
	}

	i.CoreDocument = ncd
	return nil
}

// GetData returns Invoice Data.
func (i *Invoice) GetData() interface{} {
	return i.Data
}

// loadData unmarshals json blob to Data.
func loadData(data []byte, v *Data) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	return nil
}

// DeriveFromCreatePayload unpacks the invoice data from the Payload.
func (i *Invoice) DeriveFromCreatePayload(_ context.Context, payload documents.CreatePayload) error {
	if err := loadData(payload.Data, &i.Data); err != nil {
		return errors.NewTypedError(ErrInvoiceInvalidData, err)
	}

	cd, err := documents.NewCoreDocument(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDCreate, err)
	}

	i.CoreDocument = cd
	return nil
}

// unpackFromUpdatePayloadOld unpacks the update payload and prepares a new version.
// deprecated
func (i *Invoice) unpackFromUpdatePayloadOld(old *Invoice, payload documents.UpdatePayload) error {
	if err := loadData(payload.Data, &i.Data); err != nil {
		return errors.NewTypedError(ErrInvoiceInvalidData, err)
	}

	ncd, err := old.CoreDocument.PrepareNewVersion(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return err
	}

	i.CoreDocument = ncd
	return nil
}

// DeriveFromUpdatePayload unpacks the update payload and prepares a new version.
func (i *Invoice) DeriveFromUpdatePayload(_ context.Context, payload documents.UpdatePayload) (documents.Model, error) {
	d, err := i.patch(payload)
	if err != nil {
		return nil, err
	}

	ncd, err := i.CoreDocument.PrepareNewVersion(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return nil, err
	}

	return &Invoice{
		Data:         d,
		CoreDocument: ncd,
	}, nil
}

// patch merges payload into model data
func (i *Invoice) patch(payload documents.UpdatePayload) (Data, error) {
	var d Data
	err := copier.Copy(&d, &i.Data)
	if err != nil {
		return d, err
	}

	if err := loadData(payload.Data, &d); err != nil {
		return Data{}, errors.NewTypedError(ErrInvoiceInvalidData, err)
	}

	return d, nil
}

// Patch merges payload data into model
func (i *Invoice) Patch(payload documents.UpdatePayload) error {
	d, err := i.patch(payload)
	if err != nil {
		return err
	}

	ncd, err := i.CoreDocument.Patch(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return err
	}

	i.Data = d
	i.CoreDocument = ncd
	return nil
}

// Scheme returns the invoice Scheme.
func (i *Invoice) Scheme() string {
	return Scheme
}
