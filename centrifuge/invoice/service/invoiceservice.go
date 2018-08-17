package invoiceservice

import (
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/service"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	gerrors "github.com/go-errors/errors"
	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var log = logging.Logger("rest-api")

type InvoiceDocumentService struct {
	InvoiceRepository     invoicerepository.InvoiceRepository
	CoreDocumentProcessor coredocument.Processor
}

func fillCoreDocIdentifiers(doc *invoicepb.InvoiceDocument) error {
	if doc == nil {
		return errors.NilError(doc)
	}
	filledCoreDoc, err := coredocumentservice.AutoFillDocumentIdentifiers(*doc.CoreDocument)
	if err != nil {
		log.Error(err)
		return err
	}
	doc.CoreDocument = &filledCoreDoc
	return nil
}

// HandleCreateInvoiceProof creates proofs for a list of fields
func (s *InvoiceDocumentService) HandleCreateInvoiceProof(ctx context.Context, createInvoiceProofEnvelope *invoicepb.CreateInvoiceProofEnvelope) (*invoicepb.InvoiceProof, error) {
	invdoc, err := s.InvoiceRepository.FindById(createInvoiceProofEnvelope.DocumentIdentifier)
	if err != nil {
		return nil, err
	}

	inv, err := invoice.NewInvoice(invdoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	proofs, err := inv.CreateProofs(createInvoiceProofEnvelope.Fields)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &invoicepb.InvoiceProof{FieldProofs: proofs, DocumentIdentifier: inv.Document.CoreDocument.DocumentIdentifier}, nil

}

// HandleAnchorInvoiceDocument anchors the given invoice document and returns the anchor details
func (s *InvoiceDocumentService) HandleAnchorInvoiceDocument(ctx context.Context, anchorInvoiceEnvelope *invoicepb.AnchorInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc := anchorInvoiceEnvelope.Document

	err := fillCoreDocIdentifiers(doc)
	if err != nil {
		log.Error(err)
		return nil, gerrors.Errorf("Error filling document IDs: [%v]", err.Error())
	}

	err = s.InvoiceRepository.Create(doc)
	if err != nil {
		log.Error(err)
		return nil, gerrors.Errorf("Error saving document: [%v]", err.Error())
	}

	anchoredInvoiceDocument, err := s.anchorInvoiceDocument(doc)
	if err != nil {
		log.Error(err)
		return nil, gerrors.Errorf("Error anchoring document: [%v]", err.Error())
	}

	return anchoredInvoiceDocument, nil
}

// HandleSendInvoiceDocument anchors and sends an invoice to the recipient
func (s *InvoiceDocumentService) HandleSendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc := sendInvoiceEnvelope.Document

	err := fillCoreDocIdentifiers(doc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = s.InvoiceRepository.Create(doc)
	if err != nil {
		return nil, err
	}

	anchoredInvoiceDocument, err := s.anchorInvoiceDocument(doc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	errs := []error{}
	for _, element := range sendInvoiceEnvelope.Recipients {
		err1 := s.CoreDocumentProcessor.Send(anchoredInvoiceDocument.CoreDocument, ctx, element[:])
		if err1 != nil {
			errs = append(errs, err1)
		}
	}

	if len(errs) != 0 {
		log.Errorf("%v", errs)
		return nil, fmt.Errorf("%v", errs)
	}
	return anchoredInvoiceDocument, nil
}

func (s *InvoiceDocumentService) HandleGetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc, err := s.InvoiceRepository.FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
	if err != nil {
		docFound, err1 := coredocumentrepository.GetCoreDocumentRepository().FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
		if err1 == nil {
			doc1, err1 := invoice.NewInvoiceFromCoreDocument(docFound)
			doc = doc1.Document
			err = err1
		}
		log.Errorf("%v", err)
	}
	return doc, err
}

func (s *InvoiceDocumentService) HandleGetReceivedInvoiceDocuments(ctx context.Context, empty *google_protobuf2.Empty) (*invoicepb.ReceivedInvoices, error) {
	return nil, nil
}

// anchorInvoiceDocument anchors the given invoice document and returns the anchor details
func (s *InvoiceDocumentService) anchorInvoiceDocument(doc *invoicepb.InvoiceDocument) (*invoicepb.InvoiceDocument, error) {

	// TODO: the calculated merkle root should be persisted locally as well.
	inv, err := invoice.NewInvoice(doc)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	inv.CalculateMerkleRoot()
	coreDoc := inv.ConvertToCoreDocument()

	err = s.CoreDocumentProcessor.Anchor(coreDoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	newInvoice, err := invoice.NewInvoiceFromCoreDocument(coreDoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return newInvoice.Document, nil
}
