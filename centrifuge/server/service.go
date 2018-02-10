package server

import (
	"log"
	pb "github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage/invoicestorage"
)

// coreDocumentService handles all interactions on our core documents
type centrifugeNodeService struct {
	DataStore storage.DataStore
	invoiceStorageService *invoicestorage.StorageService
}

// Sets up the service's datastore
func (s *centrifugeNodeService) Init () {
	s.invoiceStorageService = &invoicestorage.StorageService{}
	s.invoiceStorageService.SetStorageBackend(s.DataStore)
}

// anchorDocument anchors a CoreDocument
func (s *centrifugeNodeService) AnchorDocument(ctx context.Context, doc *pb.CoreDocument) (*pb.CoreDocument, error) {
	log.Fatalf("Identifier: %v", doc.GetCurrentIdentifier())
	return doc, nil
}

func (s *centrifugeNodeService) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *pb.SendInvoiceEnvelope) (*pb.InvoiceDocument, error) {
	s.invoiceStorageService.PutDocument(sendInvoiceEnvelope.Document)
	return sendInvoiceEnvelope.Document, nil
}

func (s *centrifugeNodeService) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *pb.GetInvoiceDocumentEnvelope) (*pb.InvoiceDocument, error) {
	doc, err := s.invoiceStorageService.GetDocument(getInvoiceDocumentEnvelope.DocumentIdentifier)
	return doc, err
}

// newServer creates our our service that is used by the centrifuge OS clients.
func newCentrifugeNodeService() *centrifugeNodeService {
	s := &centrifugeNodeService{
		DataStore: storage.GetStorage(),
	}
	s.Init()

	return s
}

// RegisterServices registers all endpoints to the grpc server
func RegisterServices(grpcServer *grpc.Server, ctx context.Context, gwmux *runtime.ServeMux, addr string, dopts []grpc.DialOption) {
	pb.RegisterCentrifugeNodeServiceServer(grpcServer, newCentrifugeNodeService())
	err := pb.RegisterCentrifugeNodeServiceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		panic(err)
	}

}
