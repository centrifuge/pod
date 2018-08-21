package invoicecontroller

import (
	"context"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/service"
	googleprotobuf2 "github.com/golang/protobuf/ptypes/empty"
)

// InvoiceDocumentController needed as it is used to register the grpc services attached to the grpc server
type InvoiceDocumentController struct{}

func getInvoiceDocumentService() *invoiceservice.InvoiceDocumentService {
	return &invoiceservice.InvoiceDocumentService{
		InvoiceRepository:     invoicerepository.GetInvoiceRepository(),
		CoreDocumentProcessor: coredocumentprocessor.NewDefaultProcessor(),
	}
}

func (s *InvoiceDocumentController) CreateInvoiceProof(ctx context.Context, createInvoiceProofEnvelope *invoicepb.CreateInvoiceProofEnvelope) (*invoicepb.InvoiceProof, error) {
	return getInvoiceDocumentService().HandleCreateInvoiceProof(ctx, createInvoiceProofEnvelope)
}

func (s *InvoiceDocumentController) AnchorInvoiceDocument(ctx context.Context, anchorInvoiceEnvelope *invoicepb.AnchorInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	return getInvoiceDocumentService().HandleAnchorInvoiceDocument(ctx, anchorInvoiceEnvelope)
}

func (s *InvoiceDocumentController) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	return getInvoiceDocumentService().HandleSendInvoiceDocument(ctx, sendInvoiceEnvelope)
}

func (s *InvoiceDocumentController) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	return getInvoiceDocumentService().HandleGetInvoiceDocument(ctx, getInvoiceDocumentEnvelope)
}

func (s *InvoiceDocumentController) GetReceivedInvoiceDocuments(ctx context.Context, empty *googleprotobuf2.Empty) (*invoicepb.ReceivedInvoices, error) {
	return getInvoiceDocumentService().HandleGetReceivedInvoiceDocuments(ctx, empty)
}
