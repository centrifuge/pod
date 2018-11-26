package purchaseorder

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/identity"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
)

const prefix string = "po"

// PurchaseOrder implements the documents.Model keeps track of purchase order related fields and state
type PurchaseOrder struct {
	Status            string // status of the Purchase Order
	PoNumber          string // purchase order number or reference number
	OrderName         string // name of the ordering company
	OrderStreet       string // street and address details of the ordering company
	OrderCity         string
	OrderZipcode      string
	OrderCountry      string // country ISO code of the ordering company of this purchase order
	RecipientName     string // name of the recipient company
	RecipientStreet   string
	RecipientCity     string
	RecipientZipcode  string
	RecipientCountry  string // country ISO code of the receipient of this purchase order
	Currency          string // ISO currency code
	OrderAmount       int64  // ordering gross amount including tax
	NetAmount         int64  // invoice amount excluding tax
	TaxAmount         int64
	TaxRate           int64
	Recipient         *identity.CentID
	Order             []byte
	OrderContact      string
	Comment           string
	DeliveryDate      *timestamp.Timestamp // requested delivery date
	DateCreated       *timestamp.Timestamp // purchase order date
	ExtraData         []byte
	PurchaseOrderSalt *purchaseorderpb.PurchaseOrderDataSalts
	CoreDocument      *coredocumentpb.CoreDocument
}

// ID returns the DocumentIdentifier for this document
// Note: this is not same as VersionIdentifier
func (p *PurchaseOrder) ID() ([]byte, error) {
	coreDoc, err := p.PackCoreDocument()
	if err != nil {
		return []byte{}, err
	}
	return coreDoc.DocumentIdentifier, nil
}

// getClientData returns the client data from the purchaseOrder model
func (p *PurchaseOrder) getClientData() *clientpurchaseorderpb.PurchaseOrderData {
	var recipient string
	if p.Recipient != nil {
		recipient = hexutil.Encode(p.Recipient[:])
	}

	var order string
	if p.Order != nil {
		order = hexutil.Encode(p.Order)
	}

	var extraData string
	if p.ExtraData != nil {
		extraData = hexutil.Encode(p.ExtraData)
	}

	return &clientpurchaseorderpb.PurchaseOrderData{
		PoStatus:         p.Status,
		PoNumber:         p.PoNumber,
		OrderName:        p.OrderName,
		OrderStreet:      p.OrderStreet,
		OrderCity:        p.OrderCity,
		OrderZipcode:     p.OrderZipcode,
		OrderCountry:     p.OrderCountry,
		RecipientName:    p.RecipientName,
		RecipientStreet:  p.RecipientStreet,
		RecipientCity:    p.RecipientCity,
		RecipientZipcode: p.RecipientZipcode,
		RecipientCountry: p.RecipientCountry,
		Currency:         p.Currency,
		OrderAmount:      p.OrderAmount,
		NetAmount:        p.NetAmount,
		TaxAmount:        p.TaxAmount,
		TaxRate:          p.TaxRate,
		Recipient:        recipient,
		Order:            order,
		OrderContact:     p.OrderContact,
		Comment:          p.Comment,
		DeliveryDate:     p.DeliveryDate,
		DateCreated:      p.DateCreated,
		ExtraData:        extraData,
	}

}

// createP2PProtobuf returns centrifuge protobuf specific purchaseOrderData
func (p *PurchaseOrder) createP2PProtobuf() *purchaseorderpb.PurchaseOrderData {
	var recipient []byte
	if p.Recipient != nil {
		recipient = p.Recipient[:]
	}

	return &purchaseorderpb.PurchaseOrderData{
		PoStatus:         p.Status,
		PoNumber:         p.PoNumber,
		OrderName:        p.OrderName,
		OrderStreet:      p.OrderStreet,
		OrderCity:        p.OrderCity,
		OrderZipcode:     p.OrderZipcode,
		OrderCountry:     p.OrderCountry,
		RecipientName:    p.RecipientName,
		RecipientStreet:  p.RecipientStreet,
		RecipientCity:    p.RecipientCity,
		RecipientZipcode: p.RecipientZipcode,
		RecipientCountry: p.RecipientCountry,
		Currency:         p.Currency,
		OrderAmount:      p.OrderAmount,
		NetAmount:        p.NetAmount,
		TaxAmount:        p.TaxAmount,
		TaxRate:          p.TaxRate,
		Recipient:        recipient,
		Order:            p.Order,
		OrderContact:     p.OrderContact,
		Comment:          p.Comment,
		DeliveryDate:     p.DeliveryDate,
		DateCreated:      p.DateCreated,
		ExtraData:        p.ExtraData,
	}

}

