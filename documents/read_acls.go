package documents

import (
	"bytes"
	"context"
	"fmt"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// initReadRules initiates the read rules for a given CoreDocumentModel.
// Collaborators are given Read_Sign action.
// if the rules are created already, this is a no-op.
// if collaborators are empty, it is a no-op
func (cd *CoreDocument) initReadRules(collaborators []identity.DID) {
	if len(cd.Document.Roles) > 0 && len(cd.Document.ReadRules) > 0 {
		return
	}

	if len(collaborators) < 1 {
		return
	}

	cd.addCollaboratorsToReadSignRules(collaborators)
}

// addCollaboratorsToReadSignRules adds the given collaborators to a new read rule with READ_SIGN capability.
// The operation is no-op if no collaborators are provided.
// The operation is not idempotent. So calling twice with same accounts will lead to read rules duplication.
func (cd *CoreDocument) addCollaboratorsToReadSignRules(collaborators []identity.DID) {
	role := newRoleWithCollaborators(collaborators...)
	if role == nil {
		return
	}
	cd.Document.Roles = append(cd.Document.Roles, role)
	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ_SIGN)
	cd.Modified = true
}

// addNewReadRule creates a new read rule as per the role and action.
func (cd *CoreDocument) addNewReadRule(roleKey []byte, action coredocumentpb.Action) {
	rule := &coredocumentpb.ReadRule{
		Action: action,
		Roles:  [][]byte{roleKey},
	}
	cd.Document.ReadRules = append(cd.Document.ReadRules, rule)
	cd.Modified = true
}

// findRole calls OnRole for every role that matches the actions passed in
func findReadRole(cd coredocumentpb.CoreDocument, onRole func(rridx, ridx int, role *coredocumentpb.Role) bool, actions ...coredocumentpb.Action) bool {
	am := make(map[int32]struct{})
	for _, a := range actions {
		am[int32(a)] = struct{}{}
	}

	for i, rule := range cd.ReadRules {
		if _, ok := am[int32(rule.Action)]; !ok {
			continue
		}

		for j, rk := range rule.Roles {
			role, err := getRole(rk, cd.Roles)
			if err != nil {
				// seems like roles and rules are not in sync
				// skip to next one
				continue
			}

			if onRole(i, j, role) {
				return true
			}

		}
	}

	return false
}

// findTransitionRole calls OnRole for every role that matches the actions passed in
func findTransitionRole(cd coredocumentpb.CoreDocument, onRole func(rridx, ridx int, role *coredocumentpb.Role) bool, actions ...coredocumentpb.TransitionAction) bool {
	am := make(map[int32]struct{})
	for _, a := range actions {
		am[int32(a)] = struct{}{}
	}

	for i, rule := range cd.TransitionRules {
		if _, ok := am[int32(rule.Action)]; !ok {
			continue
		}

		for j, rk := range rule.Roles {
			role, err := getRole(rk, cd.Roles)
			if err != nil {
				// seems like roles and rules are not in sync
				// skip to next one
				continue
			}

			if onRole(i, j, role) {
				return true
			}

		}
	}

	return false
}

// NFTOwnerCanRead checks if the nft owner/account can read the Document
func (cd *CoreDocument) NFTOwnerCanRead(tokenRegistry TokenRegistry, registry common.Address, tokenID []byte, account identity.DID) error {
	// check if the account can read the doc
	if cd.AccountCanRead(account) {
		return nil
	}

	// check if the nft is present in read rules
	found := findReadRole(cd.Document, func(_, _ int, role *coredocumentpb.Role) bool {
		_, found := isNFTInRole(role, registry, tokenID)
		return found
	}, coredocumentpb.Action_ACTION_READ)

	if !found {
		return ErrNftNotFound
	}

	// get the owner of the NFT
	// TODO(ved): this should check the owner on CC once the did is migrated to chain
	owner, err := tokenRegistry.OwnerOf(registry, tokenID)
	if err != nil {
		return errors.New("failed to get NFT owner: %v", err)
	}

	if !bytes.Equal(owner.Bytes(), account[:]) {
		return errors.New("account (%v) not owner of the NFT", account.ToHexString())
	}

	return nil
}

