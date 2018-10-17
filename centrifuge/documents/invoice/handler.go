package invoice

import (
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/p2p"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	legacyinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/legacy/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("invoice-api")

// grpcHandler handles all the invoice document related actions
// anchoring, sending, finding stored invoice document
type grpcHandler struct {
	legacyRepo       storage.LegacyRepository
	coreDocProcessor coredocumentprocessor.Processor
	service          Service
}

// LegacyGRPCHandler returns an handler that implements InvoiceDocumentServiceServer
// Deprecated: use GRPCHandler()
func LegacyGRPCHandler() legacyinvoicepb.InvoiceDocumentServiceServer {
	return &grpcHandler{
		legacyRepo:       GetLegacyRepository(),
		coreDocProcessor: coredocumentprocessor.DefaultProcessor(identity.IDService, p2p.NewP2PClient(), anchors.GetAnchorRepository()),
	}
}

// GRPCHandler returns an implementation of invoice.DocumentServiceServer
func GRPCHandler(service Service) clientinvoicepb.DocumentServiceServer {
	return &grpcHandler{
		service: service,
	}
}

// anchorInvoiceDocument anchors the given invoice document and returns the anchored document
// Deprecated
func (h *grpcHandler) anchorInvoiceDocument(ctx context.Context, doc *invoicepb.InvoiceDocument, collaborators [][]byte) (*invoicepb.InvoiceDocument, error) {
	inv, err := New(doc, collaborators)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	// TODO review this create, do we need to refactor this because Send method also calls this?
	err = h.legacyRepo.Create(inv.Document.CoreDocument.DocumentIdentifier, inv.Document)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("error saving invoice: %v", err))
	}

	coreDoc, err := inv.ConvertToCoreDocument()
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	err = h.coreDocProcessor.Anchor(ctx, coreDoc, nil)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	// we do not need this conversion again
	newInvoice, err := NewFromCoreDocument(coreDoc)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	return newInvoice.Document, nil
}

// CreateInvoiceProof creates proofs for a list of fields
// Deprecated
func (h *grpcHandler) CreateInvoiceProof(ctx context.Context, createInvoiceProofEnvelope *legacyinvoicepb.CreateInvoiceProofEnvelope) (*legacyinvoicepb.InvoiceProof, error) {
	invDoc := new(invoicepb.InvoiceDocument)
	err := h.legacyRepo.GetByID(createInvoiceProofEnvelope.DocumentIdentifier, invDoc)
	if err != nil {
		return nil, err
	}

	inv, err := Wrap(invDoc)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	// Set EmbeddedData
	inv.Document.CoreDocument.EmbeddedData = &any.Any{TypeUrl: documenttypes.InvoiceDataTypeUrl, Value: []byte{}}

	proofs, err := inv.CreateProofs(createInvoiceProofEnvelope.Fields)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	return &legacyinvoicepb.InvoiceProof{FieldProofs: proofs, DocumentIdentifier: inv.Document.CoreDocument.DocumentIdentifier}, nil

}

// AnchorInvoiceDocument anchors the given invoice document and returns the anchor details
// Deprecated
func (h *grpcHandler) AnchorInvoiceDocument(ctx context.Context, anchorInvoiceEnvelope *legacyinvoicepb.AnchorInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	anchoredInvDoc, err := h.anchorInvoiceDocument(ctx, anchorInvoiceEnvelope.Document, nil)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to anchor: %v", err))
	}

	// Updating invoice with autogenerated fields after anchoring
	err = h.legacyRepo.Update(anchoredInvDoc.CoreDocument.DocumentIdentifier, anchoredInvDoc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("error saving document: %v", err))
	}

	return anchoredInvDoc, nil

}

