package documents

import (
	"context"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
)

// Config defines required methods required for the documents package.
type Config interface {
	GetNetworkID() uint32
	GetIdentityID() ([]byte, error)
	GetP2PConnectionTimeout() time.Duration
}

// Client defines methods that can be implemented by any type handling p2p communications.
type Client interface {

	// GetSignaturesForDocument gets the signatures for document
	GetSignaturesForDocument(ctx context.Context, model *CoreDocumentModel) error

	// after all signatures are collected the sender sends the document including the signatures
	SendAnchoredDocument(ctx context.Context, receiverID identity.DID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error)
}

// defaultProcessor implements AnchorProcessor interface
type defaultProcessor struct {
	identityService  identity.ServiceDID
	p2pClient        Client
	anchorRepository anchors.AnchorRepository
	config           Config
}

// DefaultProcessor returns the default implementation of CoreDocument AnchorProcessor
func DefaultProcessor(idService identity.ServiceDID, p2pClient Client, repository anchors.AnchorRepository, config Config) AnchorProcessor {
	return defaultProcessor{
		identityService:  idService,
		p2pClient:        p2pClient,
		anchorRepository: repository,
		config:           config,
	}
}

// Send sends the given defaultProcessor to the given recipient on the P2P layer
func (dp defaultProcessor) Send(ctx context.Context, coreDocModel *CoreDocumentModel, recipient identity.DID) (err error) {
	if coreDocModel == nil {
		return errors.New("passed coreDocModel is nil")
	}
	if coreDocModel.Document == nil {
		return errors.New("passed coreDoc is nil")
	}
	coreDocument := coreDocModel.Document
	log.Infof("sending coredocument %x to recipient %x", coreDocument.DocumentIdentifier, recipient)

	c, _ := context.WithTimeout(ctx, dp.config.GetP2PConnectionTimeout())
	resp, err := dp.p2pClient.SendAnchoredDocument(c, recipient, &p2ppb.AnchorDocumentRequest{Document: coreDocument})
	if err != nil || !resp.Accepted {
		return errors.New("failed to send document to the node: %v", err)
	}

	log.Infof("Done opening connection against recipient [%x]\n", recipient)

	return nil
}

// PrepareForSignatureRequests gets the core document from the model, and adds the node's own signature
func (dp defaultProcessor) PrepareForSignatureRequests(ctx context.Context, model Model) error {
	dm, err := model.PackCoreDocument()

	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}
	cd := dm.Document
	dataRoot, err := model.CalculateDataRoot()
	if err != nil {
		return err
	}

	// calculate the signing root
	err = dm.CalculateSigningRoot(dataRoot)
	if err != nil {
		return errors.New("failed to calculate signing root: %v", err)
	}

	self, err := contextutil.Self(ctx)
	if err != nil {
		return err
	}

	sig := identity.Sign(self, identity.KeyPurposeSigning, cd.SigningRoot)
	cd.Signatures = append(cd.Signatures, sig)

	err = model.UnpackCoreDocument(dm)
	if err != nil {
		return errors.New("failed to unpack the core document: %v", err)
	}

	return nil
}

// RequestSignatures gets the core document from the model, validates pre signature requirements,
// collects signatures, and validates the signatures,
func (dp defaultProcessor) RequestSignatures(ctx context.Context, model Model) error {
	dm, err := model.PackCoreDocument()
	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}

	self, err := contextutil.Self(ctx)
	if err != nil {
		return err
	}

	idKeys, ok := self.Keys[identity.KeyPurposeSigning]
	if !ok {
		return errors.New("missing keys for signing")
	}

	psv := PreSignatureRequestValidator(self.ID[:], idKeys.PrivateKey, idKeys.PublicKey)
	err = psv.Validate(nil, model)
	if err != nil {
		return errors.New("failed to validate model for signature request: %v", err)
	}

	err = dp.p2pClient.GetSignaturesForDocument(ctx, dm)
	if err != nil {
		return errors.New("failed to collect signatures from the collaborators: %v", err)
	}

	err = model.UnpackCoreDocument(dm)
	if err != nil {
		return errors.New("failed to unpack core document: %v", err)
	}

	return nil
}

// PrepareForAnchoring validates the signatures and generates the document root
func (dp defaultProcessor) PrepareForAnchoring(model Model) error {
	dm, err := model.PackCoreDocument()
	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}

	psv := PostSignatureRequestValidator(dp.identityService)
	err = psv.Validate(nil, model)
	if err != nil {
		return errors.New("failed to validate signatures: %v", err)
	}

	err = dm.CalculateDocumentRoot()
	if err != nil {
		return errors.New("failed to generate document root: %v", err)
	}

	err = model.UnpackCoreDocument(dm)
	if err != nil {
		return errors.New("failed to unpack core document: %v", err)
	}

	return nil
}

// AnchorDocument validates the model, and anchors the document
func (dp defaultProcessor) AnchorDocument(ctx context.Context, model Model) error {
	dm, err := model.PackCoreDocument()
	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}

	pav := PreAnchorValidator(dp.identityService)
	err = pav.Validate(nil, model)
	if err != nil {
		return errors.New("pre anchor validation failed: %v", err)
	}

	cd := dm.Document
	rootHash, err := anchors.ToDocumentRoot(cd.DocumentRoot)
	if err != nil {
		return errors.New("failed to get document root: %v", err)
	}

	anchorID, err := anchors.ToAnchorID(cd.CurrentVersion)
	if err != nil {
		return errors.New("failed to get anchor ID: %v", err)
	}

	if err != nil {
		return errors.New("failed to generate ethereum MAC: %v", err)
	}

	log.Infof("Anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", cd.DocumentIdentifier, cd.CurrentVersion, cd.NextVersion, cd.DocumentRoot)
	err = dp.anchorRepository.CommitAnchor(ctx, anchorID, rootHash, [][anchors.DocumentProofLength]byte{utils.RandomByte32()})
	if err != nil {
		return errors.New("failed to commit anchor: %v", err)
	}

	log.Infof("Anchored document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", cd.DocumentIdentifier, cd.CurrentVersion, cd.NextVersion, cd.DocumentRoot)
	return nil
}

// SendDocument does post anchor validations and sends the document to collaborators
func (dp defaultProcessor) SendDocument(ctx context.Context, model Model) error {
	dm, err := model.PackCoreDocument()
	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}

	av := PostAnchoredValidator(dp.identityService, dp.anchorRepository)
	err = av.Validate(nil, model)
	if err != nil {
		return errors.New("post anchor validations failed: %v", err)
	}

	self, err := contextutil.Self(ctx)
	if err != nil {
		return err
	}

	extCollaborators, err := dm.GetExternalCollaborators(self.ID)
	if err != nil {
		return errors.New("get external collaborators failed: %v", err)
	}

	for _, c := range extCollaborators {
		cID := identity.NewDIDFromBytes(c)
		erri := dp.Send(ctx, dm, cID)
		if erri != nil {
			err = errors.AppendError(err, erri)
		}
	}

	return err
}
