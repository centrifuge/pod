package purchaseorderservice

import (
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/centerrors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	clientpurchaseorderpb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	googleprotobuf2 "github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var log = logging.Logger("rest-api")

// PurchaseOrderDocumentService needed as it is used to register the grpc services attached to the grpc server
type PurchaseOrderDocumentService struct {
	Repository            storage.Repository
	CoreDocumentProcessor coredocumentprocessor.Processor
}

// anchorPurchaseOrderDocument anchors the given purchaseorder document and returns the anchor details
func (s *PurchaseOrderDocumentService) anchorPurchaseOrderDocument(ctx context.Context, doc *purchaseorderpb.PurchaseOrderDocument, collaborators []identity.CentID) (*purchaseorderpb.PurchaseOrderDocument, error) {
	orderDoc, err := purchaseorder.New(doc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// TODO review this create, do we need to refactor this because Send method also calls this?
	err = s.Repository.Create(orderDoc.Document.CoreDocument.DocumentIdentifier, orderDoc.Document)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to save document: %v", err))
	}

	coreDoc, err := orderDoc.ConvertToCoreDocument()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = s.CoreDocumentProcessor.Anchor(ctx, coreDoc, collaborators)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	newPo, err := purchaseorder.NewFromCoreDocument(coreDoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return newPo.Document, nil
}

// HandleCreatePurchaseOrderProof creates proofs for a list of fields
func (s *PurchaseOrderDocumentService) HandleCreatePurchaseOrderProof(ctx context.Context, createPurchaseOrderProofEnvelope *clientpurchaseorderpb.CreatePurchaseOrderProofEnvelope) (*clientpurchaseorderpb.PurchaseOrderProof, error) {
	orderDoc := new(purchaseorderpb.PurchaseOrderDocument)
	err := s.Repository.GetByID(createPurchaseOrderProofEnvelope.DocumentIdentifier, orderDoc)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.DocumentNotFound, err.Error())
	}

	order, err := purchaseorder.Wrap(orderDoc)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	proofs, err := order.CreateProofs(createPurchaseOrderProofEnvelope.Fields)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to create proofs: %v", err))
	}

	return &clientpurchaseorderpb.PurchaseOrderProof{FieldProofs: proofs, DocumentIdentifier: order.Document.CoreDocument.DocumentIdentifier}, nil
}

// HandleAnchorPurchaseOrderDocument anchors the given purchaseorder document and returns the anchor details
func (s *PurchaseOrderDocumentService) HandleAnchorPurchaseOrderDocument(ctx context.Context, anchorPurchaseOrderEnvelope *clientpurchaseorderpb.AnchorPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	anchoredPurchaseOrder, err := s.anchorPurchaseOrderDocument(ctx, anchorPurchaseOrderEnvelope.Document, nil)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to anchor document: %v", err))
	}

	// Updating purchaseorder with autogenerated fields after anchoring
	err = s.Repository.Update(anchoredPurchaseOrder.CoreDocument.DocumentIdentifier, anchoredPurchaseOrder)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to save document: %v", err))
	}

	return anchoredPurchaseOrder, nil
}

// HandleSendPurchaseOrderDocument anchors and sends an purchaseorder to the recipient
func (s *PurchaseOrderDocumentService) HandleSendPurchaseOrderDocument(ctx context.Context, sendPurchaseOrderEnvelope *clientpurchaseorderpb.SendPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	errs, recipientIDs := identity.ParseCentIDs(sendPurchaseOrderEnvelope.Recipients)
	if len(errs) != 0 {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("%v", errs))
	}
	doc, err := s.anchorPurchaseOrderDocument(ctx, sendPurchaseOrderEnvelope.Document, recipientIDs)
	if err != nil {
		return nil, err
	}
	// Updating purchaseorder with autogenerated fields after anchoring
	err = s.Repository.Update(doc.CoreDocument.DocumentIdentifier, doc)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to save document: %v", err))
	}

	for _, recipient := range recipientIDs {
		err = s.CoreDocumentProcessor.Send(ctx, doc.CoreDocument, recipient)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		log.Errorf("%v", errs)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("%v", errs))
	}

	return doc, nil
}

func (s *PurchaseOrderDocumentService) HandleGetPurchaseOrderDocument(ctx context.Context, getPurchaseOrderDocumentEnvelope *clientpurchaseorderpb.GetPurchaseOrderDocumentEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	doc := new(purchaseorderpb.PurchaseOrderDocument)
	err := s.Repository.GetByID(getPurchaseOrderDocumentEnvelope.DocumentIdentifier, doc)
	if err == nil {
		return doc, nil
	}

	// TODO(ved): where are we saving this coredocument?
	docFound := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(getPurchaseOrderDocumentEnvelope.DocumentIdentifier, docFound)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.DocumentNotFound, fmt.Sprintf("failed to get document: %v", err))
	}

	purchaseOrder, err := purchaseorder.NewFromCoreDocument(docFound)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed convert coredoc to purchase order: %v", err))
	}

	return purchaseOrder.Document, nil
}

func (s *PurchaseOrderDocumentService) HandleGetReceivedPurchaseOrderDocuments(ctx context.Context, empty *googleprotobuf2.Empty) (*clientpurchaseorderpb.ReceivedPurchaseOrders, error) {
	return nil, nil
}
