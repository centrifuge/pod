package documents

import (
	"context"

	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	v2 "github.com/centrifuge/go-centrifuge/identity/v2"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// AnchorProcessor identifies an implementation, which can do a bunch of things with a CoreDocument.
// E.g. send, anchor, etc.
type AnchorProcessor interface {
	Send(ctx context.Context, cd *coredocumentpb.CoreDocument, recipient *types.AccountID) (err error)
	PrepareForSignatureRequests(ctx context.Context, doc Document) error
	RequestSignatures(ctx context.Context, doc Document) error
	PrepareForAnchoring(ctx context.Context, doc Document) error
	PreAnchorDocument(ctx context.Context, doc Document) error
	AnchorDocument(ctx context.Context, doc Document) error
	SendDocument(ctx context.Context, doc Document) error
}

// DocumentRequestProcessor offers methods to interact with the p2p layer to request documents.
type DocumentRequestProcessor interface {
	RequestDocumentWithAccessToken(ctx context.Context, granterDID *types.AccountID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier []byte) (*p2ppb.GetDocumentResponse, error)
}

// Client defines methods that can be implemented by any type handling p2p communications.
type Client interface {

	// GetSignaturesForDocument gets the signatures for document
	GetSignaturesForDocument(ctx context.Context, model Document) ([]*coredocumentpb.Signature, []error, error)

	// after all signatures are collected the sender sends the document including the signatures
	SendAnchoredDocument(ctx context.Context, receiverID *types.AccountID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error)

	// GetDocumentRequest requests a document from a collaborator
	GetDocumentRequest(ctx context.Context, requesterID *types.AccountID, in *p2ppb.GetDocumentRequest) (*p2ppb.GetDocumentResponse, error)
}

// defaultProcessor implements AnchorProcessor interface
type defaultProcessor struct {
	p2pClient       Client
	anchorSrv       anchors.Service
	config          config.Configuration
	identityService v2.Service
}

// DefaultProcessor returns the default implementation of CoreDocument AnchorProcessor
func DefaultProcessor(
	p2pClient Client,
	anchorSrv anchors.Service,
	config config.Configuration,
	identityService v2.Service,
) AnchorProcessor {
	return defaultProcessor{
		p2pClient:       p2pClient,
		anchorSrv:       anchorSrv,
		config:          config,
		identityService: identityService,
	}
}

// Send sends the given defaultProcessor to the given recipient on the P2P layer
func (dp defaultProcessor) Send(ctx context.Context, cd *coredocumentpb.CoreDocument, id *types.AccountID) (err error) {
	log.Infof("sending document %s to recipient %s", hexutil.Encode(cd.DocumentIdentifier), id.ToHexString())
	ctx, cancel := context.WithTimeout(ctx, dp.config.GetP2PConnectionTimeout())
	defer cancel()

	resp, err := dp.p2pClient.SendAnchoredDocument(ctx, id, &p2ppb.AnchorDocumentRequest{Document: cd})
	if err != nil || !resp.Accepted {
		return errors.New("failed to send document to the node: %v", err)
	}

	log.Infof("Sent document to %s\n", id.ToHexString())
	return nil
}

// PrepareForSignatureRequests gets the core document from the model, and adds the node's own signature
func (dp defaultProcessor) PrepareForSignatureRequests(ctx context.Context, model Document) error {
	self, err := contextutil.Account(ctx)
	if err != nil {
		return err
	}

	id := self.GetIdentity()

	err = model.AddUpdateLog(id)
	if err != nil {
		return err
	}

	// TODO(cdamian): Remove?
	//addr := dp.config.GetContractAddress(config.AnchorRepo)
	//model.SetUsedAnchorRepoAddress(addr)

	// execute compute fields
	err = model.ExecuteComputeFields(computeFieldsTimeout)
	if err != nil {
		return err
	}

	// calculate the signing root
	sr, err := model.CalculateSigningRoot()
	if err != nil {
		return errors.New("failed to calculate signing root: %v", err)
	}

	sig, err := self.SignMsg(ConsensusSignaturePayload(sr, false))
	if err != nil {
		return err
	}

	model.AppendSignatures(sig)
	return nil
}

// RequestSignatures gets the core document from the model, validates pre signature requirements,
// collects signatures, and validates the signatures,
func (dp defaultProcessor) RequestSignatures(ctx context.Context, model Document) error {
	psv := SignatureValidator(dp.identityService)
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
func (dp defaultProcessor) PrepareForAnchoring(ctx context.Context, model Document) error {
	psv := SignatureValidator(dp.identityService)
	err := psv.Validate(nil, model)
	if err != nil {
		return errors.New("failed to validate signatures: %v", err)
	}

	return nil
}

// PreAnchorDocument pre-commits a document
func (dp defaultProcessor) PreAnchorDocument(ctx context.Context, model Document) error {
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
	err = dp.anchorSrv.PreCommitAnchor(ctx, anchorID, sRoot)
	if err != nil {
		return errors.New("failed to pre-commit anchor: %v", err)
	}

	log.Infof("Pre-anchored document with identifiers: [document: %#x, current: %#x, next: %#x], signingRoot: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), sRoot)
	return nil
}

// AnchorDocument validates the model, and anchors the document
func (dp defaultProcessor) AnchorDocument(ctx context.Context, model Document) error {
	pav := PreAnchorValidator(dp.identityService)
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

	signaturesRootProof, err := model.CalculateSignaturesRoot()
	if err != nil {
		return errors.New("failed to get signature root: %v", err)
	}

	signaturesRootHash, err := utils.SliceToByte32(signaturesRootProof)
	if err != nil {
		return errors.New("failed to get signing root proof in ethereum format: %v", err)
	}

	log.Infof("Anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), dr)
	err = dp.anchorSrv.CommitAnchor(ctx, anchorIDPreimage, rootHash, signaturesRootHash)
	if err != nil {
		return errors.New("failed to commit anchor: %v", err)
	}

	log.Infof("Anchored document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), dr)
	return nil
}

// RequestDocumentWithAccessToken requests a document with an access token
func (dp defaultProcessor) RequestDocumentWithAccessToken(ctx context.Context, granterAccountID *types.AccountID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier []byte) (*p2ppb.GetDocumentResponse, error) {
	accessTokenRequest := &p2ppb.AccessTokenRequest{DelegatingDocumentIdentifier: delegatingDocumentIdentifier, AccessTokenId: tokenIdentifier}

	request := &p2ppb.GetDocumentRequest{DocumentIdentifier: documentIdentifier,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: accessTokenRequest,
	}

	response, err := dp.p2pClient.GetDocumentRequest(ctx, granterAccountID, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// SendDocument does post anchor validations and sends the document to collaborators
func (dp defaultProcessor) SendDocument(ctx context.Context, model Document) error {
	av := PostAnchoredValidator(dp.identityService, dp.anchorSrv)
	err := av.Validate(nil, model)
	if err != nil {
		return errors.New("post anchor validations failed: %v", err)
	}

	selfIdentity, err := contextutil.Identity(ctx)
	if err != nil {
		return err
	}

	cs, err := model.GetSignerCollaborators(selfIdentity)
	if err != nil {
		return errors.New("get external collaborators failed: %v", err)
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return errors.New("failed to pack core document: %v", err)
	}

	for _, c := range cs {
		doc := proto.Clone(cd).(*coredocumentpb.CoreDocument)

		err := dp.Send(ctx, doc, c)

		if err != nil {
			log.Error(err)
		}
	}

	return err
}

// ConsensusSignaturePayload forms the payload needed to be signed during the document consensus flow
func ConsensusSignaturePayload(dataRoot []byte, validated bool) []byte {
	tFlag := byte(0)
	if validated {
		tFlag = byte(1)
	}
	return append(dataRoot, tFlag)
}
