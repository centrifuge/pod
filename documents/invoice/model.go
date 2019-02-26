package invoice

import (
	"encoding/json"
	"reflect"

	"github.com/centrifuge/go-centrifuge/documents"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
)

const prefix string = "invoice"

// tree prefixes specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 1, 0, 0} }

// Invoice implements the documents.Model keeps track of invoice related fields and state
type Invoice struct {
	InvoiceNumber    string // invoice number or reference number
	SenderName       string // name of the sender company
	SenderStreet     string // street and address details of the sender company
	SenderCity       string
	SenderZipcode    string // country ISO code of the sender of this invoice
	SenderCountry    string
	RecipientName    string // name of the recipient company
	RecipientStreet  string
	RecipientCity    string
	RecipientZipcode string
	RecipientCountry string // country ISO code of the recipient of this invoice
	Currency         string // country ISO code of the recipient of this invoice
	GrossAmount      int64  // invoice amount including tax
	NetAmount        int64  // invoice amount excluding tax
	TaxAmount        int64
	TaxRate          int64
	Recipient        *identity.DID
	Sender           *identity.DID
	Payee            *identity.DID
	Comment          string
	DueDate          *timestamp.Timestamp
	DateCreated      *timestamp.Timestamp
	ExtraData        []byte

	InvoiceSalts      *proofs.Salts
	CoreDocumentModel *documents.CoreDocumentModel
}

