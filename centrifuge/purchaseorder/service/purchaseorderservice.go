package purchaseorderservice

import (
	"fmt"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var log = logging.Logger("rest-api")

// Struct needed as it is used to register the grpc services attached to the grpc server
type PurchaseOrderDocumentService struct{
	PurchaseOrderRepository purchaseorderrepository.PurchaseOrderRepository
	CoreDocumentProcessor   coredocument.CoreDocumentProcessorer
}

// HandleCreatePurchaseOrderProof creates proofs for a list of fields
func (s *PurchaseOrderDocumentService) HandleCreatePurchaseOrderProof(ctx context.Context, createPurchaseOrderProofEnvelope *purchaseorderpb.CreatePurchaseOrderProofEnvelope) (*purchaseorderpb.PurchaseOrderProof, error) {
	orderDoc, err := s.PurchaseOrderRepository.FindById(createPurchaseOrderProofEnvelope.DocumentIdentifier)
	if err != nil {
		return nil, err
	}

	order := purchaseorder.NewPurchaseOrder(orderDoc)

	proofs, err := order.CreateProofs(createPurchaseOrderProofEnvelope.Fields)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &purchaseorderpb.PurchaseOrderProof{FieldProofs: proofs, DocumentIdentifier: order.Document.CoreDocument.DocumentIdentifier}, nil

}

// HandleAnchorPurchaseOrderDocument anchors the given purchaseorder document and returns the anchor details
func (s *PurchaseOrderDocumentService) HandleAnchorPurchaseOrderDocument(ctx context.Context, anchorPurchaseOrderEnvelope *purchaseorderpb.AnchorPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	err := s.PurchaseOrderRepository.Store(anchorPurchaseOrderEnvelope.Document)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// TODO: the calculated merkle root should be persisted locally as well.
	orderDoc := purchaseorder.NewPurchaseOrder(anchorPurchaseOrderEnvelope.Document)
	orderDoc.CalculateMerkleRoot()
	coreDoc := orderDoc.ConvertToCoreDocument()

	err = s.CoreDocumentProcessor.Anchor(coreDoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return anchorPurchaseOrderEnvelope.Document, nil
}

// HandleSendPurchaseOrderDocument anchors and sends an purchaseorder to the recipient
func (s *PurchaseOrderDocumentService) HandleSendPurchaseOrderDocument(ctx context.Context, sendPurchaseOrderEnvelope *purchaseorderpb.SendPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	err := s.PurchaseOrderRepository.Store(sendPurchaseOrderEnvelope.Document)
	if err != nil {
		return nil, err
	}

	order := purchaseorder.NewPurchaseOrder(sendPurchaseOrderEnvelope.Document)
	order.CalculateMerkleRoot()
	coreDoc := order.ConvertToCoreDocument()

	errs := []error{}
	for _, element := range sendPurchaseOrderEnvelope.Recipients {
		err1 := s.CoreDocumentProcessor.Send(coreDoc, ctx, string(element[:]))
		if err1 != nil {
			errs = append(errs, err1)
		}
	}

	if len(errs) != 0 {
		log.Errorf("%v", errs)
		return nil, fmt.Errorf("%v", errs)
	}
	return sendPurchaseOrderEnvelope.Document, nil
}

func (s *PurchaseOrderDocumentService) HandleGetPurchaseOrderDocument(ctx context.Context, getPurchaseOrderDocumentEnvelope *purchaseorderpb.GetPurchaseOrderDocumentEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	doc, err := s.PurchaseOrderRepository.FindById(getPurchaseOrderDocumentEnvelope.DocumentIdentifier)
	if err != nil {
		doc1, err1 := coredocumentrepository.GetCoreDocumentRepository().FindById(getPurchaseOrderDocumentEnvelope.DocumentIdentifier)
		if err1 == nil {
			doc = purchaseorder.NewPurchaseOrderFromCoreDocument(doc1).Document
			err = err1
		}
		log.Errorf("%v", err)
	}
	return doc, err
}

func (s *PurchaseOrderDocumentService) HandleGetReceivedPurchaseOrderDocuments(ctx context.Context, empty *google_protobuf2.Empty) (*purchaseorderpb.ReceivedPurchaseOrders, error) {
	return nil, nil
}
