package invoicecontroller

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"context"
	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/service"
)

// Struct needed as it is used to register the grpc services attached to the grpc server
type InvoiceDocumentController struct {}

func (s *InvoiceDocumentController) AnchorInvoiceDocument(ctx context.Context, in *invoicepb.AnchorInvoiceEnvelope ) (*invoicepb.InvoiceDocument, error) {
	return nil, nil
}

func (s *InvoiceDocumentController) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoicepb.SendInvoiceEnvelope) (*invoicepb.InvoiceDocument, error) {
	var svc = &invoiceservice.InvoiceDocumentService{}
	return svc.HandleSendInvoiceDocument(ctx, sendInvoiceEnvelope)
}

func (s *InvoiceDocumentController) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoicepb.GetInvoiceDocumentEnvelope) (*invoicepb.InvoiceDocument, error) {
	var svc = &invoiceservice.InvoiceDocumentService{}
	return svc.HandleGetInvoiceDocument(ctx, getInvoiceDocumentEnvelope)
}

func (s *InvoiceDocumentController) GetReceivedInvoiceDocuments (ctx context.Context, empty *google_protobuf2.Empty) (*invoicepb.ReceivedInvoices, error) {
	var svc = &invoiceservice.InvoiceDocumentService{}
	return svc.HandleGetReceivedInvoiceDocuments(ctx, empty)
}