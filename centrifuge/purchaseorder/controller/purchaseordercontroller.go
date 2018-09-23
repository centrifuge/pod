package purchaseordercontroller

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/p2p"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/purchaseorder/service"
	googleprotobuf2 "github.com/golang/protobuf/ptypes/empty"
)

// PurchaseOrderDocumentController needed as it is used to register the grpc services attached to the grpc server
type PurchaseOrderDocumentController struct{}

func getPurchaseOrderDocumentService() *purchaseorderservice.PurchaseOrderDocumentService {
	return &purchaseorderservice.PurchaseOrderDocumentService{
		Repository:            purchaseorderrepository.GetRepository(),
		CoreDocumentProcessor: coredocumentprocessor.DefaultProcessor(identity.IDService, p2p.NewP2PClient()),
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

func (s *PurchaseOrderDocumentController) GetReceivedPurchaseOrderDocuments(ctx context.Context, empty *googleprotobuf2.Empty) (*clientpurchaseorderpb.ReceivedPurchaseOrders, error) {
	return getPurchaseOrderDocumentService().HandleGetReceivedPurchaseOrderDocuments(ctx, empty)
}
