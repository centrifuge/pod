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
	"github.com/ethereum/go-ethereum/common/hexutil"
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

	InvoiceNumber    string // invoice number or reference number
	InvoiceStatus    string // invoice status
	SenderName       string // name of the sender company
	SenderStreet     string // street and address details of the sender company
	SenderCity       string
	SenderZipcode    string // country ISO code of the sender of this invoice
	SenderCountry    string
	RecipientName    string // name of the recipient company
	RecipientStreet  string
	RecipientCity    string
	RecipientZipcode string
	RecipientCountry string             // country ISO code of the recipient of this invoice
	Currency         string             // country ISO code of the recipient of this invoice
	GrossAmount      *documents.Decimal // invoice amount including tax
	NetAmount        *documents.Decimal // invoice amount excluding tax
	TaxAmount        *documents.Decimal
	TaxRate          *documents.Decimal
	Recipient        *identity.DID
	Sender           *identity.DID
	Payee            *identity.DID
	Comment          string
	DueDate          *timestamp.Timestamp
	DateCreated      *timestamp.Timestamp
	ExtraData        []byte
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

	decs := documents.DecimalsToStrings(i.GrossAmount, i.NetAmount, i.TaxAmount, i.TaxRate)
	return &clientinvoicepb.InvoiceData{
		InvoiceNumber:    i.InvoiceNumber,
		InvoiceStatus:    i.InvoiceStatus,
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
		GrossAmount:      decs[0],
		NetAmount:        decs[1],
		TaxAmount:        decs[2],
		TaxRate:          decs[3],
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
func (i *Invoice) createP2PProtobuf() (data *invoicepb.InvoiceData, err error) {
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

	decs, err := documents.DecimalsToBytes(i.GrossAmount, i.NetAmount, i.TaxAmount, i.TaxRate)
	if err != nil {
		return nil, err
	}

	return &invoicepb.InvoiceData{
		InvoiceNumber:    i.InvoiceNumber,
		InvoiceStatus:    i.InvoiceStatus,
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
		GrossAmount:      decs[0],
		NetAmount:        decs[1],
		TaxAmount:        decs[2],
		TaxRate:          decs[3],
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
func (i *Invoice) InitInvoiceInput(payload *clientinvoicepb.InvoiceCreatePayload, self string) error {
	err := i.initInvoiceFromData(payload.Data)
	if err != nil {
		return err
	}

	collaborators := append([]string{self}, payload.Collaborators...)
	cd, err := documents.NewCoreDocumentWithCollaborators(collaborators)
	if err != nil {
		return errors.New("failed to init core document: %v", err)
	}

	i.CoreDocument = cd
	return nil
}

// initInvoiceFromData initialises invoice from invoiceData
func (i *Invoice) initInvoiceFromData(data *clientinvoicepb.InvoiceData) error {
	i.InvoiceNumber = data.InvoiceNumber
	i.InvoiceStatus = data.InvoiceStatus
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
	i.Comment = data.Comment
	i.DueDate = data.DueDate
	i.DateCreated = data.DateCreated

	decs, err := documents.StringsToDecimals(data.GrossAmount, data.NetAmount, data.TaxAmount, data.TaxRate)
	if err != nil {
		return err
	}

	i.GrossAmount = decs[0]
	i.NetAmount = decs[1]
	i.TaxAmount = decs[2]
	i.TaxRate = decs[3]

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
func (i *Invoice) loadFromP2PProtobuf(data *invoicepb.InvoiceData) error {
	i.InvoiceNumber = data.InvoiceNumber
	i.InvoiceStatus = data.InvoiceStatus
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

	decs, err := documents.BytesToDecimals(data.GrossAmount, data.NetAmount, data.TaxAmount, data.TaxRate)
	if err != nil {
		return err
	}

	i.GrossAmount = decs[0]
	i.NetAmount = decs[1]
	i.TaxAmount = decs[2]
	i.TaxRate = decs[3]

	if data.Recipient != nil {
		recipient := identity.NewDIDFromBytes(data.Recipient)
		i.Recipient = &recipient
	}

	if data.Sender != nil {
		sender := identity.NewDIDFromBytes(data.Sender)
		i.Sender = &sender
	}

	if data.Payee != nil {
		payee := identity.NewDIDFromBytes(data.Payee)
		i.Payee = &payee
	}

	i.Comment = data.Comment
	i.DueDate = data.DueDate
	i.DateCreated = data.DateCreated
	i.ExtraData = data.ExtraData
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

	i.CoreDocument.SetDataModified(false)
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
	return i.CoreDocument.CreateNFTProofs(
		i.DocumentType(),
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
