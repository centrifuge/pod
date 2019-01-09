package purchaseorder

import (
	"bytes"
	"context"
	"github.com/centrifuge/go-centrifuge/documents/genericdoc"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/notification"
	clientpopb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

var srvLog = logging.Logger("po-service")

// Service defines specific functions for purchase order
type Service interface {
	documents.Service

	// DeriverFromPayload derives purchase order from clientPayload
	DeriveFromCreatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderCreatePayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives purchase order from update payload
	DeriveFromUpdatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderUpdatePayload) (documents.Model, error)

	// Create validates and persists purchase order and returns a Updated model
	Create(ctx context.Context, po documents.Model) (documents.Model, uuid.UUID, error)

	// Update validates and updates the purchase order and return the updated model
	Update(ctx context.Context, po documents.Model) (documents.Model, uuid.UUID, error)

	// DerivePurchaseOrderData returns the purchase order data as client data
	DerivePurchaseOrderData(po documents.Model) (*clientpopb.PurchaseOrderData, error)

	// DerivePurchaseOrderResponse returns the purchase order in our standard client format
	DerivePurchaseOrderResponse(po documents.Model) (*clientpopb.PurchaseOrderResponse, error)
}

// service implements Service and handles all purchase order related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	repo             documents.Repository
	notifier         notification.Sender
	anchorRepository anchors.AnchorRepository
	identityService  identity.Service
	queueSrv         queue.TaskQueuer
	txService        transactions.Service
	genericdoc.Service
}

// DefaultService returns the default implementation of the service
func DefaultService(
	repo documents.Repository,
	anchorRepository anchors.AnchorRepository,
	identityService identity.Service,
	queueSrv queue.TaskQueuer,
	txService transactions.Service,
	genService genericdoc.Service,
) Service {
	return service{
		repo:             repo,
		notifier:         notification.NewWebhookSender(),
		anchorRepository: anchorRepository,
		identityService:  identityService,
		queueSrv:         queueSrv,
		txService:        txService,
		Service:       genService,
	}
}

// DeriveFromCoreDocument takes a core document and returns a purchase order
func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	var model documents.Model = new(PurchaseOrder)
	err := model.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return model, nil
}

// calculateDataRoot validates the document, calculates the data root, and persists to DB
func (s service) calculateDataRoot(ctx context.Context, old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	po, ok := new.(*PurchaseOrder)
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("unknown document type: %T", new))
	}

	// create data root, has to be done at the model level to access fields
	err = po.calculateDataRoot()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// validate the invoice
	err = validator.Validate(old, po)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(self.ID[:], po.CoreDocument.CurrentVersion, po)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	return po, nil
}

// Create validates, persists, and anchors a purchase order
func (s service) Create(ctx context.Context, po documents.Model) (documents.Model, uuid.UUID, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, uuid.Nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	po, err = s.calculateDataRoot(ctx, nil, po, CreateValidator())
	if err != nil {
		return nil, uuid.Nil, err
	}

	cd, err := po.PackCoreDocument()
	if err != nil {
		return nil, uuid.Nil, err
	}

	txID, err := documents.InitDocumentAnchorTask(s.queueSrv, s.txService, self.ID, cd.CurrentVersion)
	if err != nil {
		return nil, uuid.Nil, err
	}

	return po, txID, nil
}

// Update validates, persists, and anchors a new version of purchase order
func (s service) Update(ctx context.Context, po documents.Model) (documents.Model, uuid.UUID, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, uuid.Nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	cd, err := po.PackCoreDocument()
	if err != nil {
		return nil, uuid.Nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	old, err := s.GetCurrentVersion(ctx, cd.DocumentIdentifier)
	if err != nil {
		return nil, uuid.Nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	po, err = s.calculateDataRoot(ctx, old, po, UpdateValidator())
	if err != nil {
		return nil, uuid.Nil, err
	}

	txID, err := documents.InitDocumentAnchorTask(s.queueSrv, s.txService, self.ID, cd.CurrentVersion)
	if err != nil {
		return nil, uuid.Nil, err
	}

	return po, txID, nil
}

// DeriveFromCreatePayload derives purchase order from create payload
func (s service) DeriveFromCreatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderCreatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}

	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigTenantID
	}

	po := new(PurchaseOrder)
	err = po.InitPurchaseOrderInput(payload, idConf.ID.String())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	return po, nil
}