// AccountCanRead validate if the core document can be read by the account .
// Returns an error if not.
func (cd *CoreDocument) AccountCanRead(account identity.DID) bool {
	// loop though read rules, check all the rules
	return findReadRole(cd.Document, func(_, _ int, role *coredocumentpb.Role) bool {
		_, found := isDIDInRole(role, account)
		return found
	}, coredocumentpb.Action_ACTION_READ, coredocumentpb.Action_ACTION_READ_SIGN)
}

// addNFTToReadRules adds NFT token to the read rules of core document.
func (cd *CoreDocument) addNFTToReadRules(registry common.Address, tokenID []byte) error {
	nft, err := ConstructNFT(registry, tokenID)
	if err != nil {
		return errors.New("failed to construct NFT: %v", err)
	}

	role := newRoleWithRandomKey()
	role.Nfts = append(role.Nfts, nft)
	cd.Document.Roles = append(cd.Document.Roles, role)
	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ)
	cd.Modified = true
	return nil
}

// AddNFT returns a new CoreDocument model with nft added to the Core document. If grantReadAccess is true, the nft is added
// to the read rules.
func (cd *CoreDocument) AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte, pad bool) (*CoreDocument,
	error) {
	ncd, err := cd.PrepareNewVersion(nil, CollaboratorsAccess{}, nil)
	if err != nil {
		return nil, errors.New("failed to prepare new version: %v", err)
	}

	nft := getStoredNFT(ncd.Document.Nfts, registry.Bytes())
	if nft == nil {
		nft = new(coredocumentpb.NFT)
		nft.RegistryId = registry.Bytes()
		if pad {
			// add 12 empty bytes
			eb := make([]byte, 12)
			nft.RegistryId = append(registry.Bytes(), eb...)
		}
		ncd.Document.Nfts = append(ncd.Document.Nfts, nft)
	}
	nft.TokenId = tokenID

	if grantReadAccess {
		err = ncd.addNFTToReadRules(registry, tokenID)
		if err != nil {
			return nil, err
		}
	}

	cd.Modified = true
	return ncd, nil
}

// NFTs returns the list of NFTs created for this model
func (cd *CoreDocument) NFTs() []*coredocumentpb.NFT {
	return cd.Document.Nfts
}

// IsNFTMinted checks if the there is an NFT that is minted against this document in the given registry.
func (cd *CoreDocument) IsNFTMinted(tokenRegistry TokenRegistry, registry common.Address) bool {
	nft := getStoredNFT(cd.Document.Nfts, registry.Bytes())
	if nft == nil {
		return false
	}

	_, err := tokenRegistry.OwnerOf(registry, nft.TokenId)
	return err == nil
}

// CreateNFTProofs generate proofs returns proofs for NFT minting.
func (cd *CoreDocument) CreateNFTProofs(
	docType string,
	dataLeaves []proofs.LeafNode,
	account identity.DID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (prf *DocumentProof, err error) {

	var pfKeys []string
	if nftUniqueProof {
		pk, err := getNFTUniqueProofKey(cd.Document.Nfts, registry)
		if err != nil {
			return nil, err
		}

		pfKeys = append(pfKeys, pk)
	}

	if readAccessProof {
		pks, err := getReadAccessProofKeys(cd.Document, registry, tokenID)
		if err != nil {
			return nil, err
		}

		pfKeys = append(pfKeys, pks...)
	}
	return cd.CreateProofs(docType, dataLeaves, pfKeys)
}

// ConstructNFT appends registry and tokenID to byte slice
func ConstructNFT(registry common.Address, tokenID []byte) ([]byte, error) {
	var nft []byte
	// first 20 bytes of registry
	nft = append(nft, registry.Bytes()...)

	// next 32 bytes of the tokenID
	nft = append(nft, tokenID...)

	if len(nft) != nftByteCount {
		return nil, errors.NewTypedError(ErrNftByteLength, errors.New("provided length %d", len(nft)))
	}

	return nft, nil
}

// isNFTInRole checks if the given nft(registry + token) is part of the core document role.
// If found, returns the index of the nft in the role and true
func isNFTInRole(role *coredocumentpb.Role, registry common.Address, tokenID []byte) (nftIdx int, found bool) {
	enft, err := ConstructNFT(registry, tokenID)
	if err != nil {
		return nftIdx, false
	}

	for i, n := range role.Nfts {
		if bytes.Equal(n, enft) {
			return i, true
		}
	}

	return nftIdx, false
}

func getStoredNFT(nfts []*coredocumentpb.NFT, registry []byte) *coredocumentpb.NFT {
	for _, nft := range nfts {
		if bytes.Equal(nft.RegistryId[:common.AddressLength], registry) {
			return nft
		}
	}

	return nil
}

func getReadAccessProofKeys(cd coredocumentpb.CoreDocument, registry common.Address, tokenID []byte) (pks []string, err error) {
	var rridx int  // index of the read rules which contain the role
	var ridx int   // index of the role
	var nftIdx int // index of the NFT in the above role
	var rk []byte  // role key of the above role

	found := findReadRole(cd, func(i, j int, role *coredocumentpb.Role) bool {
		z, found := isNFTInRole(role, registry, tokenID)
		if found {
			rridx = i
			ridx = j
			rk = role.RoleKey
			nftIdx = z
		}

		return found
	}, coredocumentpb.Action_ACTION_READ)

	if !found {
		return nil, ErrNFTRoleMissing
	}

	return []string{
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].roles[%d]", rridx, ridx),          // proof that a read rule exists with the nft role
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].action", rridx),                   // proof that this read rule has read access
		fmt.Sprintf(CDTreePrefix+".roles[%s].nfts[%d]", hexutil.Encode(rk), nftIdx), // proof that role with nft exists
	}, nil
}

