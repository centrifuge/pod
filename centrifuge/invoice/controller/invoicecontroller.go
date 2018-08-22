package invoicecontroller

import (
	"context"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/service"
	clientinvoicepb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
)

// Struct needed as it is used to register the grpc services attached to the grpc server
type InvoiceDocumentController struct{}

func getInvoiceDocumentService() *invoiceservice.InvoiceDocumentService {
	return &invoiceservice.InvoiceDocumentService{
		InvoiceRepository:     invoicerepository.GetInvoiceRepository(),
		CoreDocumentProcessor: coredocument.GetDefaultCoreDocumentProcessor(),
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

func (s *InvoiceDocumentController) GetReceivedInvoiceDocuments(ctx context.Context, empty *google_protobuf2.Empty) (*clientinvoicepb.ReceivedInvoices, error) {
	return getInvoiceDocumentService().HandleGetReceivedInvoiceDocuments(ctx, empty)
}