// InitPurchaseOrderInput initialize the model based on the received parameters from the rest api call
func (p *PurchaseOrder) InitPurchaseOrderInput(payload *clientpurchaseorderpb.PurchaseOrderCreatePayload, contextHeader *header.ContextHeader) error {
	err := p.initPurchaseOrderFromData(payload.Data)
	if err != nil {
		return err
	}

	collaborators := append([]string{contextHeader.Self().ID.String()}, payload.Collaborators...)
	p.CoreDocument, err = coredocument.NewWithCollaborators(collaborators)
	if err != nil {
		return fmt.Errorf("failed to init core document: %v", err)
	}

	return nil
}

// initPurchaseOrderFromData initialises purchase order from purchaseOrderData
func (p *PurchaseOrder) initPurchaseOrderFromData(data *clientpurchaseorderpb.PurchaseOrderData) error {
	p.Status = data.PoStatus
	p.PoNumber = data.PoNumber
	p.OrderName = data.OrderName
	p.OrderStreet = data.OrderStreet
	p.OrderCity = data.OrderCity
	p.OrderZipcode = data.OrderZipcode
	p.OrderCountry = data.OrderCountry
	p.RecipientName = data.RecipientName
	p.RecipientStreet = data.RecipientStreet
	p.RecipientCity = data.RecipientCity
	p.RecipientZipcode = data.RecipientZipcode
	p.RecipientCountry = data.RecipientCountry
	p.Currency = data.Currency
	p.OrderAmount = data.OrderAmount
	p.NetAmount = data.NetAmount
	p.TaxAmount = data.TaxAmount
	p.TaxRate = data.TaxRate

	if data.Order != "" {
		order, err := hexutil.Decode(data.Order)
		if err != nil {
			return centerrors.Wrap(err, "failed to decode order")
		}

		p.Order = order
	}

	p.OrderContact = data.OrderContact
	p.Comment = data.Comment
	p.DeliveryDate = data.DeliveryDate
	p.DateCreated = data.DateCreated

	if recipient, err := identity.CentIDFromString(data.Recipient); err == nil {
		p.Recipient = &recipient
	}

	if data.ExtraData != "" {
		ed, err := hexutil.Decode(data.ExtraData)
		if err != nil {
			return centerrors.Wrap(err, "failed to decode extra data")
		}

		p.ExtraData = ed
	}

	return nil
}

// loadFromP2PProtobuf loads the purcase order from centrifuge protobuf purchase order data
func (p *PurchaseOrder) loadFromP2PProtobuf(data *purchaseorderpb.PurchaseOrderData) {
	p.Status = data.PoStatus
	p.PoNumber = data.PoNumber
	p.OrderName = data.OrderName
	p.OrderStreet = data.OrderStreet
	p.OrderCity = data.OrderCity
	p.OrderZipcode = data.OrderZipcode
	p.OrderCountry = data.OrderCountry
	p.RecipientName = data.RecipientName
	p.RecipientStreet = data.RecipientStreet
	p.RecipientCity = data.RecipientCity
	p.RecipientZipcode = data.RecipientZipcode
	p.RecipientCountry = data.RecipientCountry
	p.Currency = data.Currency
	p.OrderAmount = data.OrderAmount
	p.NetAmount = data.NetAmount
	p.TaxAmount = data.TaxAmount
	p.TaxRate = data.TaxRate
	p.Order = data.Order
	p.OrderContact = data.OrderContact
	p.Comment = data.Comment
	p.DeliveryDate = data.DeliveryDate
	p.DateCreated = data.DateCreated
	p.ExtraData = data.ExtraData

	if recipient, err := identity.ToCentID(data.Recipient); err == nil {
		p.Recipient = &recipient
	}
}

// getPurchaseOrderSalts returns the purchase oder salts. Initialises if not present
func (p *PurchaseOrder) getPurchaseOrderSalts(purchaseOrderData *purchaseorderpb.PurchaseOrderData) *purchaseorderpb.PurchaseOrderDataSalts {
	if p.PurchaseOrderSalt == nil {
		purchaseOrderSalt := &purchaseorderpb.PurchaseOrderDataSalts{}
		proofs.FillSalts(purchaseOrderData, purchaseOrderSalt)
		p.PurchaseOrderSalt = purchaseOrderSalt
	}

	return p.PurchaseOrderSalt
}