// getClientData returns the client data from the invoice model
func (i *Invoice) getClientData() *clientinvoicepb.InvoiceData {
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
func (i *Invoice) createP2PProtobuf() *invoicepb.InvoiceData {
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
func (i *Invoice) InitInvoiceInput(payload *clientinvoicepb.InvoiceCreatePayload, self string) error {
	err := i.initInvoiceFromData(payload.Data)
	if err != nil {
		return err
	}

	collaborators := append([]string{self}, payload.Collaborators...)

	i.CoreDocumentModel, err = i.CoreDocumentModel.NewWithCollaborators(collaborators)
	if err != nil {
		return errors.New("failed to init core document: %v", err)
	}

	return nil
}

// initInvoiceFromData initialises invoice from invoiceData
func (i *Invoice) initInvoiceFromData(data *clientinvoicepb.InvoiceData) error {
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

	if data.Recipient != "" {
		if recipient, err := identity.NewDIDFromString(data.Recipient); err == nil {
			i.Recipient = &recipient
		}
	}

	if data.Sender != "" {
		if sender, err := identity.NewDIDFromString(data.Sender); err == nil {
			i.Sender = &sender
		}
	}

	if data.Payee != "" {
		if payee, err := identity.NewDIDFromString(data.Payee); err == nil {
			i.Payee = &payee
		}
	}

	if data.ExtraData != "" {
		ed, err := hexutil.Decode(data.ExtraData)
		if err != nil {
			return errors.NewTypedError(err, errors.New("failed to decode extra data"))
		}

		i.ExtraData = ed
	}

	return nil
}

// loadFromP2PProtobuf  loads the invoice from centrifuge protobuf invoice data
func (i *Invoice) loadFromP2PProtobuf(invoiceData *invoicepb.InvoiceData) {
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

	if invoiceData.Recipient != nil {
		recipient := identity.NewDIDFromBytes(invoiceData.Recipient)
		i.Recipient = &recipient
	}

	if invoiceData.Sender != nil {
		sender := identity.NewDIDFromBytes(invoiceData.Sender)
		i.Sender = &sender
	}

	if invoiceData.Payee != nil {
		payee := identity.NewDIDFromBytes(invoiceData.Payee)
		i.Payee = &payee
	}

	i.Comment = invoiceData.Comment
	i.DueDate = invoiceData.DueDate
	i.DateCreated = invoiceData.DateCreated
	i.ExtraData = invoiceData.ExtraData
}

// getInvoiceSalts returns the invoice salts. Initialises if not present
func (i *Invoice) getInvoiceSalts(invoiceData *invoicepb.InvoiceData) (*proofs.Salts, error) {
	if i.InvoiceSalts == nil {
		invoiceSalts, err := documents.GenerateNewSalts(invoiceData, prefix, compactPrefix())
		if err != nil {
			return nil, errors.New("getInvoiceSalts error %v", err)
		}
		i.InvoiceSalts = invoiceSalts
	}

	return i.InvoiceSalts, nil
}

// ID returns document identifier.
// Note: this is not a unique identifier for each version of the document.
func (i *Invoice) ID() ([]byte, error) {
	coreDocModel, err := i.PackCoreDocument()
	if err != nil {
		return []byte{}, err
	}
	return coreDocModel.Document.DocumentIdentifier, nil
}

// PackCoreDocument packs the Invoice into a Core Document
// If the, Invoice is new, it creates a valid identifiers
func (i *Invoice) PackCoreDocument() (*documents.CoreDocumentModel, error) {
	invoiceData := i.createP2PProtobuf()
	serializedInvoice, err := proto.Marshal(invoiceData)
	if err != nil {
		return nil, errors.NewTypedError(err, errors.New("couldn't serialise InvoiceData"))
	}

	invoiceAny := any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   serializedInvoice,
	}

	invoiceSalts, err := i.getInvoiceSalts(invoiceData)
	if err != nil {
		return nil, errors.NewTypedError(err, errors.New("couldn't get InvoiceSalts"))
	}

	err = i.CoreDocumentModel.PackCoreDocument(&invoiceAny, documents.ConvertToProtoSalts(invoiceSalts))
	if err != nil {
		return nil, err
	}

	return i.CoreDocumentModel, nil
}

// UnpackCoreDocument unpacks the core document into Invoice
func (i *Invoice) UnpackCoreDocument(coreDocModel *documents.CoreDocumentModel) error {
	if coreDocModel == nil {
		return errors.New("coredocmodel is nil %v", coreDocModel)
	}
	if coreDocModel.Document == nil {
		return errors.New("core document provided is nil %v", coreDocModel.Document)
	}

	coreDoc := coreDocModel.Document

	if coreDoc.EmbeddedData == nil ||
		coreDoc.EmbeddedData.TypeUrl != documenttypes.InvoiceDataTypeUrl {
		return errors.New("trying to convert document with incorrect schema")
	}

	invoiceData := &invoicepb.InvoiceData{}
	err := proto.Unmarshal(coreDoc.EmbeddedData.Value, invoiceData)
	if err != nil {
		return err
	}

	i.loadFromP2PProtobuf(invoiceData)

	if coreDoc.EmbeddedDataSalts == nil {
		i.InvoiceSalts, err = i.getInvoiceSalts(invoiceData)
		if err != nil {
			return err
		}
	} else {
		i.InvoiceSalts = documents.ConvertToProofSalts(coreDoc.EmbeddedDataSalts)
	}

	err = i.CoreDocumentModel.UnpackCoreDocument()
	return err

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

// CalculateDataRoot calculates the data root and sets the root to core document
func (i *Invoice) CalculateDataRoot() ([]byte, error) {
	t, err := i.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("calculateDataRoot error %v", err)
	}
	return t.RootHash(), nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (i *Invoice) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	invProto := i.createP2PProtobuf()
	salts, err := i.getInvoiceSalts(invProto)
	if err != nil {
		return nil, err
	}
	t := documents.NewDefaultTreeWithPrefix(salts, prefix, compactPrefix())
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

// CreateProofs generates proofs for given fields
func (i *Invoice) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	// There can be failure scenarios where the core doc for the particular document
	// is still not saved with roots in db due to failures during getting signatures.
	_, err = i.PackCoreDocument()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	tree, err := i.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	proofs, err = i.CoreDocumentModel.CreateProofs(tree, fields)
	return proofs, err
}
