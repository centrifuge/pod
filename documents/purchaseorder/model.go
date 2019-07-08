package purchaseorder

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils/timeutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	prefix string = "po"

	// Scheme is purchase order scheme
	Scheme = "purchase_order"

	// ErrPOInvalidData sentinel error when data unmarshal is failed.
	ErrPOInvalidData = errors.Error("invalid purchase order data")
)

// tree prefixes for specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 2, 0, 0} }

// Data represents Purchase Order Data.
type Data struct {
	Status                  string                        `json:"status"`
	Number                  string                        `json:"number"`
	SenderOrderID           string                        `json:"sender_order_id"`
	RecipientOrderID        string                        `json:"recipient_order_id"`
	RequisitionID           string                        `json:"requisition_id"`
	RequesterName           string                        `json:"requester_name"`
	RequesterEmail          string                        `json:"requester_email"`
	ShipToCompanyName       string                        `json:"ship_to_company_name"`
	ShipToContactPersonName string                        `json:"ship_to_contact_person_name"`
	ShipToStreet1           string                        `json:"ship_to_street_1"`
	ShipToStreet2           string                        `json:"ship_to_street_2"`
	ShipToCity              string                        `json:"ship_to_city"`
	ShipToZipcode           string                        `json:"ship_to_zipcode"`
	ShipToState             string                        `json:"ship_to_state"`
	ShipToCountry           string                        `json:"ship_to_country"`
	PaymentTerms            string                        `json:"payment_terms"`
	Currency                string                        `json:"currency"`
	TotalAmount             *documents.Decimal            `json:"total_amount" swaggertype:"primitive,string"`
	Recipient               *identity.DID                 `json:"recipient" swaggertype:"primitive,string"`
	Sender                  *identity.DID                 `json:"sender" swaggertype:"primitive,string"`
	Comment                 string                        `json:"comment"`
	DateSent                *time.Time                    `json:"date_sent" swaggertype:"primitive,string"`
	DateConfirmed           *time.Time                    `json:"date_confirmed" swaggertype:"primitive,string"`
	DateUpdated             *time.Time                    `json:"date_updated" swaggertype:"primitive,string"`
	DateCreated             *time.Time                    `json:"date_created" swaggertype:"primitive,string"`
	Attachments             []*documents.BinaryAttachment `json:"attachments"`
	LineItems               []*LineItem                   `json:"line_items"`
	PaymentDetails          []*documents.PaymentDetails   `json:"payment_details"`
}

// PurchaseOrder implements the documents.Model keeps track of purchase order related fields and state
type PurchaseOrder struct {
	*documents.CoreDocument
	Data Data `json:"data"`
}

// LineItemActivity describes a single line item activity.
type LineItemActivity struct {
	ItemNumber            string             `json:"item_number"`
	Status                string             `json:"status"`
	Quantity              *documents.Decimal `json:"quantity" swaggertype:"primitive,string"`
	Amount                *documents.Decimal `json:"amount" swaggertype:"primitive,string"`
	ReferenceDocumentID   string             `json:"reference_document_id"`
	ReferenceDocumentItem string             `json:"reference_document_item"`
	Date                  *time.Time         `json:"date" swaggertype:"primitive,string"`
}

// TaxItem describes a single Purchase Order tax item.
type TaxItem struct {
	ItemNumber              string             `json:"item_number"`
	PurchaseOrderItemNumber string             `json:"purchase_order_item_number"`
	TaxAmount               *documents.Decimal `json:"tax_amount" swaggertype:"primitive,string"`
	TaxRate                 *documents.Decimal `json:"tax_rate" swaggertype:"primitive,string"`
	TaxCode                 *documents.Decimal `json:"tax_code" swaggertype:"primitive,string"`
	TaxBaseAmount           *documents.Decimal `json:"tax_base_amount" swaggertype:"primitive,string"`
}

