package purchaseorder

import (
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	legacy "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/legacy/purchaseorder"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
)

var apiLog = logging.Logger("purchaseorder-api")

// grpcHandler handles all the purchase order document related actions
// anchoring, sending, finding stored purchase order document
type grpcHandler struct {
	repo        storage.LegacyRepository
	coreDocProc coredocumentprocessor.Processor
	service     Service
}

// LegacyGRPCHandler handles legacy grpc requests
func LegacyGRPCHandler(repo storage.LegacyRepository, proc coredocumentprocessor.Processor) legacy.PurchaseOrderDocumentServiceServer {
	return grpcHandler{
		repo:        repo,
		coreDocProc: proc,
	}
}

// GRPCHandler returns an implementation of the purchaseorder DocumentServiceServer
func GRPCHandler() (clientpurchaseorderpb.DocumentServiceServer, error) {
	srv, err := documents.GetRegistryInstance().LocateService(documenttypes.PurchaseOrderDataTypeUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch purchase order service")
	}

	return grpcHandler{
		service: srv.(Service),
	}, nil
}

// anchorPurchaseOrderDocument anchors the given purchaseorder document and returns the anchor details
// Deprecated
func (h grpcHandler) anchorPurchaseOrderDocument(ctx context.Context, doc *purchaseorderpb.PurchaseOrderDocument, collaborators [][]byte) (*purchaseorderpb.PurchaseOrderDocument, error) {
	orderDoc, err := New(doc, collaborators)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = h.repo.Create(orderDoc.Document.CoreDocument.DocumentIdentifier, orderDoc.Document)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to save document: %v", err))
	}

	coreDoc, err := orderDoc.ConvertToCoreDocument()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = h.coreDocProc.Anchor(ctx, coreDoc, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	newPo, err := NewFromCoreDocument(coreDoc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return newPo.Document, nil
}

// HandleCreatePurchaseOrderProof creates proofs for a list of fields
// Deprecated
func (h grpcHandler) CreatePurchaseOrderProof(ctx context.Context, createPurchaseOrderProofEnvelope *legacy.CreatePurchaseOrderProofEnvelope) (*legacy.PurchaseOrderProof, error) {
	orderDoc := new(purchaseorderpb.PurchaseOrderDocument)
	err := h.repo.GetByID(createPurchaseOrderProofEnvelope.DocumentIdentifier, orderDoc)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.DocumentNotFound, err.Error())
	}

	order, err := Wrap(orderDoc)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	// Set EmbeddedData
	order.Document.CoreDocument.EmbeddedData = &any.Any{TypeUrl: documenttypes.PurchaseOrderDataTypeUrl, Value: []byte{}}

	proofs, err := order.CreateProofs(createPurchaseOrderProofEnvelope.Fields)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to create proofs: %v", err))
	}

	return &legacy.PurchaseOrderProof{FieldProofs: proofs, DocumentIdentifier: order.Document.CoreDocument.DocumentIdentifier}, nil
}

// HandleAnchorPurchaseOrderDocument anchors the given purchaseorder document and returns the anchor details
// Deprecated
func (h grpcHandler) AnchorPurchaseOrderDocument(ctx context.Context, anchorPurchaseOrderEnvelope *legacy.AnchorPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	anchoredPurchaseOrder, err := h.anchorPurchaseOrderDocument(ctx, anchorPurchaseOrderEnvelope.Document, nil)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to anchor document: %v", err))
	}

	// Updating purchaseorder with autogenerated fields after anchoring
	err = h.repo.Update(anchoredPurchaseOrder.CoreDocument.DocumentIdentifier, anchoredPurchaseOrder)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to save document: %v", err))
	}

	return anchoredPurchaseOrder, nil
}

// HandleSendPurchaseOrderDocument anchors and sends an purchaseorder to the recipient
// Deprecated
func (h grpcHandler) SendPurchaseOrderDocument(ctx context.Context, sendPurchaseOrderEnvelope *legacy.SendPurchaseOrderEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	var errs []error
	doc, err := h.anchorPurchaseOrderDocument(ctx, sendPurchaseOrderEnvelope.Document, sendPurchaseOrderEnvelope.Recipients)
	if err != nil {
		return nil, err
	}
	// Updating purchaseorder with autogenerated fields after anchoring
	err = h.repo.Update(doc.CoreDocument.DocumentIdentifier, doc)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to save document: %v", err))
	}

	// Load the CoreDocument stored in DB before sending to Collaborators
	// So it contains the whole CoreDocument Data
	coreDoc := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(doc.CoreDocument.DocumentIdentifier, coreDoc)
	if err != nil {
		return nil, centerrors.New(code.DocumentNotFound, err.Error())
	}

	for _, recipient := range sendPurchaseOrderEnvelope.Recipients {
		recipientID, err := identity.ToCentID(recipient)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		err = h.coreDocProc.Send(ctx, coreDoc, recipientID)
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

// GetPurchaseOrderDocument returns the purchase order with specific identifier provided
// Deprecated
func (h grpcHandler) GetPurchaseOrderDocument(ctx context.Context, getPurchaseOrderDocumentEnvelope *legacy.GetPurchaseOrderDocumentEnvelope) (*purchaseorderpb.PurchaseOrderDocument, error) {
	doc := new(purchaseorderpb.PurchaseOrderDocument)
	err := h.repo.GetByID(getPurchaseOrderDocumentEnvelope.DocumentIdentifier, doc)
	if err == nil {
		return doc, nil
	}

	docFound := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(getPurchaseOrderDocumentEnvelope.DocumentIdentifier, docFound)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.DocumentNotFound, fmt.Sprintf("failed to get document: %v", err))
	}

	purchaseOrder, err := NewFromCoreDocument(docFound)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed convert coredoc to purchase order: %v", err))
	}

	return purchaseOrder.Document, nil
}

// GetReceivedPurchaseOrderDocuments returns received purchase order documents
// Deprecated
func (h grpcHandler) GetReceivedPurchaseOrderDocuments(ctx context.Context, empty *empty.Empty) (*legacy.ReceivedPurchaseOrders, error) {
	return nil, nil
}

// Create validates the purchase order, persists it to DB, and anchors it the chain
func (h grpcHandler) Create(ctx context.Context, req *clientpurchaseorderpb.PurchaseOrderCreatePayload) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	ctxh, err := documents.NewContextHeader()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	doc, err := h.service.DeriveFromCreatePayload(req, ctxh)
	if err != nil {
		return nil, err
	}

	// validate, persist, and anchor
	doc, err = h.service.Create(ctx, doc)
	if err != nil {
		return nil, err
	}

	return h.service.DerivePurchaseOrderResponse(doc)
}

func (grpcHandler) Update(context.Context, *clientpurchaseorderpb.PurchaseOrderUpdatePayload) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	apiLog.Error("Implement me")
	return nil, centerrors.New(code.Unknown, "Implement me")
}

func (grpcHandler) GetVersion(context.Context, *clientpurchaseorderpb.GetVersionRequest) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	apiLog.Error("Implement me")
	return nil, centerrors.New(code.Unknown, "Implement me")
}

func (grpcHandler) Get(context.Context, *clientpurchaseorderpb.GetRequest) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	apiLog.Error("Implement me")
	return nil, centerrors.New(code.Unknown, "Implement me")
}
