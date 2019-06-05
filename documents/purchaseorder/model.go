package purchaseorder

import (
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
)

const prefix string = "po"

// tree prefixes for specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 2, 0, 0} }

// PurchaseOrder implements the documents.Model keeps track of purchase order related fields and state
type PurchaseOrder struct {
	*documents.CoreDocument
	Status                  string // status of the Purchase Order
	Number                  string // purchase order number or reference number
	SenderOrderID           string
	RecipientOrderID        string
	RequisitionID           string
	RequesterName           string
	RequesterEmail          string
	ShipToCompanyName       string
	ShipToContactPersonName string
	ShipToStreet1           string
	ShipToStreet2           string
	ShipToCity              string
	ShipToZipcode           string
	ShipToState             string
	ShipToCountry           string
	PaymentTerms            string
	Currency                string
	TotalAmount             *documents.Decimal
	Recipient               *identity.DID
	Sender                  *identity.DID
	Comment                 string
	DateSent                *timestamp.Timestamp
	DateConfirmed           *timestamp.Timestamp
	DateUpdated             *timestamp.Timestamp
	DateCreated             *timestamp.Timestamp
	Attachments             []*documents.BinaryAttachment
	LineItems               []*LineItem
	PaymentDetails          []*documents.PaymentDetails
}

// LineItemActivity describes a single line item activity.
type LineItemActivity struct {
	ItemNumber            string
	Status                string
	Quantity              *documents.Decimal
	Amount                *documents.Decimal
	ReferenceDocumentID   string
	ReferenceDocumentItem string
	Date                  *timestamp.Timestamp
}

// TaxItem describes a single Purchase Order tax item.
type TaxItem struct {
	ItemNumber              string
	PurchaseOrderItemNumber string
	TaxAmount               *documents.Decimal
	TaxRate                 *documents.Decimal
	TaxCode                 *documents.Decimal
	TaxBaseAmount           *documents.Decimal
}

// LineItem describes a single LineItem Activity
type LineItem struct {
	Status            string
	ItemNumber        string
	Description       string
	AmountInvoiced    *documents.Decimal
	AmountTotal       *documents.Decimal
	RequisitionNumber string
	RequisitionItem   string
	PartNumber        string
	PricePerUnit      *documents.Decimal
	UnitOfMeasure     *documents.Decimal
	Quantity          *documents.Decimal
	ReceivedQuantity  *documents.Decimal
	DateUpdated       *timestamp.Timestamp
	DateCreated       *timestamp.Timestamp
	RevisionNumber    int64
	Activities        []*LineItemActivity
	TaxItems          []*TaxItem
}

// getClientData returns the client data from the purchaseOrder model
func (p *PurchaseOrder) getClientData() (*clientpurchaseorderpb.PurchaseOrderData, error) {
	decs := documents.DecimalsToStrings(p.TotalAmount)
	dids := identity.DIDsToStrings(p.Recipient, p.Sender)
	attr, err := documents.ToClientAttributes(p.Attributes)
	if err != nil {
		return nil, err
	}

	pd, err := documents.ToClientPaymentDetails(p.PaymentDetails)
	if err != nil {
		return nil, err
	}

	return &clientpurchaseorderpb.PurchaseOrderData{
		Status:                  p.Status,
		Number:                  p.Number,
		SenderOrderId:           p.SenderOrderID,
		TotalAmount:             decs[0],
		Recipient:               dids[0],
		Sender:                  dids[1],
		DateCreated:             p.DateCreated,
		DateUpdated:             p.DateUpdated,
		RequesterName:           p.RequesterName,
		RequesterEmail:          p.RequesterEmail,
		Comment:                 p.Comment,
		Currency:                p.Currency,
		ShipToCountry:           p.ShipToCountry,
		ShipToState:             p.ShipToState,
		ShipToZipcode:           p.ShipToZipcode,
		ShipToCity:              p.ShipToCity,
		ShipToStreet1:           p.ShipToStreet1,
		ShipToStreet2:           p.ShipToStreet2,
		ShipToContactPersonName: p.ShipToContactPersonName,
		ShipToCompanyName:       p.ShipToCompanyName,
		DateConfirmed:           p.DateConfirmed,
		DateSent:                p.DateSent,
		PaymentTerms:            p.PaymentTerms,
		RecipientOrderId:        p.RecipientOrderID,
		RequisitionId:           p.RequisitionID,
		PaymentDetails:          pd,
		Attachments:             documents.ToClientAttachments(p.Attachments),
		LineItems:               toClientLineItems(p.LineItems),
		Attributes:              attr,
	}, nil
}

