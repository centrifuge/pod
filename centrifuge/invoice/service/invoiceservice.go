package invoiceservice

import (
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	clientinvoicepb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var log = logging.Logger("rest-api")

type InvoiceDocumentService struct {
	InvoiceRepository     invoicerepository.InvoiceRepository
	CoreDocumentProcessor coredocumentprocessor.Processor
}

// anchorInvoiceDocument anchors the given invoice document and returns the anchored document
func (s *InvoiceDocumentService) anchorInvoiceDocument(doc *invoicepb.InvoiceDocument) (*invoicepb.InvoiceDocument, error) {

	inv, err := invoice.New(doc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// TODO(ved): should we also persist this on invoice repo?
	coreDoc := inv.ConvertToCoreDocument()
	err = s.CoreDocumentProcessor.Anchor(coreDoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	newInvoice, err := invoice.NewFromCoreDocument(coreDoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return newInvoice.Document, nil
}

// HandleCreateInvoiceProof creates proofs for a list of fields
func (s *InvoiceDocumentService) HandleCreateInvoiceProof(ctx context.Context, createInvoiceProofEnvelope *clientinvoicepb.CreateInvoiceProofEnvelope) (*clientinvoicepb.InvoiceProof, error) {
	invdoc, err := s.InvoiceRepository.FindById(createInvoiceProofEnvelope.DocumentIdentifier)
	if err != nil {
		return nil, err
	}

	inv, err := invoice.New(invdoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	proofs, err := inv.CreateProofs(createInvoiceProofEnvelope.Fields)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &clientinvoicepb.InvoiceProof{FieldProofs: proofs, DocumentIdentifier: inv.Document.CoreDocument.DocumentIdentifier}, nil

}

// HandleAnchorInvoiceDocument anchors the given invoice document and returns the anchor details
func (s *InvoiceDocumentService) HandleAnchorInvoiceDocument(ctx context.Context, anchorInvoiceEnvelope *clientinvoicepb.AnchorInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	inv, err := invoice.New(anchorInvoiceEnvelope.Document)
	if err != nil {
		log.Error(err)
		return nil, errors.New(code.DocumentInvalid, err.Error())
	}

	if valid, msg, errs := invoice.Validate(inv.Document); !valid {
		return nil, errors.NewWithErrors(code.DocumentInvalid, msg, errs)
	}

	err = s.InvoiceRepository.Create(inv.Document)
	if err != nil {
		log.Error(err)
		return nil, errors.New(code.Unknown, fmt.Sprintf("error saving invoice: %v", err))
	}

	anchoredInvDoc, err := s.anchorInvoiceDocument(inv.Document)
	if err != nil {
		log.Error(err)
		return nil, errors.New(code.Unknown, fmt.Sprintf("failed to anchor: %v", err))
	}

	return anchoredInvDoc, nil
}

// HandleSendInvoiceDocument anchors and sends an invoice to the recipient
func (s *InvoiceDocumentService) HandleSendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *clientinvoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc, err := s.HandleAnchorInvoiceDocument(ctx, &clientinvoicepb.AnchorInvoiceEnvelope{Document: sendInvoiceEnvelope.Document})
	if err != nil {
		return nil, err
	}

	var errs []error
	for _, element := range sendInvoiceEnvelope.Recipients {
		err = s.CoreDocumentProcessor.Send(doc.CoreDocument, ctx, element[:])
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		log.Errorf("%v", errs)
		return nil, errors.New(code.Unknown, fmt.Sprintf("%v", errs))
	}

	return doc, nil
}

// HandleGetInvoiceDocument returns already stored invoice document
func (s *InvoiceDocumentService) HandleGetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *clientinvoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc, err := s.InvoiceRepository.FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
	if err == nil {
		return doc, nil
	}

	coreDoc, err := coredocumentrepository.GetRepository().FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
	if err != nil {
		return nil, errors.New(code.DocumentNotFound, err.Error())
	}

	inv, err := invoice.NewFromCoreDocument(coreDoc)
	if err != nil {
		return nil, errors.New(code.Unknown, err.Error())
	}

	return inv.Document, nil
}

// HandleGetReceivedInvoiceDocuments returns all the received invoice documents
func (s *InvoiceDocumentService) HandleGetReceivedInvoiceDocuments(ctx context.Context, empty *empty.Empty) (*clientinvoicepb.ReceivedInvoices, error) {
	return nil, nil
}