// LineItem describes a single LineItem Activity
type LineItem struct {
	Status            string              `json:"status"`
	ItemNumber        string              `json:"item_number"`
	Description       string              `json:"description"`
	AmountInvoiced    *documents.Decimal  `json:"amount_invoiced" swaggertype:"primitive,string"`
	AmountTotal       *documents.Decimal  `json:"amount_total" swaggertype:"primitive,string"`
	RequisitionNumber string              `json:"requisition_number"`
	RequisitionItem   string              `json:"requisition_item"`
	PartNumber        string              `json:"part_number"`
	PricePerUnit      *documents.Decimal  `json:"price_per_unit" swaggertype:"primitive,string"`
	UnitOfMeasure     *documents.Decimal  `json:"unit_of_measure" swaggertype:"primitive,string"`
	Quantity          *documents.Decimal  `json:"quantity" swaggertype:"primitive,string"`
	ReceivedQuantity  *documents.Decimal  `json:"received_quantity" swaggertype:"primitive,string"`
	DateUpdated       *time.Time          `json:"date_updated" swaggertype:"primitive,string"`
	DateCreated       *time.Time          `json:"date_created" swaggertype:"primitive,string"`
	RevisionNumber    int                 `json:"revision_number"`
	Activities        []*LineItemActivity `json:"activities"`
	TaxItems          []*TaxItem          `json:"tax_items"`
}

// createP2PProtobuf returns centrifuge protobuf specific purchaseOrderData
func (p *PurchaseOrder) createP2PProtobuf() (*purchaseorderpb.PurchaseOrderData, error) {
	data := p.Data
	decs, err := documents.DecimalsToBytes(data.TotalAmount)
	if err != nil {
		return nil, err
	}

	pd, err := documents.ToProtocolPaymentDetails(data.PaymentDetails)
	if err != nil {
		return nil, err
	}

	li, err := toP2PLineItems(data.LineItems)
	if err != nil {
		return nil, err
	}

	pts, err := timeutils.ToProtoTimestamps(data.DateCreated, data.DateUpdated, data.DateConfirmed, data.DateSent)
	if err != nil {
		return nil, err
	}

	dids := identity.DIDsToBytes(data.Recipient, data.Sender)
	return &purchaseorderpb.PurchaseOrderData{
		Status:                  data.Status,
		Number:                  data.Number,
		SenderOrderId:           data.SenderOrderID,
		TotalAmount:             decs[0],
		Recipient:               dids[0],
		Sender:                  dids[1],
		DateCreated:             pts[0],
		DateUpdated:             pts[1],
		RequesterName:           data.RequesterName,
		RequesterEmail:          data.RequesterEmail,
		Comment:                 data.Comment,
		Currency:                data.Currency,
		ShipToCountry:           data.ShipToCountry,
		ShipToState:             data.ShipToState,
		ShipToZipcode:           data.ShipToZipcode,
		ShipToCity:              data.ShipToCity,
		ShipToStreet1:           data.ShipToStreet1,
		ShipToStreet2:           data.ShipToStreet2,
		ShipToContactPersonName: data.ShipToContactPersonName,
		ShipToCompanyName:       data.ShipToCompanyName,
		DateConfirmed:           pts[2],
		DateSent:                pts[3],
		PaymentTerms:            data.PaymentTerms,
		RecipientOrderId:        data.RecipientOrderID,
		RequisitionId:           data.RequisitionID,
		PaymentDetails:          pd,
		Attachments:             documents.ToProtocolAttachments(data.Attachments),
		LineItems:               li,
	}, nil

}

// loadFromP2PProtobuf loads the purcase order from centrifuge protobuf purchase order data
func (p *PurchaseOrder) loadFromP2PProtobuf(data *purchaseorderpb.PurchaseOrderData) error {
	pdetails, err := documents.FromProtocolPaymentDetails(data.PaymentDetails)
	if err != nil {
		return err
	}

	decs, err := documents.BytesToDecimals(data.TotalAmount)
	if err != nil {
		return err
	}

	dids, err := identity.BytesToDIDs(data.Recipient, data.Sender)
	if err != nil {
		return err
	}

	li, err := fromP2PLineItems(data.LineItems)
	if err != nil {
		return err
	}

	tms, err := timeutils.FromProtoTimestamps(data.DateSent, data.DateUpdated, data.DateCreated, data.DateConfirmed)
	if err != nil {
		return err
	}

	var d Data
	d.Status = data.Status
	d.Number = data.Number
	d.SenderOrderID = data.SenderOrderId
	d.RecipientOrderID = data.RecipientOrderId
	d.RequisitionID = data.RequisitionId
	d.RequesterEmail = data.RequesterEmail
	d.RequesterName = data.RequesterName
	d.ShipToCompanyName = data.ShipToCompanyName
	d.ShipToContactPersonName = data.ShipToContactPersonName
	d.ShipToStreet1 = data.ShipToStreet1
	d.ShipToStreet2 = data.ShipToStreet2
	d.ShipToCity = data.ShipToCity
	d.ShipToZipcode = data.ShipToZipcode
	d.ShipToState = data.ShipToState
	d.ShipToCountry = data.ShipToCountry
	d.PaymentTerms = data.PaymentTerms
	d.Currency = data.Currency
	d.Comment = data.Comment
	d.DateSent = tms[0]
	d.DateUpdated = tms[1]
	d.DateCreated = tms[2]
	d.DateConfirmed = tms[3]
	d.Attachments = documents.FromProtocolAttachments(data.Attachments)
	d.PaymentDetails = pdetails
	d.TotalAmount = decs[0]
	d.Recipient = dids[0]
	d.Sender = dids[1]
	d.LineItems = li
	p.Data = d
	return nil
}