func getNFTUniqueProofKey(nfts []*coredocumentpb.NFT, registry common.Address) (pk string, err error) {
	nft := getStoredNFT(nfts, registry.Bytes())
	if nft == nil {
		return pk, ErrNftNotFound
	}

	key := hexutil.Encode(nft.RegistryId)
	return fmt.Sprintf(CDTreePrefix+".nfts[%s]", key), nil
}

// isDIDInRole returns the index of the collaborator and true if did is in the given role as collaborators.
func isDIDInRole(role *coredocumentpb.Role, did identity.DID) (idx int, found bool) {
	for i, id := range role.Collaborators {
		if bytes.Equal(id, did[:]) {
			return i, true
		}
	}

	return idx, false
}

func getRole(key []byte, roles []*coredocumentpb.Role) (*coredocumentpb.Role, error) {
	for _, role := range roles {
		if utils.IsSameByteSlice(role.RoleKey, key) {
			return role, nil
		}
	}

	return nil, errors.New("role %d not found", key)
}

// validateAT validates that given access token against its signature
func validateAT(publicKey []byte, token *coredocumentpb.AccessToken, requesterID []byte) error {
	// assemble token message from the token for validation
	reqID, err := identity.NewDIDFromBytes(requesterID)
	if err != nil {
		return err
	}
	granterID, err := identity.NewDIDFromBytes(token.Granter)
	if err != nil {
		return err
	}
	tm, err := assembleTokenMessage(token.Identifier, granterID, reqID, token.RoleIdentifier, token.DocumentIdentifier, token.DocumentVersion)
	if err != nil {
		return err
	}
	validated := crypto.VerifyMessage(publicKey, tm, token.Signature, crypto.CurveEd25519)
	if !validated {
		return ErrAccessTokenInvalid
	}
	return nil
}

func (cd *CoreDocument) findAT(tokenID []byte) (at *coredocumentpb.AccessToken, err error) {
	// check if the access token is present on the document indicated in the AT request
	for _, at := range cd.Document.AccessTokens {
		if bytes.Equal(tokenID, at.Identifier) {
			return at, nil
		}
	}
	return at, ErrAccessTokenNotFound
}

// ATGranteeCanRead checks that the grantee of the access token can read the document requested
func (cd *CoreDocument) ATGranteeCanRead(ctx context.Context, docService Service, idService identity.Service, tokenID, docID []byte, requesterID identity.DID) (err error) {
	// find the access token
	at, err := cd.findAT(tokenID)
	if err != nil {
		return err
	}
	granterID, err := identity.NewDIDFromBytes(at.Granter)
	if err != nil {
		return err
	}
	granteeID, err := identity.NewDIDFromBytes(at.Grantee)
	if err != nil {
		return err
	}
	// check that the peer requesting access is the same identity as the access token grantee
	if !requesterID.Equal(granteeID) {
		return ErrRequesterNotGrantee
	}
	// check that the granter of the access token is a collaborator on the document
	verified := cd.AccountCanRead(granterID)
	if !verified {
		return ErrGranterNotCollab
	}
	// check if the requested document is the document indicated in the access token
	if !bytes.Equal(at.DocumentIdentifier, docID) {
		return ErrReqDocNotMatch
	}
	// validate that the public key of the granter is the public key that has been used to sign the access token
	doc, err := docService.GetVersion(ctx, cd.Document.DocumentIdentifier, at.DocumentVersion)
	if err != nil {
		return err
	}
	ts, err := doc.Timestamp()
	if err != nil {
		return err
	}
	err = idService.ValidateKey(ctx, granterID, at.Key, &(identity.KeyPurposeSigning.Value), &ts)
	if err != nil {
		return err
	}
	return validateAT(at.Key, at, granteeID[:])
}

