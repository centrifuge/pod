package invoice

import (
	"crypto/sha256"
	"encoding/hex"
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
	Recipient   identity.CentID
	Sender      identity.CentID
	Payee       identity.CentID
	Comment     string
	DueDate     *timestamp.Timestamp
	DateCreated *timestamp.Timestamp
	ExtraData   []byte

	// This will move to core document once its implemented
	Collaborators []identity.CentID

	InvoiceSalts *invoicepb.InvoiceDataSalts
	CoreDocument *coredocumentpb.CoreDocument
}

// getClientData returns the client data from the invoice model
func (i *InvoiceModel) getClientData() (*clientinvoicepb.InvoiceData, error) {
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
		Recipient:        hex.EncodeToString(recipient),
		Sender:           hex.EncodeToString(sender),
		Payee:            hex.EncodeToString(payee),
		Comment:          i.Comment,
		DueDate:          i.DueDate,
		DateCreated:      i.DateCreated,
		ExtraData:        hex.EncodeToString(i.ExtraData),
	}, nil

}

// createP2PProtobuf returns centrifuge protobuf specific invoiceData
func (i *InvoiceModel) createP2PProtobuf() (*invoicepb.InvoiceData, error) {

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

	var err error
	i.Recipient, err = identity.CentIDFromString(data.Recipient)
	if err != nil {
		return centerrors.Wrap(err, "failed to decode recipient")
	}

	i.Sender, err = identity.CentIDFromString(data.Sender)
	if err != nil {
		return centerrors.Wrap(err, "failed to decode sender")
	}

	i.Payee, err = identity.CentIDFromString(data.Payee)
	if err != nil {
		return centerrors.Wrap(err, "failed to decode payee")
	}

	i.ExtraData, err = hex.DecodeString(data.ExtraData)
	if err != nil {
		return centerrors.Wrap(err, "failed to decode extra data")
	}

	for _, id := range payload.Collaborators {
		cid, err := identity.CentIDFromString(id)
		if err != nil {
			return centerrors.Wrap(err, "failed to decode collaborator")
		}

		i.Collaborators = append(i.Collaborators, cid)
	}

	return nil
}

// loadFromP2PProtobuf  loads the invoice from centrifuge protobuf invoice data
func (i *InvoiceModel) loadFromP2PProtobuf(invoiceData *invoicepb.InvoiceData) error {
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
// TODO: once coredoc has collaborators, take the collaborators from the model
func (i *InvoiceModel) PackCoreDocument() (*coredocumentpb.CoreDocument, error) {
	if i.CoreDocument == nil {
		// this is the new invoice create. so create identifiers
		i.CoreDocument = coredocument.New()
	}

	invoiceData, err := i.createP2PProtobuf()
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

	err = i.loadFromP2PProtobuf(invoiceData)
	if err != nil {
		return err
	}

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
	pb, err := i.createP2PProtobuf()
	if err != nil {
		return fmt.Errorf("failed to create protobuf: %v", err)
	}

	t := proofs.NewDocumentTree(proofs.TreeOptions{EnableHashSorting: true, Hash: sha256.New()})
	if err = t.AddLeavesFromDocument(pb, i.getInvoiceSalts(pb)); err != nil {
		return fmt.Errorf("failed to add leaves from invoice: %v", err)
	}

	if err = t.Generate(); err != nil {
		return fmt.Errorf("failed to generate merkle root: %v", err)
	}

	if i.CoreDocument == nil {
		i.CoreDocument = coredocument.New()
	}

	i.CoreDocument.DataRoot = t.RootHash()
	return nil
}
