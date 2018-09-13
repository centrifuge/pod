package invoicecontroller

import (
	"context"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/service"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	clientinvoicepb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/golang/protobuf/ptypes/empty"
)

// InvoiceDocumentController needed as it is used to register the grpc services attached to the grpc server
type InvoiceDocumentController struct{}

func getInvoiceDocumentService() *invoiceservice.InvoiceDocumentService {
	return &invoiceservice.InvoiceDocumentService{
		InvoiceRepository:     invoicerepository.GetRepository(),
		CoreDocumentProcessor: coredocumentprocessor.DefaultProcessor(identity.IDService, p2p.NewP2PClient()),
	}
}

func (s *InvoiceDocumentController) CreateInvoiceProof(ctx context.Context, createInvoiceProofEnvelope *clientinvoicepb.CreateInvoiceProofEnvelope) (*clientinvoicepb.InvoiceProof, error) {
	return getInvoiceDocumentService().HandleCreateInvoiceProof(ctx, createInvoiceProofEnvelope)
}

func (s *InvoiceDocumentController) AnchorInvoiceDocument(ctx context.Context, anchorInvoiceEnvelope *clientinvoicepb.AnchorInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	return getInvoiceDocumentService().HandleAnchorInvoiceDocument(ctx, anchorInvoiceEnvelope)
}

func (s *InvoiceDocumentController) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *clientinvoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	return getInvoiceDocumentService().HandleSendInvoiceDocument(ctx, sendInvoiceEnvelope)
}

func (s *InvoiceDocumentController) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *clientinvoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	return getInvoiceDocumentService().HandleGetInvoiceDocument(ctx, getInvoiceDocumentEnvelope)
}

func (s *InvoiceDocumentController) GetReceivedInvoiceDocuments(ctx context.Context, empty *empty.Empty) (*clientinvoicepb.ReceivedInvoices, error) {
	return getInvoiceDocumentService().HandleGetReceivedInvoiceDocuments(ctx, empty)
}