// AddAccessToken adds the AccessToken to the document
func (cd *CoreDocument) AddAccessToken(ctx context.Context, payload AccessTokenParams) (*CoreDocument, error) {
	ncd, err := cd.PrepareNewVersion(nil, CollaboratorsAccess{}, nil)
	if err != nil {
		return nil, err
	}

	at, err := assembleAccessToken(ctx, payload, cd.CurrentVersion())
	if err != nil {
		return nil, errors.New("failed to construct access token: %v", err)
	}

	ncd.Document.AccessTokens = append(ncd.Document.AccessTokens, at)
	ncd.Modified = true
	return ncd, nil
}

// DeleteAccessToken deletes an access token on the Document
func (cd *CoreDocument) DeleteAccessToken(granteeID identity.DID) (*CoreDocument, error) {
	ncd, err := cd.PrepareNewVersion(nil, CollaboratorsAccess{}, nil)
	if err != nil {
		return nil, err
	}

	accessTokens := ncd.Document.AccessTokens
	for i, t := range accessTokens {
		if bytes.Equal(t.Grantee, granteeID[:]) {
			ncd.Document.AccessTokens = removeTokenAtIndex(i, accessTokens)
			ncd.Modified = true
			return ncd, nil
		}
	}
	return nil, ErrAccessTokenNotFound
}

// RemoveTokenAtIndex removes the access token at index i from slice a
// Note: changes the order of the slice elements
func removeTokenAtIndex(idx int, tokens []*coredocumentpb.AccessToken) []*coredocumentpb.AccessToken {
	tokens[len(tokens)-1], tokens[idx] = tokens[idx], tokens[len(tokens)-1]
	return tokens[:len(tokens)-1]
}

// assembleAccessToken assembles a Read Access Token from the payload received
func assembleAccessToken(ctx context.Context, payload AccessTokenParams, docVersion []byte) (*coredocumentpb.AccessToken, error) {
	account, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}
	tokenIdentifier := utils.RandomSlice(32)
	granterID := account.GetIdentity()

	// TODO: this roleID will be specified later with field level read access
	roleID := utils.RandomSlice(32)
	granteeID, err := identity.NewDIDFromString(payload.Grantee)
	if err != nil {
		return nil, err
	}
	// assemble access token message to be signed
	docID, err := hexutil.Decode(payload.DocumentIdentifier)
	if err != nil {
		return nil, err
	}

	tm, err := assembleTokenMessage(tokenIdentifier, granterID, granteeID, roleID, docID, docVersion)
	if err != nil {
		return nil, err
	}

	// fetch key pair from identity
	sig, err := account.SignMsg(tm)
	if err != nil {
		return nil, err
	}

	// assemble the access token, appending the signature and public keys
	at := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            granterID[:],
		Grantee:            granteeID[:],
		RoleIdentifier:     roleID,
		DocumentIdentifier: docID,
		Signature:          sig.Signature,
		Key:                account.GetSigningPublicKey(),
		DocumentVersion:    docVersion,
	}

	return at, nil
}

// assembleTokenMessage assembles a token message
func assembleTokenMessage(tokenIdentifier []byte, granterID identity.DID, granteeID identity.DID, roleID []byte, docID []byte, docVersion []byte) ([]byte, error) {
	ids := [][]byte{tokenIdentifier, roleID, docID}
	for _, id := range ids {
		if len(id) != idSize {
			return nil, ErrInvalidIDLength
		}
	}

	tm := append(tokenIdentifier, granterID[:]...)
	tm = append(tm, granteeID[:]...)
	tm = append(tm, roleID...)
	tm = append(tm, docID...)
	tm = append(tm, docVersion...)
	return tm, nil
}
