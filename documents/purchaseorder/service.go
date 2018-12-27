package purchaseorder

import (
	"bytes"
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/go-centrifuge/crypto"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/notification"
	clientpopb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes"
	logging "github.com/ipfs/go-log"
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
	Create(ctx context.Context, po documents.Model) (documents.Model, error)

	// Update validates and updates the purchase order and return the updated model
	Update(ctx context.Context, po documents.Model) (documents.Model, error)

	// DerivePurchaseOrderData returns the purchase order data as client data
	DerivePurchaseOrderData(po documents.Model) (*clientpopb.PurchaseOrderData, error)

	// DerivePurchaseOrderResponse returns the purchase order in our standard client format
	DerivePurchaseOrderResponse(po documents.Model) (*clientpopb.PurchaseOrderResponse, error)
}

// service implements Service and handles all purchase order related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	// TODO [multi-tenancy] replace this with config service
	config           documents.Config
	repo             documents.Repository
	coreDocProcessor coredocument.Processor
	notifier         notification.Sender
	anchorRepository anchors.AnchorRepository
	identityService  identity.Service
}

// DefaultService returns the default implementation of the service
func DefaultService(config config.Configuration, repo documents.Repository, processor coredocument.Processor, anchorRepository anchors.AnchorRepository, identityService identity.Service) Service {
	return service{config: config, repo: repo, coreDocProcessor: processor, notifier: notification.NewWebhookSender(config), anchorRepository: anchorRepository, identityService: identityService}
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
func (s service) calculateDataRoot(old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	po, ok := new.(*PurchaseOrder)
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("unknown document type: %T", new))
	}

	// create data root, has to be done at the model level to access fields
	err := po.calculateDataRoot()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// validate the invoice
	err = validator.Validate(old, po)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// get tenant ID
	tenantID, err := s.config.GetIdentityID()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(tenantID, po.CoreDocument.CurrentVersion, po)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	return po, nil
}

// Create validates, persists, and anchors a purchase order
func (s service) Create(ctx context.Context, po documents.Model) (documents.Model, error) {
	po, err := s.calculateDataRoot(nil, po, CreateValidator())
	if err != nil {
		return nil, err
	}

	po, err = documents.AnchorDocument(ctx, po, s.coreDocProcessor, s.updater)
	if err != nil {
		return nil, err
	}

	return po, nil
}

// updater wraps logic related to updating documents so that it can be executed as a closure
func (s service) updater(id []byte, model documents.Model) error {
	// get tenant ID
	tenantID, err := s.config.GetIdentityID()
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}
	return s.repo.Update(tenantID, id, model)
}

// Update validates, persists, and anchors a new version of purchase order
func (s service) Update(ctx context.Context, po documents.Model) (documents.Model, error) {
	cd, err := po.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	old, err := s.GetCurrentVersion(cd.DocumentIdentifier)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	po, err = s.calculateDataRoot(old, po, UpdateValidator())
	if err != nil {
		return nil, err
	}

	po, err = documents.AnchorDocument(ctx, po, s.coreDocProcessor, s.updater)
	if err != nil {
		return nil, err
	}

	return po, nil
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

	old, err := s.GetCurrentVersion(id)
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

func (s service) getPurchaseOrderVersion(documentID, version []byte) (model *PurchaseOrder, err error) {
	// get tenant ID
	tenantID, err := s.config.GetIdentityID()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}
	doc, err := s.repo.Get(tenantID, version)
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
func (s service) GetCurrentVersion(documentID []byte) (documents.Model, error) {
	model, err := s.getPurchaseOrderVersion(documentID, documentID)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}
	nextVersion := model.CoreDocument.NextVersion
	for nextVersion != nil {
		temp, err := s.getPurchaseOrderVersion(documentID, nextVersion)
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
func (s service) GetVersion(documentID []byte, version []byte) (documents.Model, error) {
	po, err := s.getPurchaseOrderVersion(documentID, version)
	if err != nil {
		return nil, err
	}
	return po, nil

}

// purchaseOrderProof creates proofs for purchaseOrder model fields
func (s service) purchaseOrderProof(model documents.Model, fields []string) (*documents.DocumentProof, error) {
	po, ok := model.(*PurchaseOrder)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}
	if err := coredocument.PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, po); err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}
	coreDoc, proofs, err := po.CreateProofs(fields)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentProof, err)
	}
	return &documents.DocumentProof{
		DocumentID:  coreDoc.DocumentIdentifier,
		VersionID:   coreDoc.CurrentVersion,
		FieldProofs: proofs,
	}, nil
}