// DeriveFromUpdatePayload derives purchase order from update payload
func (s service) DeriveFromUpdatePayload(ctx context.Context, payload *clientpopb.PurchaseOrderUpdatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}

	// get latest old version of the document
	id, err := hexutil.Decode(payload.Identifier)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentIdentifier, errors.New("failed to decode identifier: %v", err))
	}

	old, err := s.GetCurrentVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	// load purchase order data
	po := new(PurchaseOrder)
	err = po.initPurchaseOrderFromData(payload.Data)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, errors.New("failed to load purchase order from data: %v", err))
	}

	// update core document
	oldCD, err := old.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigTenantID
	}

	collaborators := append([]string{idConf.ID.String()}, payload.Collaborators...)
	po.CoreDocument, err = coredocument.PrepareNewVersion(*oldCD, collaborators)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPrepareCoreDocument, err)
	}

	return po, nil
}

// DerivePurchaseOrderData returns po data from the model
func (s service) DerivePurchaseOrderData(doc documents.Model) (*clientpopb.PurchaseOrderData, error) {
	po, ok := doc.(*PurchaseOrder)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return po.getClientData(), nil
}

// DerivePurchaseOrderResponse returns po response from the model
func (s service) DerivePurchaseOrderResponse(doc documents.Model) (*clientpopb.PurchaseOrderResponse, error) {
	cd, err := doc.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	collaborators := make([]string, len(cd.Collaborators))
	for i, c := range cd.Collaborators {
		cid, err := identity.ToCentID(c)
		if err != nil {
			return nil, errors.NewTypedError(documents.ErrDocumentCollaborator, err)
		}
		collaborators[i] = cid.String()
	}

	h := &clientpopb.ResponseHeader{
		DocumentId:    hexutil.Encode(cd.DocumentIdentifier),
		VersionId:     hexutil.Encode(cd.CurrentVersion),
		Collaborators: collaborators,
	}

	data, err := s.DerivePurchaseOrderData(doc)
	if err != nil {
		return nil, err
	}

	return &clientpopb.PurchaseOrderResponse{
		Header: h,
		Data:   data,
	}, nil
}

func (s service) getPurchaseOrderVersion(ctx context.Context, documentID, version []byte) (model *PurchaseOrder, err error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}
	doc, err := s.repo.Get(self.ID[:], version)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentVersionNotFound, err)
	}
	model, ok := doc.(*PurchaseOrder)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	if !bytes.Equal(model.CoreDocument.DocumentIdentifier, documentID) {
		return nil, errors.NewTypedError(documents.ErrDocumentVersionNotFound, errors.New("version is not valid for this identifier"))
	}
	return model, nil
}

// GetLastVersion returns the latest version of the document
func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	model, err := s.getPurchaseOrderVersion(ctx, documentID, documentID)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}
	nextVersion := model.CoreDocument.NextVersion
	for nextVersion != nil {
		temp, err := s.getPurchaseOrderVersion(ctx, documentID, nextVersion)
		if err != nil {
			// here the err is returned as nil because it is expected that the nextVersion is not available in the db at some stage of the iteration
			return model, nil
		}

		model = temp
		nextVersion = model.CoreDocument.NextVersion
	}
	return model, nil
}

// GetVersion returns the specific version of the document
func (s service) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Model, error) {
	po, err := s.getPurchaseOrderVersion(ctx, documentID, version)
	if err != nil {
		return nil, err
	}
	return po, nil

}

// Exists checks if an purchase order exists
func (s service) Exists(ctx context.Context, documentID []byte) bool {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return false
	}
	return s.repo.Exists(self.ID[:], documentID)
}