// createP2PProtobuf returns centrifuge protobuf specific purchaseOrderData
func (p *PurchaseOrder) createP2PProtobuf() (*purchaseorderpb.PurchaseOrderData, error) {
	decs, err := documents.DecimalsToBytes(p.TotalAmount)
	if err != nil {
		return nil, err
	}

	pd, err := documents.ToProtocolPaymentDetails(p.PaymentDetails)
	if err != nil {
		return nil, err
	}

	li, err := toP2PLineItems(p.LineItems)
	if err != nil {
		return nil, err
	}

	dids := identity.DIDsToBytes(p.Recipient, p.Sender)
	return &purchaseorderpb.PurchaseOrderData{
		Status:                  p.Status,
		Number:                  p.Number,
		SenderOrderId:           p.SenderOrderID,
		TotalAmount:             decs[0],
		Recipient:               dids[0],
		Sender:                  dids[1],
		DateCreated:             p.DateCreated,
		DateUpdated:             p.DateUpdated,
		RequesterName:           p.RequesterName,
		RequesterEmail:          p.RequesterEmail,
		Comment:                 p.Comment,
		Currency:                p.Currency,
		ShipToCountry:           p.ShipToCountry,
		ShipToState:             p.ShipToState,
		ShipToZipcode:           p.ShipToZipcode,
		ShipToCity:              p.ShipToCity,
		ShipToStreet1:           p.ShipToStreet1,
		ShipToStreet2:           p.ShipToStreet2,
		ShipToContactPersonName: p.ShipToContactPersonName,
		ShipToCompanyName:       p.ShipToCompanyName,
		DateConfirmed:           p.DateConfirmed,
		DateSent:                p.DateSent,
		PaymentTerms:            p.PaymentTerms,
		RecipientOrderId:        p.RecipientOrderID,
		RequisitionId:           p.RequisitionID,
		PaymentDetails:          pd,
		Attachments:             documents.ToProtocolAttachments(p.Attachments),
		LineItems:               li,
	}, nil

}

// InitPurchaseOrderInput initialize the model based on the received parameters from the rest api call
func (p *PurchaseOrder) InitPurchaseOrderInput(payload *clientpurchaseorderpb.PurchaseOrderCreatePayload, self identity.DID) error {
	err := p.initPurchaseOrderFromData(payload.Data)
	if err != nil {
		return err
	}

	cs, err := documents.FromClientCollaboratorAccess(payload.ReadAccess, payload.WriteAccess)
	if err != nil {
		return err
	}
	cs.ReadWriteCollaborators = append(cs.ReadWriteCollaborators, self)

	attrs, err := documents.FromClientAttributes(payload.Data.Attributes)
	if err != nil {
		return err
	}

	cd, err := documents.NewCoreDocument(compactPrefix(), cs, attrs)
	if err != nil {
		return errors.New("failed to init core document: %v", err)
	}

	p.CoreDocument = cd
	return nil
}

// initPurchaseOrderFromData initialises purchase order from purchaseOrderData
func (p *PurchaseOrder) initPurchaseOrderFromData(data *clientpurchaseorderpb.PurchaseOrderData) error {
	atts, err := documents.FromClientAttachments(data.Attachments)
	if err != nil {
		return err
	}

	pdetails, err := documents.FromClientPaymentDetails(data.PaymentDetails)
	if err != nil {
		return err
	}

	decs, err := documents.StringsToDecimals(data.TotalAmount)
	if err != nil {
		return err
	}

	dids, err := identity.StringsToDIDs(data.Recipient, data.Sender)
	if err != nil {
		return err
	}

	li, err := fromClientLineItems(data.LineItems)
	if err != nil {
		return err
	}

	p.Status = data.Status
	p.Number = data.Number
	p.SenderOrderID = data.SenderOrderId
	p.RecipientOrderID = data.RecipientOrderId
	p.RequisitionID = data.RequisitionId
	p.RequesterEmail = data.RequesterEmail
	p.RequesterName = data.RequesterName
	p.ShipToCompanyName = data.ShipToCompanyName
	p.ShipToContactPersonName = data.ShipToContactPersonName
	p.ShipToStreet1 = data.ShipToStreet1
	p.ShipToStreet2 = data.ShipToStreet2
	p.ShipToCity = data.ShipToCity
	p.ShipToZipcode = data.ShipToZipcode
	p.ShipToState = data.ShipToState
	p.ShipToCountry = data.ShipToCountry
	p.PaymentTerms = data.PaymentTerms
	p.Currency = data.Currency
	p.Comment = data.Comment
	p.DateSent = data.DateSent
	p.DateUpdated = data.DateUpdated
	p.DateCreated = data.DateCreated
	p.DateConfirmed = data.DateConfirmed
	p.Attachments = atts
	p.PaymentDetails = pdetails
	p.TotalAmount = decs[0]
	p.Recipient = dids[0]
	p.Sender = dids[1]
	p.LineItems = li
	return nil
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

	p.Status = data.Status
	p.Number = data.Number
	p.SenderOrderID = data.SenderOrderId
	p.RecipientOrderID = data.RecipientOrderId
	p.RequisitionID = data.RequisitionId
	p.RequesterEmail = data.RequesterEmail
	p.RequesterName = data.RequesterName
	p.ShipToCompanyName = data.ShipToCompanyName
	p.ShipToContactPersonName = data.ShipToContactPersonName
	p.ShipToStreet1 = data.ShipToStreet1
	p.ShipToStreet2 = data.ShipToStreet2
	p.ShipToCity = data.ShipToCity
	p.ShipToZipcode = data.ShipToZipcode
	p.ShipToState = data.ShipToState
	p.ShipToCountry = data.ShipToCountry
	p.PaymentTerms = data.PaymentTerms
	p.Currency = data.Currency
	p.Comment = data.Comment
	p.DateSent = data.DateSent
	p.DateUpdated = data.DateUpdated
	p.DateCreated = data.DateCreated
	p.DateConfirmed = data.DateConfirmed
	p.Attachments = documents.FromProtocolAttachments(data.Attachments)
	p.PaymentDetails = pdetails
	p.TotalAmount = decs[0]
	p.Recipient = dids[0]
	p.Sender = dids[1]
	p.LineItems = li
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
	t, err := p.getDataTree()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}

	return t.RootHash(), nil
}

