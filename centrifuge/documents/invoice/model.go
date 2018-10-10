package invoice

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// InvoiceModel implements the documents.Model keeps track of invoice related fields and state
type InvoiceModel struct {
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
	Recipient   *identity.CentID
	Sender      *identity.CentID
	Payee       *identity.CentID
	Comment     string
	DueDate     *timestamp.Timestamp
	DateCreated *timestamp.Timestamp
	ExtraData   []byte

	InvoiceSalts *invoicepb.InvoiceDataSalts
	CoreDocument *coredocumentpb.CoreDocument
}

// getClientData returns the client data from the invoice model
func (i *InvoiceModel) getClientData() *clientinvoicepb.InvoiceData {
	var recipient string
	if i.Recipient != nil {
		recipient = hexutil.Encode(i.Recipient[:])
	}

	var sender string
	if i.Sender != nil {
		sender = hexutil.Encode(i.Sender[:])
	}

	var payee string
	if i.Payee != nil {
		payee = hexutil.Encode(i.Payee[:])
	}

	var extraData string
	if i.ExtraData != nil {
		extraData = hexutil.Encode(i.ExtraData)
	}

	return &clientinvoicepb.InvoiceData{
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
		ExtraData:        extraData,
	}

}

// createP2PProtobuf returns centrifuge protobuf specific invoiceData
func (i *InvoiceModel) createP2PProtobuf() *invoicepb.InvoiceData {

	var recipient, sender, payee []byte
	if i.Recipient != nil {
		recipient = i.Recipient[:]
	}

	if i.Sender != nil {
		sender = i.Sender[:]
	}

	if i.Payee != nil {
		payee = i.Payee[:]
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
	}

}

// InitInvoiceInput initialize the model based on the received parameters from the rest api call
func (i *InvoiceModel) InitInvoiceInput(payload *clientinvoicepb.InvoiceCreatePayload) error {
	data := payload.Data
	i.InvoiceNumber = data.InvoiceNumber
	i.SenderName = data.SenderName
	i.SenderStreet = data.SenderStreet
	i.SenderCity = data.SenderCity
	i.SenderZipcode = data.SenderZipcode
	i.SenderCountry = data.SenderCountry
	i.RecipientName = data.RecipientName
	i.RecipientStreet = data.RecipientStreet
	i.RecipientCity = data.RecipientCity
	i.RecipientZipcode = data.RecipientZipcode
	i.RecipientCountry = data.RecipientCountry
	i.Currency = data.Currency
	i.GrossAmount = data.GrossAmount
	i.NetAmount = data.NetAmount
	i.TaxAmount = data.TaxAmount
	i.TaxRate = data.TaxRate
	i.Comment = data.Comment
	i.DueDate = data.DueDate
	i.DateCreated = data.DateCreated

	if recipient, err := identity.CentIDFromString(data.Recipient); err == nil {
		i.Recipient = &recipient
	}

	if sender, err := identity.CentIDFromString(data.Sender); err == nil {
		i.Sender = &sender
	}

	if payee, err := identity.CentIDFromString(data.Payee); err == nil {
		i.Payee = &payee
	}

	if data.ExtraData != "" {
		ed, err := hexutil.Decode(data.ExtraData)
		if err != nil {
			return centerrors.Wrap(err, "failed to decode extra data")
		}

		i.ExtraData = ed
	}

	i.CoreDocument = coredocument.New()
	i.CoreDocument.Collaborators = [][]byte{}

	for _, id := range payload.Collaborators {
		cid, err := identity.CentIDFromString(id)
		if err != nil {
			return centerrors.Wrap(err, "failed to decode collaborator")
		}

		cidb := cid.ByteArray()
		i.CoreDocument.Collaborators = append(i.CoreDocument.Collaborators, cidb[:])
	}

	coredocument.FillSalts(i.CoreDocument)

	return nil
}

// loadFromP2PProtobuf  loads the invoice from centrifuge protobuf invoice data
func (i *InvoiceModel) loadFromP2PProtobuf(invoiceData *invoicepb.InvoiceData) {
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

	if recipient, err := identity.ToCentID(invoiceData.Recipient); err == nil {
		i.Recipient = &recipient
	}

	if sender, err := identity.ToCentID(invoiceData.Sender); err == nil {
		i.Sender = &sender
	}

	if payee, err := identity.ToCentID(invoiceData.Payee); err == nil {
		i.Payee = &payee
	}

	i.Comment = invoiceData.Comment
	i.DueDate = invoiceData.DueDate
	i.DateCreated = invoiceData.DateCreated
	i.ExtraData = invoiceData.ExtraData
}