// PackCoreDocument packs the PurchaseOrder into a Core Document
func (p *PurchaseOrder) PackCoreDocument() (cd coredocumentpb.CoreDocument, err error) {
	poData, err := p.createP2PProtobuf()
	if err != nil {
		return cd, err
	}

	data, err := proto.Marshal(poData)
	if err != nil {
		return cd, errors.New("failed to marshal po data: %v", err)
	}

	embedData := &any.Any{
		TypeUrl: p.DocumentType(),
		Value:   data,
	}

	return p.CoreDocument.PackCoreDocument(embedData), nil
}

// UnpackCoreDocument unpacks the core document into PurchaseOrder
func (p *PurchaseOrder) UnpackCoreDocument(cd coredocumentpb.CoreDocument) error {
	if cd.EmbeddedData == nil ||
		cd.EmbeddedData.TypeUrl != p.DocumentType() {
		return errors.New("trying to convert document with incorrect schema")
	}

	poData := new(purchaseorderpb.PurchaseOrderData)
	err := proto.Unmarshal(cd.EmbeddedData.Value, poData)
	if err != nil {
		return err
	}

	err = p.loadFromP2PProtobuf(poData)
	if err != nil {
		return err
	}

	p.CoreDocument, err = documents.NewCoreDocumentFromProtobuf(cd)
	return err

}

// JSON marshals PurchaseOrder into a json bytes
func (p *PurchaseOrder) JSON() ([]byte, error) {
	return p.CoreDocument.MarshalJSON(p)
}

// FromJSON unmarshals the json bytes into PurchaseOrder
func (p *PurchaseOrder) FromJSON(jsonData []byte) error {
	if p.CoreDocument == nil {
		p.CoreDocument = new(documents.CoreDocument)
	}

	return p.CoreDocument.UnmarshalJSON(jsonData, p)
}

// Type gives the PurchaseOrder type
func (p *PurchaseOrder) Type() reflect.Type {
	return reflect.TypeOf(p)
}

// CalculateDataRoot calculates the data root and sets the root to core document
func (p *PurchaseOrder) CalculateDataRoot() ([]byte, error) {
	t, err := p.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("failed to get data tree: %v", err)
	}

	return t.RootHash(), nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (p *PurchaseOrder) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	poProto, err := p.createP2PProtobuf()
	if err != nil {
		return nil, err
	}
	t := p.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	err = t.AddLeavesFromDocument(poProto)
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
func (p *PurchaseOrder) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	tree, err := p.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	return p.CoreDocument.CreateProofs(p.DocumentType(), tree, fields)
}

// DocumentType returns the po document type.
func (*PurchaseOrder) DocumentType() string {
	return documenttypes.PurchaseOrderDataTypeUrl
}

// AddNFT adds NFT to the Purchase Order.
func (p *PurchaseOrder) AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte) error {
	cd, err := p.CoreDocument.AddNFT(grantReadAccess, registry, tokenID)
	if err != nil {
		return err
	}

	p.CoreDocument = cd
	return nil
}

// CalculateSigningRoot returns the signing root of the document.
// Calculates it if not generated yet.
func (p *PurchaseOrder) CalculateSigningRoot() ([]byte, error) {
	dr, err := p.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return p.CoreDocument.CalculateSigningRoot(p.DocumentType(), dr)
}

