package documents

import (
	"context"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
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
	GetSignaturesForDocument(ctx context.Context, doc *coredocumentpb.CoreDocument) error

	// after all signatures are collected the sender sends the document including the signatures
	SendAnchoredDocument(ctx context.Context, receiverID identity.CentID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error)
}

// defaultProcessor implements AnchorProcessor interface
type defaultProcessor struct {
	identityService  identity.Service
	p2pClient        Client
	anchorRepository anchors.AnchorRepository
	config           Config
}

// DefaultProcessor returns the default implementation of CoreDocument AnchorProcessor
func DefaultProcessor(idService identity.Service, p2pClient Client, repository anchors.AnchorRepository, config Config) AnchorProcessor {
	return defaultProcessor{
		identityService:  idService,
		p2pClient:        p2pClient,
		anchorRepository: repository,
		config:           config,
	}
}

// Send sends the given defaultProcessor to the given recipient on the P2P layer
func (dp defaultProcessor) Send(ctx context.Context, coreDocument *coredocumentpb.CoreDocument, recipient identity.CentID) (err error) {
	if coreDocument == nil {
		return errors.New("passed coreDoc is nil")
	}
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
	cd, err := model.PackCoreDocument()
	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}

	dataRoot, err := model.CalculateDataRoot()
	if err != nil {
		return err
	}

	// calculate the signing root
	err = coredocument.CalculateSigningRoot(cd, dataRoot)
	if err != nil {
		return errors.New("failed to calculate signing root: %v", err)
	}

	self, err := contextutil.Self(ctx)
	if err != nil {
		return err
	}

	sig := identity.Sign(self, identity.KeyPurposeSigning, cd.SigningRoot)
	if cd.SignatureData == nil {
		cd.SignatureData = new(coredocumentpb.SignatureData)
	}
	cd.SignatureData.Signatures = append(cd.SignatureData.Signatures, sig)

	err = model.UnpackCoreDocument(cd)
	if err != nil {
		return errors.New("failed to unpack the core document: %v", err)
	}

	return nil
}

// RequestSignatures gets the core document from the model, validates pre signature requirements,
// collects signatures, and validates the signatures,
func (dp defaultProcessor) RequestSignatures(ctx context.Context, model Model) error {
	cd, err := model.PackCoreDocument()
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

	err = dp.p2pClient.GetSignaturesForDocument(ctx, cd)
	if err != nil {
		return errors.New("failed to collect signatures from the collaborators: %v", err)
	}

	err = model.UnpackCoreDocument(cd)
	if err != nil {
		return errors.New("failed to unpack core document: %v", err)
	}

	return nil
}

// PrepareForAnchoring validates the signatures and generates the document root
func (dp defaultProcessor) PrepareForAnchoring(model Model) error {
	cd, err := model.PackCoreDocument()
	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}

	psv := PostSignatureRequestValidator(dp.identityService)
	err = psv.Validate(nil, model)
	if err != nil {
		return errors.New("failed to validate signatures: %v", err)
	}

	err = coredocument.CalculateDocumentRoot(cd)
	if err != nil {
		return errors.New("failed to generate document root: %v", err)
	}

	err = model.UnpackCoreDocument(cd)
	if err != nil {
		return errors.New("failed to unpack core document: %v", err)
	}

	return nil
}

// AnchorDocument validates the model, and anchors the document
func (dp defaultProcessor) AnchorDocument(ctx context.Context, model Model) error {
	cd, err := model.PackCoreDocument()
	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}

	pav := PreAnchorValidator(dp.identityService)
	err = pav.Validate(nil, model)
	if err != nil {
		return errors.New("pre anchor validation failed: %v", err)
	}

	rootHash, err := anchors.ToDocumentRoot(cd.DocumentRoot)
	if err != nil {
		return errors.New("failed to get document root: %v", err)
	}

	anchorID, err := anchors.ToAnchorID(cd.CurrentVersion)
	if err != nil {
		return errors.New("failed to get anchor ID: %v", err)
	}

	self, err := contextutil.Self(ctx)
	if err != nil {
		return err
	}

	// generate message authentication code for the anchor call
	mac, err := secp256k1.SignEthereum(anchors.GenerateCommitHash(anchorID, self.ID, rootHash), self.Keys[identity.KeyPurposeEthMsgAuth].PrivateKey)
	if err != nil {
		return errors.New("failed to generate ethereum MAC: %v", err)
	}

	log.Infof("Anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", cd.DocumentIdentifier, cd.CurrentVersion, cd.NextVersion, cd.DocumentRoot)
	confirmations, err := dp.anchorRepository.CommitAnchor(ctx, anchorID, rootHash, self.ID, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, mac)
	if err != nil {
		return errors.New("failed to commit anchor: %v", err)
	}

	<-confirmations
	log.Infof("Anchored document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", cd.DocumentIdentifier, cd.CurrentVersion, cd.NextVersion, cd.DocumentRoot)
	return nil
}

// SendDocument does post anchor validations and sends the document to collaborators
func (dp defaultProcessor) SendDocument(ctx context.Context, model Model) error {
	cd, err := model.PackCoreDocument()
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

	extCollaborators, err := coredocument.GetExternalCollaborators(self.ID, cd)
	if err != nil {
		return errors.New("get external collaborators failed: %v", err)
	}

	for _, c := range extCollaborators {
		cID, erri := identity.ToCentID(c)
		if erri != nil {
			err = errors.AppendError(err, erri)
			continue
		}

		erri = dp.Send(ctx, cd, cID)
		if erri != nil {
			err = errors.AppendError(err, erri)
		}
	}

	return err
}
