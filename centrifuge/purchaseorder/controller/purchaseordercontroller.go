package purchaseordercontroller

import (
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/service"
	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
)

// Struct needed as it is used to register the grpc services attached to the grpc server
type PurchaseOrderDocumentController struct{}

func (s *PurchaseOrderDocumentController) CreatePurchaseOrderProof(ctx context.Context, createPurchaseOrderProofEnvelope *purchaseorderpb.CreatePurchaseOrderProofEnvelope) (*purchaseorderpb.PurchaseOrderProof, error) {
	var svc = &purchaseorderservice.PurchaseOrderDocumentService{}
	return svc.HandleCreatePurchaseOrderProof(ctx, createPurchaseOrderProofEnvelope)
}

func (s *PurchaseOrderDocumentController) AnchorPurchaseOrderDocument(ctx context.Context, anchorPurchaseOrderEnvelope *purchaseorderpb.AnchorPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	var svc = &purchaseorderservice.PurchaseOrderDocumentService{}
	return svc.HandleAnchorPurchaseOrderDocument(ctx, anchorPurchaseOrderEnvelope)
}

func (s *PurchaseOrderDocumentController) SendPurchaseOrderDocument(ctx context.Context, sendPurchaseOrderEnvelope *purchaseorderpb.SendPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	var svc = &purchaseorderservice.PurchaseOrderDocumentService{}
	return svc.HandleSendPurchaseOrderDocument(ctx, sendPurchaseOrderEnvelope)
}

func (s *PurchaseOrderDocumentController) GetPurchaseOrderDocument(ctx context.Context, getPurchaseOrderDocumentEnvelope *purchaseorderpb.GetPurchaseOrderDocumentEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	var svc = &purchaseorderservice.PurchaseOrderDocumentService{}
	return svc.HandleGetPurchaseOrderDocument(ctx, getPurchaseOrderDocumentEnvelope)
}

func (s *PurchaseOrderDocumentController) GetReceivedPurchaseOrderDocuments(ctx context.Context, empty *google_protobuf2.Empty) (*purchaseorderpb.ReceivedPurchaseOrders, error) {
	var svc = &purchaseorderservice.PurchaseOrderDocumentService{}
	return svc.HandleGetReceivedPurchaseOrderDocuments(ctx, empty)
}
