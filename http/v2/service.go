package v2

import (
	"context"
	"fmt"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
)

// Service is the entry point for all the V2 APIs.
type Service struct {
	pendingDocSrv   pending.Service
	dispatcher      jobs.Dispatcher
	cfgService      config.Service
	entitySrv       entity.Service
	identityService v2.Service
	erSrv           entityrelationship.Service
	docSrv          documents.Service

	p2pPublicKey         []byte
	podOperatorAccountID *types.AccountID
}

func NewService(
	pendingDocSrv pending.Service,
	dispatcher jobs.Dispatcher,
	cfgService config.Service,
	entitySrv entity.Service,
	identityService v2.Service,
	erSrv entityrelationship.Service,
	docSrv documents.Service,
) (*Service, error) {
	p2pPublicKey, err := getP2PPublicKey(cfgService)

	if err != nil {
		return nil, err
	}

	podOperatorAccountID, err := getPodOperatorAccountID(cfgService)

	if err != nil {
		return nil, err
	}

	return &Service{
		pendingDocSrv:        pendingDocSrv,
		dispatcher:           dispatcher,
		cfgService:           cfgService,
		entitySrv:            entitySrv,
		erSrv:                erSrv,
		docSrv:               docSrv,
		identityService:      identityService,
		p2pPublicKey:         p2pPublicKey,
		podOperatorAccountID: podOperatorAccountID,
	}, nil
}

// CreateDocument creates a pending document from the given payload.
// if the document_id is provided, next version of the document is created.
func (s *Service) CreateDocument(ctx context.Context, req documents.UpdatePayload) (documents.Document, error) {
	return s.pendingDocSrv.Create(ctx, req)
}

// CloneDocument creates a new cloned document from the template (docID specified in payload).
func (s *Service) CloneDocument(ctx context.Context, payload documents.ClonePayload) (documents.Document, error) {
	return s.pendingDocSrv.Clone(ctx, payload)
}

// UpdateDocument updates a pending document with the given payload
func (s *Service) UpdateDocument(ctx context.Context, req documents.UpdatePayload) (documents.Document, error) {
	return s.pendingDocSrv.Update(ctx, req)
}

// Commit creates a document out of a pending document.
func (s *Service) Commit(ctx context.Context, docID []byte) (documents.Document, gocelery.JobID, error) {
	return s.pendingDocSrv.Commit(ctx, docID)
}

// GetDocument returns the document associated with docID and status.
func (s *Service) GetDocument(ctx context.Context, docID []byte, status documents.Status) (documents.Document, error) {
	return s.pendingDocSrv.Get(ctx, docID, status)
}

// GetDocumentVersion returns the specific version of the document.
func (s *Service) GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Document, error) {
	return s.pendingDocSrv.GetVersion(ctx, docID, versionID)
}

// AddSignedAttribute signs the payload with acc signing key and add it the document associated with docID.
func (s *Service) AddSignedAttribute(ctx context.Context, docID []byte, label string, payload []byte, valType documents.AttributeType) (documents.Document, error) {
	return s.pendingDocSrv.AddSignedAttribute(ctx, docID, label, payload, valType)
}

// RemoveCollaborators removes collaborators from the document.
func (s *Service) RemoveCollaborators(ctx context.Context, docID []byte, dids []*types.AccountID) (documents.Document, error) {
	return s.pendingDocSrv.RemoveCollaborators(ctx, docID, dids)
}

// AddRole adds a new role to the document
func (s *Service) AddRole(ctx context.Context, docID []byte, roleKey string, dids []*types.AccountID) (*coredocumentpb.Role, error) {
	return s.pendingDocSrv.AddRole(ctx, docID, roleKey, dids)
}

// GetRole gets the role from the document
func (s *Service) GetRole(ctx context.Context, docID, roleID []byte) (*coredocumentpb.Role, error) {
	return s.pendingDocSrv.GetRole(ctx, docID, roleID)
}

// UpdateRole updates the role in the document
func (s *Service) UpdateRole(ctx context.Context, docID, roleID []byte, dids []*types.AccountID) (*coredocumentpb.Role, error) {
	return s.pendingDocSrv.UpdateRole(ctx, docID, roleID, dids)
}

// AddTransitionRules adds new rules to the document
func (s *Service) AddTransitionRules(
	ctx context.Context,
	docID []byte,
	addRules pending.AddTransitionRules,
) ([]*coredocumentpb.TransitionRule, error) {
	return s.pendingDocSrv.AddTransitionRules(ctx, docID, addRules)
}