// SendInvoiceDocument anchors and sends an invoice to the recipient
// Deprecated
func (h *grpcHandler) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *legacyinvoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	errs := []error{}
	doc, err := h.anchorInvoiceDocument(ctx, sendInvoiceEnvelope.Document, sendInvoiceEnvelope.Recipients)
	if err != nil {
		return nil, centerrors.Wrap(err, "error when anchoring document")
	}
	// Updating invoice with autogenerated fields after anchoring
	err = h.legacyRepo.Update(doc.CoreDocument.DocumentIdentifier, doc)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("error saving document: %v", err))
	}

	// Load the CoreDocument stored in DB before sending to Collaborators
	// So it contains the whole CoreDocument Data
	coreDoc := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(doc.CoreDocument.DocumentIdentifier, coreDoc)
	if err != nil {
		return nil, centerrors.New(code.DocumentNotFound, err.Error())
	}

	for _, recipient := range sendInvoiceEnvelope.Recipients {
		recipientID, err := identity.ToCentID(recipient)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		err = h.coreDocProcessor.Send(ctx, coreDoc, recipientID)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		apiLog.Errorf("%v", errs)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("%v", errs))
	}

	return doc, nil
}

// GetInvoiceDocument returns already stored invoice document
// Deprecated
func (h *grpcHandler) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *legacyinvoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc := new(invoicepb.InvoiceDocument)
	err := h.legacyRepo.GetByID(getInvoiceDocumentEnvelope.DocumentIdentifier, doc)
	if err == nil {
		return doc, nil
	}

	coreDoc := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(getInvoiceDocumentEnvelope.DocumentIdentifier, coreDoc)
	if err != nil {
		return nil, centerrors.New(code.DocumentNotFound, err.Error())
	}

	inv, err := NewFromCoreDocument(coreDoc)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return inv.Document, nil
}

// GetReceivedInvoiceDocuments returns all the received invoice documents
// Deprecated
func (h *grpcHandler) GetReceivedInvoiceDocuments(ctx context.Context, empty *empty.Empty) (*legacyinvoicepb.ReceivedInvoices, error) {
	return nil, nil
}

// Create handles the creation of the invoices and anchoring the documents on chain
func (h *grpcHandler) Create(ctx context.Context, req *clientinvoicepb.InvoiceCreatePayload) (*clientinvoicepb.InvoiceResponse, error) {
	doc, err := h.service.DeriveFromCreatePayload(req)
	if err != nil {
		return nil, err
	}

	// validate and persist
	doc, err = h.service.Create(ctx, doc)
	if err != nil {
		return nil, err
	}

	return h.service.DeriveInvoiceResponse(doc)
}

// Update handles the document update and anchoring
func (h *grpcHandler) Update(ctx context.Context, payload *clientinvoicepb.InvoiceUpdatePayload) (*clientinvoicepb.InvoiceResponse, error) {
	doc, err := h.service.DeriveFromUpdatePayload(payload)
	if err != nil {
		return nil, err
	}

	doc, err = h.service.Update(ctx, doc)
	if err != nil {
		return nil, err
	}

	return h.service.DeriveInvoiceResponse(doc)
}

// GetVersion returns the requested version of the document
func (h *grpcHandler) GetVersion(ctx context.Context, getVersionRequest *clientinvoicepb.GetVersionRequest) (*clientinvoicepb.InvoiceResponse, error) {
	identifier, err := hexutil.Decode(getVersionRequest.Identifier)
	if err != nil {
		return nil, centerrors.Wrap(err, "identifier is invalid")
	}
	version, err := hexutil.Decode(getVersionRequest.Version)
	if err != nil {
		return nil, centerrors.Wrap(err, "version is invalid")
	}
	doc, err := h.service.GetVersion(identifier, version)
	if err != nil {
		return nil, centerrors.Wrap(err, "document not found")
	}
	resp, err := h.service.DeriveInvoiceResponse(doc)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Get returns the invoice the latest version of the document with given identifier
func (h *grpcHandler) Get(ctx context.Context, getRequest *clientinvoicepb.GetRequest) (*clientinvoicepb.InvoiceResponse, error) {
	identifier, err := hexutil.Decode(getRequest.Identifier)
	if err != nil {
		return nil, centerrors.Wrap(err, "identifier is an invalid hex string")
	}
	doc, err := h.service.GetLastVersion(identifier)
	if err != nil {
		return nil, centerrors.Wrap(err, "document not found")
	}
	resp, err := h.service.DeriveInvoiceResponse(doc)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