// CalculateDocumentRoot calculates the document root
func (p *PurchaseOrder) CalculateDocumentRoot() ([]byte, error) {
	dr, err := p.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return p.CoreDocument.CalculateDocumentRoot(p.DocumentType(), dr)
}

// DocumentRootTree creates and returns the document root tree
func (p *PurchaseOrder) DocumentRootTree() (tree *proofs.DocumentTree, err error) {
	dr, err := p.CalculateDataRoot()
	if err != nil {
		return nil, err
	}
	return p.CoreDocument.DocumentRootTree(p.DocumentType(), dr)
}

// CreateNFTProofs creates proofs specific to NFT minting.
func (p *PurchaseOrder) CreateNFTProofs(
	account identity.DID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error) {

	tree, err := p.getDocumentDataTree()
	if err != nil {
		return nil, err
	}

	return p.CoreDocument.CreateNFTProofs(
		p.DocumentType(),
		tree,
		account, registry, tokenID, nftUniqueProof, readAccessProof)
}

// CollaboratorCanUpdate checks if the account can update the document.
func (p *PurchaseOrder) CollaboratorCanUpdate(updated documents.Model, collaborator identity.DID) error {
	newPo, ok := updated.(*PurchaseOrder)
	if !ok {
		return errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("expecting a purchase order but got %T", updated))
	}

	// check the core document changes
	err := p.CoreDocument.CollaboratorCanUpdate(newPo.CoreDocument, collaborator, p.DocumentType())
	if err != nil {
		return err
	}

	// check purchase order specific changes
	oldTree, err := p.getDocumentDataTree()
	if err != nil {
		return err
	}

	newTree, err := newPo.getDocumentDataTree()
	if err != nil {
		return err
	}

	rules := p.CoreDocument.TransitionRulesFor(collaborator)
	cf := documents.GetChangedFields(oldTree, newTree)
	return documents.ValidateTransitions(rules, cf)
}

// AddAttributes adds attributes to the PurchaseOrder model.
func (p *PurchaseOrder) AddAttributes(ca documents.CollaboratorsAccess, prepareNewVersion bool, attrs ...documents.Attribute) error {
	ncd, err := p.CoreDocument.AddAttributes(ca, prepareNewVersion, compactPrefix(), attrs...)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDAttribute, err)
	}

	p.CoreDocument = ncd
	return nil
}

// DeleteAttribute deletes the attribute from the model.
func (p *PurchaseOrder) DeleteAttribute(key documents.AttrKey, prepareNewVersion bool) error {
	ncd, err := p.CoreDocument.DeleteAttribute(key, prepareNewVersion, compactPrefix())
	if err != nil {
		return errors.NewTypedError(documents.ErrCDAttribute, err)
	}

	p.CoreDocument = ncd
	return nil
}

// GetData returns purchase order data
func (p *PurchaseOrder) GetData() interface{} {
	return p.Data
}

// loadData unmarshals json blob to Data.
func (p *PurchaseOrder) loadData(data []byte) error {
	var d Data
	err := json.Unmarshal(data, &d)
	if err != nil {
		return err
	}

	p.Data = d
	return nil
}

// unpackFromCreatePayload unpacks the invoice data from the Payload.
func (p *PurchaseOrder) unpackFromCreatePayload(did identity.DID, payload documents.CreatePayload) error {
	if err := p.loadData(payload.Data); err != nil {
		return errors.NewTypedError(ErrPOInvalidData, err)
	}

	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	cd, err := documents.NewCoreDocument(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDCreate, err)
	}

	p.CoreDocument = cd
	return nil
}

// unpackFromUpdatePayload unpacks the update payload and prepares a new version.
func (p *PurchaseOrder) unpackFromUpdatePayload(old *PurchaseOrder, payload documents.UpdatePayload) error {
	if err := p.loadData(payload.Data); err != nil {
		return errors.NewTypedError(ErrPOInvalidData, err)
	}

	ncd, err := old.CoreDocument.PrepareNewVersion(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return err
	}

	p.CoreDocument = ncd
	return nil
}

// Scheme returns the purchase order scheme.
func (p *PurchaseOrder) Scheme() string {
	return Scheme
}