// GetTransitionRule returns the transition rule associated with ruleID in the document.
func (s *Service) GetTransitionRule(ctx context.Context, docID, ruleID []byte) (*coredocumentpb.TransitionRule, error) {
	return s.pendingDocSrv.GetTransitionRule(ctx, docID, ruleID)
}

// DeleteTransitionRule deletes the transition rule associated with ruleID from the document.
func (s *Service) DeleteTransitionRule(ctx context.Context, docID, ruleID []byte) error {
	return s.pendingDocSrv.DeleteTransitionRule(ctx, docID, ruleID)
}

// AddAttributes add attributes to pending document
func (s *Service) AddAttributes(ctx context.Context, docID []byte, attrs []documents.Attribute) (documents.Document, error) {
	return s.pendingDocSrv.AddAttributes(ctx, docID, attrs)
}

// DeleteAttribute deletes attribute on a pending document
func (s *Service) DeleteAttribute(ctx context.Context, docID []byte, key documents.AttrKey) (documents.Document, error) {
	return s.pendingDocSrv.DeleteAttribute(ctx, docID, key)
}

// Job returns the job details
func (s *Service) Job(accID *types.AccountID, jobID []byte) (*gocelery.Job, error) {
	return s.dispatcher.Job(accID, jobID)
}

// GenerateAccount generates a new account
func (s *Service) GenerateAccount(ctx context.Context, req *v2.CreateIdentityRequest) (acc config.Account, err error) {
	return s.identityService.CreateIdentity(ctx, req)
}

// SignPayload uses the accountID's secret key to sign the payload and returns the signature
func (s *Service) SignPayload(accountID, payload []byte) (*coredocumentpb.Signature, error) {
	acc, err := s.cfgService.GetAccount(accountID)

	if err != nil {
		return nil, fmt.Errorf("couldn't retreive account: %w", err)
	}

	return acc.SignMsg(payload)
}

// GetEntityByRelationship returns an entity through a relationship ID.
func (s *Service) GetEntityByRelationship(ctx context.Context, docID []byte) (documents.Document, error) {
	return s.entitySrv.GetEntityByRelationship(ctx, docID)
}

// GetEntityRelationShips returns the entity relationships under the given entity
func (s *Service) GetEntityRelationShips(ctx context.Context, entityID []byte) ([]documents.Document, error) {
	return s.erSrv.GetEntityRelationships(ctx, entityID)
}

// GetAccount returns the Account associated with accountID
func (s *Service) GetAccount(accountID []byte) (config.Account, error) {
	return s.cfgService.GetAccount(accountID)
}

// GetAccounts returns all the accounts.
func (s *Service) GetAccounts() ([]config.Account, error) {
	return s.cfgService.GetAccounts()
}

// GenerateProofs returns the proofs for the latest version of the document.
func (s *Service) GenerateProofs(ctx context.Context, docID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docSrv.CreateProofs(ctx, docID, fields)
}

// GenerateProofsForVersion returns the proofs for the specific version of the document.
func (s *Service) GenerateProofsForVersion(ctx context.Context, docID, versionID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docSrv.CreateProofsForVersion(ctx, docID, versionID, fields)
}

func (s *Service) ToClientAccounts(accounts ...config.Account) []coreapi.Account {
	var res []coreapi.Account

	for _, account := range accounts {
		res = append(res, toClientAccount(account, s.p2pPublicKey, s.podOperatorAccountID))
	}

	return res
}

func toClientAccount(account config.Account, p2pPublicKey []byte, podOperatorAccountID *types.AccountID) coreapi.Account {
	return coreapi.Account{
		Identity:                 account.GetIdentity(),
		WebhookURL:               account.GetWebhookURL(),
		PrecommitEnabled:         account.GetPrecommitEnabled(),
		DocumentSigningPublicKey: account.GetSigningPublicKey(),
		P2PPublicSigningKey:      p2pPublicKey,
		PodOperatorAccountID:     podOperatorAccountID,
	}
}

func getP2PPublicKey(cfgService config.Service) ([]byte, error) {
	cfg, err := cfgService.GetConfig()

	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve config: %w", err)
	}

	_, pubKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve P2P public key: %w", err)
	}

	return pubKey.Raw()
}

func getPodOperatorAccountID(cfgService config.Service) (*types.AccountID, error) {
	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator: %w", err)
	}

	return podOperator.GetAccountID(), nil
}