func (p *PurchaseOrder) getDataLeaves() ([]proofs.LeafNode, error) {
	t, err := p.getRawDataTree()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return t.GetLeaves(), nil
}

func (p *PurchaseOrder) getRawDataTree() (*proofs.DocumentTree, error) {
	poProto, err := p.createP2PProtobuf()
	if err != nil {
		return nil, err
	}
	if p.CoreDocument == nil {
		return nil, errors.New("getDataTree error CoreDocument not set")
	}
	t, err := p.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	err = t.AddLeavesFromDocument(poProto)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return t, nil
}

// getDataTree creates precise-proofs data tree for the model
func (p *PurchaseOrder) getDataTree() (*proofs.DocumentTree, error) {
	tree, err := p.getRawDataTree()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	err = tree.Generate()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}

	return tree, nil
}

// CreateProofs generates proofs for given fields.
func (p *PurchaseOrder) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	dataLeaves, err := p.getDataLeaves()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	return p.CoreDocument.CreateProofs(p.DocumentType(), dataLeaves, fields)
}

// DocumentType returns the po document type.
func (*PurchaseOrder) DocumentType() string {
	return documenttypes.PurchaseOrderDataTypeUrl
}

// PrepareNewVersion prepares new version from the old invoice.
func (p *PurchaseOrder) PrepareNewVersion(old documents.Model, data *clientpurchaseorderpb.PurchaseOrderData, collaborators documents.CollaboratorsAccess) error {
	err := p.initPurchaseOrderFromData(data)
	if err != nil {
		return err
	}

	attrs, err := documents.FromClientAttributes(data.Attributes)
	if err != nil {
		return err
	}

	oldCD := old.(*PurchaseOrder).CoreDocument
	p.CoreDocument, err = oldCD.PrepareNewVersion(compactPrefix(), collaborators, attrs)
	if err != nil {
		return err
	}

	return nil
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

// CalculateDocumentDataRoot returns the document data root of the document.
// Calculates it if not generated yet.
func (p *PurchaseOrder) CalculateDocumentDataRoot() ([]byte, error) {
	dataLeaves, err := p.getDataLeaves()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return p.CoreDocument.CalculateDocumentDataRoot(p.DocumentType(), dataLeaves)
}

// CalculateDocumentRoot calculates the document root
func (p *PurchaseOrder) CalculateDocumentRoot() ([]byte, error) {
	dataLeaves, err := p.getDataLeaves()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return p.CoreDocument.CalculateDocumentRoot(p.DocumentType(), dataLeaves)
}

// DocumentRootTree creates and returns the document root tree
func (p *PurchaseOrder) DocumentRootTree() (tree *proofs.DocumentTree, err error) {
	dataLeaves, err := p.getDataLeaves()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return p.CoreDocument.DocumentRootTree(p.DocumentType(), dataLeaves)
}

// CreateNFTProofs creates proofs specific to NFT minting.
func (p *PurchaseOrder) CreateNFTProofs(
	account identity.DID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error) {

	dataLeaves, err := p.getDataLeaves()
	if err != nil {
		return nil, err
	}

	return p.CoreDocument.CreateNFTProofs(
		p.DocumentType(),
		dataLeaves,
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
	oldTree, err := p.getDataTree()
	if err != nil {
		return err
	}

	newTree, err := newPo.getDataTree()
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
