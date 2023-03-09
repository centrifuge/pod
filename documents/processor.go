package documents

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/pallets/anchors"
	"github.com/centrifuge/pod/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
)

//go:generate mockery --name Client --structname ClientMock --filename client_mock.go --inpackage

// Client defines methods that can be implemented by any type handling p2p communications.
type Client interface {
	// GetSignaturesForDocument gets the signatures for document
	GetSignaturesForDocument(ctx context.Context, model Document) ([]*coredocumentpb.Signature, []error, error)

	// SendAnchoredDocument after all signatures are collected the sender sends the document including the signatures
	SendAnchoredDocument(ctx context.Context, receiverID *types.AccountID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error)

	// GetDocumentRequest requests a document from a collaborator
	GetDocumentRequest(ctx context.Context, documentOwner *types.AccountID, in *p2ppb.GetDocumentRequest) (*p2ppb.GetDocumentResponse, error)
}

//go:generate mockery --name AnchorProcessor --structname AnchorProcessorMock --filename anchor_processor_mock.go --inpackage

type AnchorProcessor interface {
	Send(ctx context.Context, cd *coredocumentpb.CoreDocument, recipient *types.AccountID) (err error)
	PrepareForSignatureRequests(ctx context.Context, doc Document) error
	RequestSignatures(ctx context.Context, doc Document) error
	PrepareForAnchoring(ctx context.Context, doc Document) error
	PreAnchorDocument(ctx context.Context, doc Document) error
	AnchorDocument(ctx context.Context, doc Document) error
	SendDocument(ctx context.Context, doc Document) error
	RequestDocumentWithAccessToken(ctx context.Context, granterDID *types.AccountID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier []byte) (*p2ppb.GetDocumentResponse, error)
}

type anchorProcessor struct {
	p2pClient       Client
	anchorSrv       anchors.API
	config          config.Configuration
	identityService v2.Service
}

func NewAnchorProcessor(
	p2pClient Client,
	anchorSrv anchors.API,
	config config.Configuration,
	identityService v2.Service,
) AnchorProcessor {
	return &anchorProcessor{
		p2pClient:       p2pClient,
		anchorSrv:       anchorSrv,
		config:          config,
		identityService: identityService,
	}
}

// Send sends the given anchorProcessor to the given recipient on the P2P layer
func (ap *anchorProcessor) Send(ctx context.Context, cd *coredocumentpb.CoreDocument, id *types.AccountID) (err error) {
	log.Infof("sending document %s to recipient %s", hexutil.Encode(cd.DocumentIdentifier), id.ToHexString())

	ctx, cancel := context.WithTimeout(ctx, ap.config.GetP2PConnectionTimeout())
	defer cancel()

	resp, err := ap.p2pClient.SendAnchoredDocument(ctx, id, &p2ppb.AnchorDocumentRequest{Document: cd})

	if err != nil || !resp.Accepted {
		log.Errorf("Couldn't send document to the node: %s", err)

		return ErrP2PDocumentSend
	}

	log.Infof("Sent document to %s\n", id.ToHexString())
	return nil
}

// PrepareForSignatureRequests gets the core document from the model, and adds the node's own signature
func (ap *anchorProcessor) PrepareForSignatureRequests(ctx context.Context, model Document) error {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		log.Errorf("Couldn't retrieve account from context: %s", err)

		return errors.ErrContextAccountRetrieval
	}

	id := acc.GetIdentity()

	model.AddUpdateLog(id)

	// execute compute fields
	err = model.ExecuteComputeFields(computeFieldsTimeout)
	if err != nil {
		log.Errorf("Couldn't execute compute fields: %s", err)

		return ErrDocumentExecuteComputeFields
	}

	// calculate the signing root
	sr, err := model.CalculateSigningRoot()
	if err != nil {
		log.Errorf("Couldn't calculate signing root: %s", err)

		return ErrDocumentCalculateSigningRoot
	}

	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, false))
	if err != nil {
		log.Errorf("Couldn't calculate signing root: %s", err)

		return ErrAccountSignMessage
	}

	model.AppendSignatures(sig)

	return nil
}

// RequestSignatures gets the core document from the model, validates pre signature requirements,
// collects signatures, and validates the signatures,
func (ap *anchorProcessor) RequestSignatures(ctx context.Context, model Document) error {
	psv := SignatureValidator(ap.identityService)

	err := psv.Validate(nil, model)

	if err != nil {
		log.Errorf("Couldn't validate document: %s", err)

		return ErrDocumentValidation
	}

	// we ignore signature collection errors and anchor anyways
	signs, _, err := ap.p2pClient.GetSignaturesForDocument(ctx, model)
	if err != nil {
		log.Errorf("Couldn't get signatures for document: %s", err)

		return ErrDocumentSignaturesRetrieval
	}

	model.AppendSignatures(signs...)
	return nil
}

// PrepareForAnchoring validates the signatures and generates the document root
func (ap *anchorProcessor) PrepareForAnchoring(_ context.Context, model Document) error {
	psv := SignatureValidator(ap.identityService)
	err := psv.Validate(nil, model)
	if err != nil {
		log.Errorf("Couldn't validate document: %s", err)

		return ErrDocumentValidation
	}

	return nil
}