// CreateProofs generates proofs for given document
func (s service) CreateProofs(documentID []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetCurrentVersion(documentID)
	if err != nil {
		return nil, err
	}
	return s.purchaseOrderProof(model, fields)
}

// CreateProofsForVersion generates proofs for specific version of the document
func (s service) CreateProofsForVersion(documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetVersion(documentID, version)
	if err != nil {
		return nil, err
	}
	return s.purchaseOrderProof(model, fields)
}

// RequestDocumentSignature validates the document and returns the signature
// Note: this is document agnostic. But since we do not have a common implementation, adding it here.
// will remove this once we have a common implementation for documents.Service
func (s service) RequestDocumentSignature(ctx context.Context, model documents.Model) (*coredocumentpb.Signature, error) {
	if err := coredocument.SignatureRequestValidator(s.identityService).Validate(nil, model); err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	srvLog.Infof("coredoc received %x with signing root %x", cd.DocumentIdentifier, cd.SigningRoot)

	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigTenantID
	}

	idKeys, ok := idConf.Keys[identity.KeyPurposeSigning]
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentSigning, errors.New("missing signing key"))
	}
	sig := crypto.Sign(idConf.ID[:], idKeys.PrivateKey, idKeys.PublicKey, cd.SigningRoot)
	cd.Signatures = append(cd.Signatures, sig)
	err = model.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	// get tenant ID
	tenantID, err := s.config.GetIdentityID()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	// Logic for receiving version n (n > 1) of the document for the first time
	if !s.repo.Exists(tenantID, cd.DocumentIdentifier) && !utils.IsSameByteSlice(cd.DocumentIdentifier, cd.CurrentVersion) {
		err = s.repo.Create(tenantID, cd.DocumentIdentifier, model)
		if err != nil {
			return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
		}
	}

	err = s.repo.Create(tenantID, cd.CurrentVersion, model)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	srvLog.Infof("signed coredoc %x with version %x", cd.DocumentIdentifier, cd.CurrentVersion)
	return sig, nil
}

// ReceiveAnchoredDocument validates the anchored document and updates it on DB
// Note: this is document agnostic. But since we do not have a common implementation, adding it here.
// will remove this once we have a common implementation for documents.Service
func (s service) ReceiveAnchoredDocument(model documents.Model, headers *p2ppb.CentrifugeHeader) error {
	if err := coredocument.PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, model); err != nil {
		return errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	doc, err := model.PackCoreDocument()
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	// get tenant ID
	tenantID, err := s.config.GetIdentityID()
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	err = s.repo.Update(tenantID, doc.CurrentVersion, model)
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	ts, _ := ptypes.TimestampProto(time.Now().UTC())
	notificationMsg := &notificationpb.NotificationMessage{
		EventType:    uint32(notification.ReceivedPayload),
		CentrifugeId: hexutil.Encode(headers.SenderCentrifugeId),
		Recorded:     ts,
		DocumentType: doc.EmbeddedData.TypeUrl,
		DocumentId:   hexutil.Encode(doc.DocumentIdentifier),
	}

	// Async until we add queuing
	go s.notifier.Send(notificationMsg)

	return nil
}

// Exists checks if an purchase order exists
func (s service) Exists(documentID []byte) bool {
	// get tenant ID
	tenantID, err := s.config.GetIdentityID()
	if err != nil {
		return false
	}
	return s.repo.Exists(tenantID, documentID)
}