// PackCoreDocument packs the PurchaseOrder into a Core Document
// If the, PurchaseOrder is new, it creates a valid identifiers
func (p *PurchaseOrder) PackCoreDocument() (*coredocumentpb.CoreDocument, error) {
	poData := p.createP2PProtobuf()
	poSerialized, err := proto.Marshal(poData)
	if err != nil {
		return nil, centerrors.Wrap(err, "couldn't serialise PurchaseOrderData")
	}

	poAny := any.Any{
		TypeUrl: documenttypes.PurchaseOrderDataTypeUrl,
		Value:   poSerialized,
	}

	poSalt := p.getPurchaseOrderSalts(poData)

	serializedSalts, err := proto.Marshal(poSalt)
	if err != nil {
		return nil, centerrors.Wrap(err, "couldn't serialise PurchaseOrderSalt")
	}

	poSaltsAny := any.Any{
		TypeUrl: documenttypes.PurchaseOrderSaltsTypeUrl,
		Value:   serializedSalts,
	}

	coreDoc := new(coredocumentpb.CoreDocument)
	proto.Merge(coreDoc, p.CoreDocument)
	coreDoc.EmbeddedData = &poAny
	coreDoc.EmbeddedDataSalts = &poSaltsAny
	return coreDoc, err
}

// UnpackCoreDocument unpacks the core document into PurchaseOrder
func (p *PurchaseOrder) UnpackCoreDocument(coreDoc *coredocumentpb.CoreDocument) error {
	if coreDoc == nil {
		return centerrors.NilError(coreDoc)
	}

	if coreDoc.EmbeddedData == nil ||
		coreDoc.EmbeddedData.TypeUrl != documenttypes.PurchaseOrderDataTypeUrl ||
		coreDoc.EmbeddedDataSalts == nil ||
		coreDoc.EmbeddedDataSalts.TypeUrl != documenttypes.PurchaseOrderSaltsTypeUrl {
		return fmt.Errorf("trying to convert document with incorrect schema")
	}

	poData := &purchaseorderpb.PurchaseOrderData{}
	err := proto.Unmarshal(coreDoc.EmbeddedData.Value, poData)
	if err != nil {
		return err
	}

	p.loadFromP2PProtobuf(poData)
	poSalt := &purchaseorderpb.PurchaseOrderDataSalts{}
	err = proto.Unmarshal(coreDoc.EmbeddedDataSalts.Value, poSalt)
	if err != nil {
		return err
	}

	p.PurchaseOrderSalt = poSalt

	p.CoreDocument = new(coredocumentpb.CoreDocument)
	proto.Merge(p.CoreDocument, coreDoc)
	p.CoreDocument.EmbeddedDataSalts = nil
	p.CoreDocument.EmbeddedData = nil
	return err
}

// JSON marshals PurchaseOrder into a json bytes
func (p *PurchaseOrder) JSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON unmarshals the json bytes into PurchaseOrder
func (p *PurchaseOrder) FromJSON(jsonData []byte) error {
	return json.Unmarshal(jsonData, p)
}

// Type gives the PurchaseOrder type
func (p *PurchaseOrder) Type() reflect.Type {
	return reflect.TypeOf(p)
}

// calculateDataRoot calculates the data root and sets the root to core document
func (p *PurchaseOrder) calculateDataRoot() error {
	t, err := p.getDocumentDataTree()
	if err != nil {
		return fmt.Errorf("calculateDataRoot error %v", err)
	}
	p.CoreDocument.DataRoot = t.RootHash()
	return nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (p *PurchaseOrder) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	t := proofs.NewDocumentTree(proofs.TreeOptions{EnableHashSorting: true, Hash: sha256.New(), ParentPrefix: prefix})
	poData := p.createP2PProtobuf()
	err = t.AddLeavesFromDocument(poData, p.getPurchaseOrderSalts(poData))
	if err != nil {
		return nil, fmt.Errorf("getDocumentDataTree error %v", err)
	}
	err = t.Generate()
	if err != nil {
		return nil, fmt.Errorf("getDocumentDataTree error %v", err)
	}
	return &t, nil
}

// CreateProofs generates proofs for given fields
func (p *PurchaseOrder) createProofs(fields []string) (coreDoc *coredocumentpb.CoreDocument, proofs []*proofspb.Proof, err error) {
	// There can be failure scenarios where the core doc for the particular document
	// is still not saved with roots in db due to failures during getting signatures.
	coreDoc, err = p.PackCoreDocument()
	if err != nil {
		return nil, nil, fmt.Errorf("createProofs error %v", err)
	}

	tree, err := p.getDocumentDataTree()
	if err != nil {
		return coreDoc, nil, fmt.Errorf("createProofs error %v", err)
	}

	proofs, err = coredocument.CreateProofs(tree, coreDoc, fields)
	return coreDoc, proofs, err
}