// PreAnchorDocument pre-commits a document
func (ap *anchorProcessor) PreAnchorDocument(ctx context.Context, model Document) error {
	signingRoot, err := model.CalculateSigningRoot()
	if err != nil {
		log.Errorf("Couldn't calculate signing root: %s", err)

		return ErrDocumentCalculateSigningRoot
	}

	anchorID, err := anchors.ToAnchorID(model.CurrentVersion())
	if err != nil {
		log.Errorf("Couldn't get anchor ID: %s", err)

		return ErrAnchorIDCreation
	}

	sRoot, err := anchors.ToDocumentRoot(signingRoot)
	if err != nil {
		log.Errorf("Couldn't get document root: %s", err)

		return ErrDocumentRootCreation
	}

	log.Infof("Pre-anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], signingRoot: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), sRoot)

	err = ap.anchorSrv.PreCommitAnchor(ctx, anchorID, sRoot)

	if err != nil {
		log.Errorf("Couldn't pre-commit anchor: %s", err)

		return ErrPreCommitAnchor
	}

	log.Infof("Pre-anchored document with identifiers: [document: %#x, current: %#x, next: %#x], signingRoot: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), sRoot)

	return nil
}

// AnchorDocument validates the model, and anchors the document
func (ap *anchorProcessor) AnchorDocument(ctx context.Context, model Document) error {
	pav := PreAnchorValidator(ap.identityService)
	err := pav.Validate(nil, model)
	if err != nil {
		log.Errorf("Couldn't validate document: %s", err)

		return ErrDocumentValidation
	}

	dr, err := model.CalculateDocumentRoot()
	if err != nil {
		log.Errorf("Couldn't calculate document root: %s", err)

		return ErrDocumentCalculateDocumentRoot
	}

	rootHash, err := anchors.ToDocumentRoot(dr)
	if err != nil {
		log.Errorf("Couldn't create document root: %s", err)

		return ErrDocumentRootCreation
	}

	anchorIDPreimage, err := anchors.ToAnchorID(model.CurrentVersionPreimage())
	if err != nil {
		log.Errorf("Couldn't create anchor ID for pre-image: %s", err)

		return ErrAnchorIDCreation
	}

	signaturesRootProof, err := model.CalculateSignaturesRoot()
	if err != nil {
		log.Errorf("Couldn't create anchor ID for pre-image: %s", err)

		return ErrDocumentCalculateSignaturesRoot
	}

	signaturesRootHash, err := utils.SliceToByte32(signaturesRootProof)
	if err != nil {
		log.Errorf("Couldn't convert signatures root proof: %s", err)

		return ErrSignaturesRootProofConversion
	}

	log.Infof("Anchoring document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), dr)

	err = ap.anchorSrv.CommitAnchor(ctx, anchorIDPreimage, rootHash, signaturesRootHash)

	if err != nil {
		log.Errorf("Couldn't commit anchor: %s", err)

		return ErrCommitAnchor
	}

	log.Infof("Anchored document with identifiers: [document: %#x, current: %#x, next: %#x], rootHash: %#x", model.ID(), model.CurrentVersion(), model.NextVersion(), dr)

	return nil
}

// RequestDocumentWithAccessToken requests a document with an access token
func (ap *anchorProcessor) RequestDocumentWithAccessToken(
	ctx context.Context,
	granterAccountID *types.AccountID,
	tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier []byte,
) (*p2ppb.GetDocumentResponse, error) {
	accessTokenRequest := &p2ppb.AccessTokenRequest{
		DelegatingDocumentIdentifier: delegatingDocumentIdentifier,
		AccessTokenId:                tokenIdentifier,
	}

	request := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: documentIdentifier,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: accessTokenRequest,
	}

	response, err := ap.p2pClient.GetDocumentRequest(ctx, granterAccountID, request)
	if err != nil {
		log.Errorf("Couldn't get document: %s", err)

		return nil, ErrP2PDocumentRetrieval
	}

	return response, nil
}

// SendDocument does post anchor validations and sends the document to collaborators
func (ap *anchorProcessor) SendDocument(ctx context.Context, model Document) error {
	av := PostAnchoredValidator(ap.identityService, ap.anchorSrv)
	err := av.Validate(nil, model)
	if err != nil {
		log.Errorf("Couldn't validate document: %s", err)

		return ErrDocumentValidation
	}

	selfIdentity, err := contextutil.Identity(ctx)
	if err != nil {
		log.Errorf("Couldn't get identity from context: %s", err)

		return errors.ErrContextIdentityRetrieval
	}

	cs, err := model.GetSignerCollaborators(selfIdentity)
	if err != nil {
		log.Errorf("Couldn't get document collaborators: %s", err)

		return ErrDocumentCollaboratorsRetrieval
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		log.Errorf("Couldn't pack core document: %s", err)

		return ErrDocumentPackingCoreDocument
	}

	for _, c := range cs {
		doc := proto.Clone(cd).(*coredocumentpb.CoreDocument)

		// TODO(cdamian): Why not propagate this error?
		err := ap.Send(ctx, doc, c)

		if err != nil {
			log.Errorf("Couldn't send document: %s", err)
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
