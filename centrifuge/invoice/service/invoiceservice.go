package invoiceservice

import (
	"fmt"
	invoicepb "github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var log = logging.Logger("rest-api")


type InvoiceDocumentService struct {
	InvoiceRepository invoicerepository.InvoiceRepository
}

// HandleCreateInvoiceProof creates proofs for a list of fields
func (s *InvoiceDocumentService) HandleCreateInvoiceProof(ctx context.Context, createInvoiceProofEnvelope *invoicepb.CreateInvoiceProofEnvelope) (*invoicepb.InvoiceProof, error) {
	invdoc, err := s.InvoiceRepository.FindById(createInvoiceProofEnvelope.DocumentIdentifier)
	if err != nil {
		return nil, err
	}

	inv := invoice.NewInvoice(invdoc)

	proofs, err := inv.CreateProofs(createInvoiceProofEnvelope.Fields)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &invoicepb.InvoiceProof{FieldProofs: proofs, DocumentIdentifier: inv.Document.CoreDocument.DocumentIdentifier}, nil

}

// HandleAnchorInvoiceDocument anchors the given invoice document and returns the anchor details
func (s *InvoiceDocumentService) HandleAnchorInvoiceDocument(ctx context.Context, anchorInvoiceEnvelope *invoicepb.AnchorInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	err := s.InvoiceRepository.Store(anchorInvoiceEnvelope.Document)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// TODO: the calculated merkle root should be persisted locally as well.
	inv := invoice.NewInvoice(anchorInvoiceEnvelope.Document)
	inv.CalculateMerkleRoot()
	coreDoc := inv.ConvertToCoreDocument()

	err = coreDoc.Anchor()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return anchorInvoiceEnvelope.Document, nil
}

// HandleSendInvoiceDocument anchors and sends an invoice to the recipient
func (s *InvoiceDocumentService) HandleSendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	err := s.InvoiceRepository.Store(sendInvoiceEnvelope.Document)
	if err != nil {
		return nil, err
	}

	inv := invoice.NewInvoice(sendInvoiceEnvelope.Document)
	inv.CalculateMerkleRoot()
	coreDoc := inv.ConvertToCoreDocument()

	errs := []error{}
	for _, element := range sendInvoiceEnvelope.Recipients {
		err1 := coredocument.SendP2P{}.Send(&coreDoc, ctx, string(element[:]))
		if err1 != nil {
			errs = append(errs, err1)
		}
	}

	if len(errs) != 0 {
		log.Errorf("%v", errs)
		return nil, fmt.Errorf("%v", errs)
	}
	return sendInvoiceEnvelope.Document, nil
}

func (s *InvoiceDocumentService) HandleGetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	doc, err := s.InvoiceRepository.FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
	if err != nil {
		doc1, err1 := coredocumentrepository.GetCoreDocumentRepository().FindById(getInvoiceDocumentEnvelope.DocumentIdentifier)
		if err1 == nil {
			doc = invoice.NewInvoiceFromCoreDocument(&coredocument.CoreDocument{doc1}).Document
			err = err1
		}
		log.Errorf("%v", err)
	}
	return doc, err
}

func (s *InvoiceDocumentService) HandleGetReceivedInvoiceDocuments(ctx context.Context, empty *google_protobuf2.Empty) (*invoicepb.ReceivedInvoices, error) {
	return nil, nil
}
