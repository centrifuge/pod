package invoicecontroller

import (
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/service"
	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
)

// Struct needed as it is used to register the grpc services attached to the grpc server
type InvoiceDocumentController struct{}

func getInvoicDocumentService() *invoiceservice.InvoiceDocumentService {
	return &invoiceservice.InvoiceDocumentService{
		InvoiceRepository: invoicerepository.GetInvoiceRepository(),
	}
}

func (s *InvoiceDocumentController) CreateInvoiceProof(ctx context.Context, createInvoiceProofEnvelope *invoicepb.CreateInvoiceProofEnvelope) (*invoicepb.InvoiceProof, error) {
	return getInvoicDocumentService().HandleCreateInvoiceProof(ctx, createInvoiceProofEnvelope)
}

func (s *InvoiceDocumentController) AnchorInvoiceDocument(ctx context.Context, anchorInvoiceEnvelope *invoicepb.AnchorInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	return getInvoicDocumentService().HandleAnchorInvoiceDocument(ctx, anchorInvoiceEnvelope)
}

func (s *InvoiceDocumentController) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	return getInvoicDocumentService().HandleSendInvoiceDocument(ctx, sendInvoiceEnvelope)
}

func (s *InvoiceDocumentController) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	return getInvoicDocumentService().HandleGetInvoiceDocument(ctx, getInvoiceDocumentEnvelope)
}

func (s *InvoiceDocumentController) GetReceivedInvoiceDocuments(ctx context.Context, empty *google_protobuf2.Empty) (*invoicepb.ReceivedInvoices, error) {
	return getInvoicDocumentService().HandleGetReceivedInvoiceDocuments(ctx, empty)
}
