package server

import (
	"log"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage/invoicestorage"
)

// coreDocumentService handles all interactions on our core documents
type invoiceDocumentService struct {
	DataStore storage.DataStore
	invoiceStorageService *invoicestorage.StorageService
}

// Sets up the service's datastore
func (s *invoiceDocumentService) Init () {
	s.invoiceStorageService = &invoicestorage.StorageService{}
	s.invoiceStorageService.SetStorageBackend(s.DataStore)
}

// anchorDocument anchors a CoreDocument
func (s *invoiceDocumentService) AnchorDocument(ctx context.Context, doc *invoice.InvoiceDocument) (*invoice.InvoiceDocument, error) {
	log.Fatalf("Identifier: %v", doc.GetDocumentIdentifier())
	return doc, nil
}

func (s *invoiceDocumentService) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *invoice.SendInvoiceEnvelope) (*invoice.InvoiceDocument, error) {
	s.invoiceStorageService.PutDocument(sendInvoiceEnvelope.Document)
	return sendInvoiceEnvelope.Document, nil
}

func (s *invoiceDocumentService) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *invoice.GetInvoiceDocumentEnvelope) (*invoice.InvoiceDocument, error) {
	doc, err := s.invoiceStorageService.GetDocument(getInvoiceDocumentEnvelope.DocumentIdentifier)
	return doc, err
}

// newServer creates our our service that is used by the centrifuge OS clients.
func newInvoiceDocumentService() *invoiceDocumentService {
	s := &invoiceDocumentService{
		DataStore: storage.GetStorage(),
	}
	s.Init()

	return s
}

// RegisterServices registers all endpoints to the grpc server
func RegisterServices(grpcServer *grpc.Server, ctx context.Context, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) {
	invoice.RegisterInvoiceDocumentServiceServer(grpcServer, newInvoiceDocumentService())
	err := invoice.RegisterInvoiceDocumentServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}

}
