package documents

import (
	"context"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Config defines required methods required for the documents package.
type Config interface {
	GetNetworkID() uint32
	GetIdentityID() ([]byte, error)
	GetP2PConnectionTimeout() time.Duration
	GetContractAddress(contractName config.ContractName) common.Address
}

// DocumentRequestProcessor offers methods to interact with the p2p layer to request documents.
type DocumentRequestProcessor interface {
	RequestDocumentWithAccessToken(ctx context.Context, granterDID identity.DID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier []byte) (*p2ppb.GetDocumentResponse, error)
}

// Client defines methods that can be implemented by any type handling p2p communications.
type Client interface {

	// GetSignaturesForDocument gets the signatures for document
	GetSignaturesForDocument(ctx context.Context, model Model) ([]*coredocumentpb.Signature, []error, error)

	// after all signatures are collected the sender sends the document including the signatures
	SendAnchoredDocument(ctx context.Context, receiverID identity.DID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error)

	// GetDocumentRequest requests a document from a collaborator
	GetDocumentRequest(ctx context.Context, requesterID identity.DID, in *p2ppb.GetDocumentRequest) (*p2ppb.GetDocumentResponse, error)
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
func (dp defaultProcessor) Send(ctx context.Context, cd coredocumentpb.CoreDocument, id identity.DID) (err error) {
	log.Infof("sending document %s to recipient %s", hexutil.Encode(cd.DocumentIdentifier), id.String())
	ctx, cancel := context.WithTimeout(ctx, dp.config.GetP2PConnectionTimeout())
	defer cancel()

	resp, err := dp.p2pClient.SendAnchoredDocument(ctx, id, &p2ppb.AnchorDocumentRequest{Document: &cd})
	if err != nil || !resp.Accepted {
		return errors.New("failed to send document to the node: %v", err)
	}

	log.Infof("Sent document to %s\n", id.String())
	return nil
}

// PrepareForSignatureRequests gets the core document from the model, and adds the node's own signature
func (dp defaultProcessor) PrepareForSignatureRequests(ctx context.Context, model Model) error {
	self, err := contextutil.Account(ctx)
	if err != nil {
		return err
	}

	_, err = model.CalculateDataRoot()
	if err != nil {
		return err
	}

	id := self.GetIdentityID()
	did, err := identity.NewDIDFromBytes(id)
	if err != nil {
		return err
	}

	err = model.AddUpdateLog(did)
	if err != nil {
		return err
	}

	addr := dp.config.GetContractAddress(config.AnchorRepo)
	model.SetUsedAnchorRepoAddress(addr)

	// calculate the signing root
	sr, err := model.CalculateSigningRoot()
	if err != nil {
		return errors.New("failed to calculate signing root: %v", err)
	}

	sigs, err := self.SignMsg(ConsensusSignaturePayload(sr, byte(0)))
	if err != nil {
		return err
	}

	model.AppendSignatures(sigs...)
	return nil
}

// RequestSignatures gets the core document from the model, validates pre signature requirements,
// collects signatures, and validates the signatures,
func (dp defaultProcessor) RequestSignatures(ctx context.Context, model Model) error {
	psv := SignatureValidator(dp.identityService, dp.anchorRepository)
	err := psv.Validate(nil, model)
	if err != nil {
		return errors.New("failed to validate model for signature request: %v", err)
	}

	// we ignore signature collection errors and anchor anyways
	signs, _, err := dp.p2pClient.GetSignaturesForDocument(ctx, model)
	if err != nil {
		return errors.New("failed to collect signatures from the collaborators: %v", err)
	}

	model.AppendSignatures(signs...)
	return nil
}

// PrepareForAnchoring validates the signatures and generates the document root
func (dp defaultProcessor) PrepareForAnchoring(model Model) error {
	psv := SignatureValidator(dp.identityService, dp.anchorRepository)
	err := psv.Validate(nil, model)
	if err != nil {
		return errors.New("failed to validate signatures: %v", err)
	}

	return nil
}

// PreAnchorDocument pre-commits a document
func (dp defaultProcessor) PreAnchorDocument(ctx context.Context, model Model) error {
	signingRoot, err := model.CalculateSigningRoot()
	if err != nil {
		return err
	}

	anchorID, err := anchors.ToAnchorID(model.CurrentVersion())
	if err != nil {
		return err
	}

	sRoot, err := anchors.ToDocumentRoot(signingRoot)
	if err != nil {
		return err
	}

	log.Infof("Pre-anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], signingRoot: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), sRoot)
	done, err := dp.anchorRepository.PreCommitAnchor(ctx, anchorID, sRoot)
	if err != nil {
		return err
	}

	err = <-done
	if err != nil {
		return errors.New("failed to pre-commit anchor: %v", err)
	}

	log.Infof("Pre-anchored document with identifiers: [document: %#x, current: %#x, next: %#x], signingRoot: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), sRoot)
	return nil
}

// AnchorDocument validates the model, and anchors the document
func (dp defaultProcessor) AnchorDocument(ctx context.Context, model Model) error {
	pav := PreAnchorValidator(dp.identityService, dp.anchorRepository)
	err := pav.Validate(nil, model)
	if err != nil {
		return errors.New("pre anchor validation failed: %v", err)
	}

	dr, err := model.CalculateDocumentRoot()
	if err != nil {
		return errors.New("failed to get document root: %v", err)
	}

	rootHash, err := anchors.ToDocumentRoot(dr)
	if err != nil {
		return errors.New("failed to convert document root: %v", err)
	}

	anchorIDPreimage, err := anchors.ToAnchorID(model.CurrentVersionPreimage())
	if err != nil {
		return errors.New("failed to get anchor ID: %v", err)
	}

	signingRootProof, err := model.CalculateSignaturesRoot()
	if err != nil {
		return errors.New("failed to get signature root: %v", err)
	}

	signingRootHash, err := utils.SliceToByte32(signingRootProof)
	if err != nil {
		return errors.New("failed to get signing root proof in ethereum format: %v", err)
	}

	log.Infof("Anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), dr)
	done, err := dp.anchorRepository.CommitAnchor(ctx, anchorIDPreimage, rootHash, signingRootHash)
	if err != nil {
		return errors.New("failed to commit anchor: %v", err)
	}

	err = <-done
	if err != nil {
		return errors.New("failed to commit anchor: %v", err)
	}

	log.Infof("Anchored document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), dr)
	return nil
}

// RequestDocumentWithAccessToken requests a document with an access token
func (dp defaultProcessor) RequestDocumentWithAccessToken(ctx context.Context, granterDID identity.DID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier []byte) (*p2ppb.GetDocumentResponse, error) {
	accessTokenRequest := &p2ppb.AccessTokenRequest{DelegatingDocumentIdentifier: delegatingDocumentIdentifier, AccessTokenId: tokenIdentifier}

	request := &p2ppb.GetDocumentRequest{DocumentIdentifier: documentIdentifier,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: accessTokenRequest,
	}

	response, err := dp.p2pClient.GetDocumentRequest(ctx, granterDID, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// SendDocument does post anchor validations and sends the document to collaborators
func (dp defaultProcessor) SendDocument(ctx context.Context, model Model) error {
	av := PostAnchoredValidator(dp.identityService, dp.anchorRepository)
	err := av.Validate(nil, model)
	if err != nil {
		return errors.New("post anchor validations failed: %v", err)
	}

	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return err
	}

	cs, err := model.GetSignerCollaborators(selfDID)
	if err != nil {
		return errors.New("get external collaborators failed: %v", err)
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}

	for _, c := range cs {
		erri := dp.Send(ctx, cd, c)
		if erri != nil {
			err = errors.AppendError(err, erri)
		}
	}

	return err
}

// ConsensusSignaturePayload forms the payload needed to be signed during the document consensus flow
func ConsensusSignaturePayload(dataRoot []byte, validationFlag byte) []byte {
	return append(dataRoot, []byte{validationFlag}...)
}