package genericdoc

import (
	"bytes"
	"context"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/common"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes"
	logging "github.com/ipfs/go-log"
)

// Service provides an interface for generic document methods
type Service interface {

	// GetCurrentVersion reads a document from the database
	GetCurrentVersion(documentID []byte) (documents.Model, error)

	// Exists checks if a document exists
	Exists(documentID []byte) bool

	// GetVersion reads a document from the database
	GetVersion(documentID []byte, version []byte) (documents.Model, error)

	// CreateProofs creates proofs for the latest version document given the fields
	CreateProofs(documentID []byte, fields []string) (*documents.DocumentProof, error)

	// CreateProofsForVersion creates proofs for a particular version of the document given the fields
	CreateProofsForVersion(documentID, version []byte, fields []string) (*documents.DocumentProof, error)

	// RequestDocumentSignature Validates and Signs document received over the p2p layer
	RequestDocumentSignature(ctx context.Context, model documents.Model) (*coredocumentpb.Signature, error)

	// ReceiveAnchoredDocument receives a new anchored document over the p2p layer, validates and updates the document in DB
	ReceiveAnchoredDocument(model documents.Model, headers *p2ppb.CentrifugeHeader) error
}

// service implements Service
type service struct {
	// TODO [multi-tenancy] replace this with config service
	config           documents.Config
	repo             documents.Repository
	identityService  identity.Service
	notifier         notification.Sender
	anchorRepository anchors.AnchorRepository
}

var srvLog = logging.Logger("document-service")

// DefaultService returns the default implementation of the service
func DefaultService(config config.Configuration, repo documents.Repository,
	anchorRepo anchors.AnchorRepository, idService identity.Service) documents.Service {
	return service{repo: repo,
		config:           config,
		anchorRepository: anchorRepo,
		notifier:         notification.NewWebhookSender(config),
		identityService:  idService}
}

func getIDs(model documents.Model) ([]byte, []byte, error) {
	cd, err := model.PackCoreDocument()

	if err != nil {
		return nil, nil, err
	}

	return cd.DocumentIdentifier, cd.NextVersion, nil
}

func (s service) searchVersion(m documents.Model) (documents.Model, error) {
	id, next, err := getIDs(m)

	if err != nil {
		return nil, err
	}

	if s.Exists(next) {
		nm, err := s.getVersion(id, next)
		if err != nil {

			return nil, err
		}
		return s.searchVersion(nm)
	}

	return m, nil

}

func (s service) GetCurrentVersion(documentID []byte) (documents.Model, error) {
	model, err := s.getVersion(documentID, documentID)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}
	return s.searchVersion(model)

}

func (s service) GetVersion(documentID []byte, version []byte) (documents.Model, error) {
	return s.getVersion(documentID, version)
}

func (s service) CreateProofs(documentID []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetCurrentVersion(documentID)
	if err != nil {
		return nil, err
	}
	return s.createProofs(model, fields)

}

func (s service) createProofs(model documents.Model, fields []string) (*documents.DocumentProof, error) {
	if err := coredocument.PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, model); err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}
	coreDoc, proofs, err := model.CreateProofs(fields)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentProof, err)
	}
	return &documents.DocumentProof{
		DocumentID:  coreDoc.DocumentIdentifier,
		VersionID:   coreDoc.CurrentVersion,
		FieldProofs: proofs,
	}, nil

}

func (s service) CreateProofsForVersion(documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.getVersion(documentID, version)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}
	return s.createProofs(model, fields)
}

func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {

	return nil, nil
}

func (s service) RequestDocumentSignature(ctx context.Context, model documents.Model) (*coredocumentpb.Signature, error) {
	if err := coredocument.SignatureRequestValidator(s.identityService).Validate(nil, model); err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	doc, err := model.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	srvLog.Infof("coredoc received %x with signing root %x", doc.DocumentIdentifier, doc.SigningRoot)
	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigTenantID
	}

	idKeys, ok := idConf.Keys[identity.KeyPurposeSigning]
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentSigning, errors.New("missing signing key"))
	}
	sig := crypto.Sign(idConf.ID[:], idKeys.PrivateKey, idKeys.PublicKey, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	err = model.UnpackCoreDocument(doc)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	tenantID := common.DummyIdentity.Bytes()

	// Logic for receiving version n (n > 1) of the document for the first time
	if !s.repo.Exists(tenantID, doc.DocumentIdentifier) && !utils.IsSameByteSlice(doc.DocumentIdentifier, doc.CurrentVersion) {
		err = s.repo.Create(tenantID, doc.DocumentIdentifier, model)
		if err != nil {
			return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
		}
	}

	err = s.repo.Create(tenantID, doc.CurrentVersion, model)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	srvLog.Infof("signed coredoc %x with version %x", doc.DocumentIdentifier, doc.CurrentVersion)
	return sig, nil
}

func (s service) ReceiveAnchoredDocument(model documents.Model, headers *p2ppb.CentrifugeHeader) error {
	if err := coredocument.PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, model); err != nil {
		return errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	doc, err := model.PackCoreDocument()
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	err = s.repo.Update(common.DummyIdentity.Bytes(), doc.CurrentVersion, model)
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

func (s service) Exists(documentID []byte) bool {
	return s.repo.Exists(common.DummyIdentity.Bytes(), documentID)
}

func (s service) getVersion(documentID, version []byte) (documents.Model, error) {
	model, err := s.repo.Get(common.DummyIdentity.Bytes(), version)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentVersionNotFound, err)
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(cd.DocumentIdentifier, documentID) {
		return nil, errors.NewTypedError(documents.ErrDocumentVersionNotFound, errors.New("version is not valid for this identifier"))
	}
	return model, nil
}
