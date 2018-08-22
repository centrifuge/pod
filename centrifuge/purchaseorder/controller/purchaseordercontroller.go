package purchaseordercontroller

import (
	"context"

	purchaseorderpb "github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	clientpurchaseorderpb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/service"
	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
)

// Struct needed as it is used to register the grpc services attached to the grpc server
type PurchaseOrderDocumentController struct{}

func getPurchaseOrderDocumentService() *purchaseorderservice.PurchaseOrderDocumentService {
	return &purchaseorderservice.PurchaseOrderDocumentService{
		PurchaseOrderRepository: purchaseorderrepository.GetPurchaseOrderRepository(),
		CoreDocumentProcessor:   coredocument.GetDefaultCoreDocumentProcessor(),
	}
}

func (s *PurchaseOrderDocumentController) CreatePurchaseOrderProof(ctx context.Context, createPurchaseOrderProofEnvelope *clientpurchaseorderpb.CreatePurchaseOrderProofEnvelope) (*clientpurchaseorderpb.PurchaseOrderProof, error) {
	return getPurchaseOrderDocumentService().HandleCreatePurchaseOrderProof(ctx, createPurchaseOrderProofEnvelope)
}

func (s *PurchaseOrderDocumentController) AnchorPurchaseOrderDocument(ctx context.Context, anchorPurchaseOrderEnvelope *clientpurchaseorderpb.AnchorPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	return getPurchaseOrderDocumentService().HandleAnchorPurchaseOrderDocument(ctx, anchorPurchaseOrderEnvelope)
}

func (s *PurchaseOrderDocumentController) SendPurchaseOrderDocument(ctx context.Context, sendPurchaseOrderEnvelope *clientpurchaseorderpb.SendPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	return getPurchaseOrderDocumentService().HandleSendPurchaseOrderDocument(ctx, sendPurchaseOrderEnvelope)
}

func (s *PurchaseOrderDocumentController) GetPurchaseOrderDocument(ctx context.Context, getPurchaseOrderDocumentEnvelope *clientpurchaseorderpb.GetPurchaseOrderDocumentEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	return getPurchaseOrderDocumentService().HandleGetPurchaseOrderDocument(ctx, getPurchaseOrderDocumentEnvelope)
}

func (s *PurchaseOrderDocumentController) GetReceivedPurchaseOrderDocuments(ctx context.Context, empty *google_protobuf2.Empty) (*clientpurchaseorderpb.ReceivedPurchaseOrders, error) {
	return getPurchaseOrderDocumentService().HandleGetReceivedPurchaseOrderDocuments(ctx, empty)
}
