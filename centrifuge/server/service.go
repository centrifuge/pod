package server

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"log"
	pb "github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Below commands let you generate the coredocument protobuf stuff. It requires
// grpc-gateway protobuf labels to be checked out in a separate folder.
// NB: for now you will have to manually check out the protobuf & grpc-gateway project and update the path here
// To generate the go files, run: `cd centrifuge/server && go generate`
//go:generate protoc -I/Users/lucasvo/Code/fuge/protobuf/src/ -I ../coredocument  -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway --go_out=plugins=grpc:../coredocument ../coredocument/coredocument.proto
//go:generate protoc -I/Users/lucasvo/Code/fuge/protobuf/src/ -I ../coredocument  -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway --grpc-gateway_out=logtostderr=true:../coredocument ../coredocument/coredocument.proto
//go:generate protoc -I/Users/lucasvo/Code/fuge/protobuf/src/ -I ../coredocument  -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway --swagger_out=logtostderr=true:../coredocument ../coredocument/coredocument.proto

// coreDocumentService handles all interactions on our core documents
type centrifugeNodeService struct {}

// anchorDocument anchors a CoreDocument
func (s *centrifugeNodeService) AnchorDocument(ctx context.Context, doc *pb.CoreDocument) (*pb.CoreDocument, error) {
	log.Fatalf("Identifier: %v", doc.GetCurrentIdentifier())
	return doc, nil
}

func (s *centrifugeNodeService) SendInvoiceDocument(ctx context.Context, sendInvoiceEnvelope *pb.SendInvoiceEnvelope) (*pb.InvoiceDocument, error) {
	db := storage.GetStorage()
	db.StoreInvoiceDocument(sendInvoiceEnvelope.Document)
	return sendInvoiceEnvelope.Document, nil
}

func (s * centrifugeNodeService) GetInvoiceDocument(ctx context.Context, getInvoiceDocumentEnvelope *pb.GetInvoiceDocumentEnvelope) (*pb.InvoiceDocument, error) {
	db := storage.GetStorage()
	doc := db.GetInvoiceDocument(getInvoiceDocumentEnvelope.DocumentIdentifier)
	return doc, nil
}

// newServer creates our our service that is used by the centrifuge OS clients.
func newCentrifugeNodeService() *centrifugeNodeService {
	s := &centrifugeNodeService{}
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