// getInvoiceSalts returns the invoice salts. Initialises if not present
func (i *InvoiceModel) getInvoiceSalts(invoiceData *invoicepb.InvoiceData) *invoicepb.InvoiceDataSalts {
	if i.InvoiceSalts == nil {
		invoiceSalts := &invoicepb.InvoiceDataSalts{}
		proofs.FillSalts(invoiceData, invoiceSalts)
		i.InvoiceSalts = invoiceSalts
	}

	return i.InvoiceSalts
}

// PackCoreDocument packs the InvoiceModel into a Core Document
// If the, InvoiceModel is new, it creates a valid identifiers
func (i *InvoiceModel) PackCoreDocument() (*coredocumentpb.CoreDocument, error) {
	invoiceData := i.createP2PProtobuf()
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

	coreDoc := new(coredocumentpb.CoreDocument)
	proto.Merge(coreDoc, i.CoreDocument)
	coreDoc.EmbeddedData = &invoiceAny
	coreDoc.EmbeddedDataSalts = &invoiceSaltsAny
	return coreDoc, err
}

// UnpackCoreDocument unpacks the core document into InvoiceModel
func (i *InvoiceModel) UnpackCoreDocument(coreDoc *coredocumentpb.CoreDocument) error {
	if coreDoc == nil {
		return centerrors.NilError(coreDoc)
	}

	if coreDoc.EmbeddedData == nil ||
		coreDoc.EmbeddedData.TypeUrl != documenttypes.InvoiceDataTypeUrl ||
		coreDoc.EmbeddedDataSalts == nil ||
		coreDoc.EmbeddedDataSalts.TypeUrl != documenttypes.InvoiceSaltsTypeUrl {
		return fmt.Errorf("trying to convert document with incorrect schema")
	}

	invoiceData := &invoicepb.InvoiceData{}
	err := proto.Unmarshal(coreDoc.EmbeddedData.Value, invoiceData)
	if err != nil {
		return err
	}

	i.loadFromP2PProtobuf(invoiceData)
	invoiceSalts := &invoicepb.InvoiceDataSalts{}
	err = proto.Unmarshal(coreDoc.EmbeddedDataSalts.Value, invoiceSalts)
	if err != nil {
		return err
	}

	i.InvoiceSalts = invoiceSalts
	if i.CoreDocument == nil {
		i.CoreDocument = new(coredocumentpb.CoreDocument)
	}
	proto.Merge(i.CoreDocument, coreDoc)
	i.CoreDocument.EmbeddedDataSalts = nil
	i.CoreDocument.EmbeddedData = nil
	return err
}

// JSON marshals InvoiceModel into a json bytes
func (i *InvoiceModel) JSON() ([]byte, error) {
	return json.Marshal(i)
}

// FromJSON unmarshals the json bytes into InvoiceModel
func (i *InvoiceModel) FromJSON(jsonData []byte) error {
	return json.Unmarshal(jsonData, i)
}

// Type gives the InvoiceModel type
func (i *InvoiceModel) Type() reflect.Type {
	return reflect.TypeOf(i)
}

// calculateDataRoot calculates the data root and sets the root to core document
func (i *InvoiceModel) calculateDataRoot() error {
	t, err := i.getDocumentDataTree()
	if err != nil {
		return fmt.Errorf("calculateDataRoot error %v", err)
	}
	if i.CoreDocument == nil {
		i.CoreDocument = coredocument.New()
	}

	i.CoreDocument.DataRoot = t.RootHash()
	return nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (i *InvoiceModel) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	t := proofs.NewDocumentTree(proofs.TreeOptions{EnableHashSorting: true, Hash: sha256.New()})
	invoiceData := i.createP2PProtobuf()
	err = t.AddLeavesFromDocument(invoiceData, i.getInvoiceSalts(invoiceData))
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
func (i *InvoiceModel) createProofs(fields []string) (coreDoc *coredocumentpb.CoreDocument, proofs []*proofspb.Proof, err error) {
	// There can be failure scenarios where the core doc for the particular document
	// is still not saved with roots in db due to failures during getting signatures.
	coreDoc, err = i.PackCoreDocument()
	if err != nil {
		return nil, nil, fmt.Errorf("createProofs error %v", err)
	}
	dataRootHashes, err := coredocument.GetDataProofHashes(coreDoc)
	if err != nil {
		return coreDoc, nil, fmt.Errorf("createProofs error %v", err)
	}

	tree, err := i.getDocumentDataTree()
	if err != nil {
		return coreDoc, nil, fmt.Errorf("createProofs error %v", err)
	}
	for _, field := range fields {
		proof, err := tree.CreateProof(field)
		if err != nil {
			return coreDoc, nil, fmt.Errorf("createProofs error %v", err)
		}
		proofs = append(proofs, &proof)
	}

	for _, proof := range proofs {
		proof.SortedHashes = append(proof.SortedHashes, dataRootHashes...)
	}

	return
}
