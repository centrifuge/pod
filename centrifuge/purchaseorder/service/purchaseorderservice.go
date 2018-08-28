package purchaseorderservice

import (
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
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

func fillCoreDocIdentifiers(doc *purchaseorderpb.PurchaseOrderDocument) error {
	if doc == nil {
		return errors.NilError(doc)
	}

	filledCoreDoc, err := coredocument.FillIdentifiers(*doc.CoreDocument)
	if err != nil {
		log.Error(err)
		return err
	}

	doc.CoreDocument = &filledCoreDoc
	return nil
}

// anchorPurchaseOrderDocument anchors the given purchaseorder document and returns the anchor details
func (s *PurchaseOrderDocumentService) anchorPurchaseOrderDocument(doc *purchaseorderpb.PurchaseOrderDocument) (*purchaseorderpb.PurchaseOrderDocument, error) {
	// TODO: the calculated merkle root should be persisted locally as well.
	orderDoc, err := purchaseorder.NewPurchaseOrder(doc)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	orderDoc.CalculateMerkleRoot()
	coreDoc := orderDoc.ConvertToCoreDocument()

	err = s.CoreDocumentProcessor.Anchor(coreDoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	newPo, err := purchaseorder.NewPurchaseOrderFromCoreDocument(coreDoc)
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
		return nil, errors.New(code.DocumentNotFound, err.Error())
	}

	order, err := purchaseorder.NewPurchaseOrder(orderDoc)
	if err != nil {
		log.Error(err)
		return nil, errors.New(code.Unknown, err.Error())
	}

	proofs, err := order.CreateProofs(createPurchaseOrderProofEnvelope.Fields)
	if err != nil {
		log.Error(err)
		return nil, errors.New(code.Unknown, fmt.Sprintf("failed to create proofs: %v", err))
	}

	return &clientpurchaseorderpb.PurchaseOrderProof{FieldProofs: proofs, DocumentIdentifier: order.Document.CoreDocument.DocumentIdentifier}, nil
}

// HandleAnchorPurchaseOrderDocument anchors the given purchaseorder document and returns the anchor details
func (s *PurchaseOrderDocumentService) HandleAnchorPurchaseOrderDocument(ctx context.Context, anchorPurchaseOrderEnvelope *clientpurchaseorderpb.AnchorPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	doc, err := purchaseorder.NewPurchaseOrder(anchorPurchaseOrderEnvelope.Document)
	if err != nil {
		return nil, errors.New(code.Unknown, err.Error())
	}

	err = fillCoreDocIdentifiers(doc.Document)
	if err != nil {
		log.Error(err)
		return nil, errors.New(code.Unknown, fmt.Sprintf("failed to fill document IDs: %v", err))
	}

	if valid, msg, errs := purchaseorder.Validate(doc.Document); !valid {
		return nil, errors.NewWithErrors(code.DocumentInvalid, msg, errs)
	}

	err = s.Repository.Create(doc.Document.CoreDocument.DocumentIdentifier, doc.Document)
	if err != nil {
		log.Error(err)
		return nil, errors.New(code.Unknown, fmt.Sprintf("failed to save document: %v", err))
	}

	anchoredPurchaseOrder, err := s.anchorPurchaseOrderDocument(doc.Document)
	if err != nil {
		log.Error(err)
		return nil, errors.New(code.Unknown, fmt.Sprintf("failed to anchor document: %v", err))
	}

	return anchoredPurchaseOrder, nil
}

// HandleSendPurchaseOrderDocument anchors and sends an purchaseorder to the recipient
func (s *PurchaseOrderDocumentService) HandleSendPurchaseOrderDocument(ctx context.Context, sendPurchaseOrderEnvelope *clientpurchaseorderpb.SendPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	doc, err := s.HandleAnchorPurchaseOrderDocument(ctx, &clientpurchaseorderpb.AnchorPurchaseOrderEnvelope{Document: sendPurchaseOrderEnvelope.Document})
	if err != nil {
		return nil, err
	}

	errs := make(map[string]string)
	for _, recipient := range sendPurchaseOrderEnvelope.Recipients {
		err = s.CoreDocumentProcessor.Send(doc.CoreDocument, ctx, recipient)
		if err != nil {
			errs[string(recipient)] = err.Error()
		}
	}

	if len(errs) != 0 {
		log.Errorf("%v", errs)
		return nil, errors.NewWithErrors(code.Unknown, "failed to send purchase order", errs)
	}

	return doc, nil
}

func (s *PurchaseOrderDocumentService) HandleGetPurchaseOrderDocument(ctx context.Context, getPurchaseOrderDocumentEnvelope *clientpurchaseorderpb.GetPurchaseOrderDocumentEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	doc := new(purchaseorderpb.PurchaseOrderDocument)
	err := s.Repository.GetByID(getPurchaseOrderDocumentEnvelope.DocumentIdentifier, doc)
	if err == nil {
		return doc, nil
	}

	docFound, err := coredocumentrepository.GetRepository().FindById(getPurchaseOrderDocumentEnvelope.DocumentIdentifier)
	if err != nil {
		log.Error(err)
		return nil, errors.New(code.DocumentNotFound, fmt.Sprintf("failed to get document: %v", err))
	}

	purchaseOrder, err := purchaseorder.NewPurchaseOrderFromCoreDocument(docFound)
	if err != nil {
		return nil, errors.New(code.Unknown, fmt.Sprintf("failed convert coredoc to purchase order: %v", err))
	}

	return purchaseOrder.Document, nil
}

func (s *PurchaseOrderDocumentService) HandleGetReceivedPurchaseOrderDocuments(ctx context.Context, empty *googleprotobuf2.Empty) (*clientpurchaseorderpb.ReceivedPurchaseOrders, error) {
	return nil, nil
}
