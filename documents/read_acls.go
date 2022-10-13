package documents

import (
	"bytes"
	"context"
	"fmt"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// initReadRules initiates the read rules for a given CoreDocumentModel.
// Collaborators are given Read_Sign action.
// if the rules are created already, this is a no-op.
// if collaborators are empty, it is a no-op
func (cd *CoreDocument) initReadRules(collaborators []*types.AccountID) {
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
func (cd *CoreDocument) addCollaboratorsToReadSignRules(collaborators []*types.AccountID) {
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
func findReadRole(cd *coredocumentpb.CoreDocument, onRole func(ruleIndex, roleIndex int, role *coredocumentpb.Role) bool, actions ...coredocumentpb.Action) bool {
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
func findTransitionRole(cd *coredocumentpb.CoreDocument, onRole func(rridx, ridx int, role *coredocumentpb.Role) bool, actions ...coredocumentpb.TransitionAction) bool {
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

func (cd *CoreDocument) NFTCanRead(encodedCollectionID []byte, encodedItemID []byte) bool {
	return findReadRole(
		cd.Document,
		func(_, _ int, role *coredocumentpb.Role) bool {
			_, found := isNFTInRole(role, encodedCollectionID, encodedItemID)
			return found
		},
		coredocumentpb.Action_ACTION_READ,
	)
}

// AccountCanRead validate if the core document can be read by the account .
func (cd *CoreDocument) AccountCanRead(accountID *types.AccountID) bool {
	return findReadRole(
		cd.Document,
		func(_, _ int, role *coredocumentpb.Role) bool {
			_, found := isAccountIDinRole(role, accountID)
			return found
		},
		coredocumentpb.Action_ACTION_READ, coredocumentpb.Action_ACTION_READ_SIGN,
	)
}

// addNFTToReadRules adds NFT token to the read rules of core document.
func (cd *CoreDocument) addNFTToReadRules(encodedCollectionID, encodedItemID []byte) error {
	nft, err := ConstructNFT(encodedCollectionID, encodedItemID)
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
func (cd *CoreDocument) AddNFT(grantReadAccess bool, collectionID types.U64, itemID types.U128) (*CoreDocument, error) {
	ncd, err := cd.PrepareNewVersion(nil, CollaboratorsAccess{}, nil)
	if err != nil {
		return nil, errors.New("failed to prepare new version: %v", err)
	}

	encodedCollectionID, err := codec.Encode(collectionID)

	if err != nil {
		return nil, fmt.Errorf("couldn't encode collection ID to bytes: %w", err)
	}

	encodedItemID, err := codec.Encode(itemID)

	if err != nil {
		return nil, fmt.Errorf("couldn't encode item ID to bytes: %w", err)
	}

	var nft *coredocumentpb.NFT

	for _, docNFT := range ncd.Document.GetNfts() {
		if bytes.Equal(docNFT.GetCollectionId(), encodedCollectionID) {
			if bytes.Equal(docNFT.GetItemId(), encodedItemID) {
				return nil, errors.New("nft already exists")
			}

			// TODO(cdamian): Confirm replacement of item ID.
			// Found an NFT with the current collection ID, in this case, we will overwrite the item ID, if any,
			// with the new one.
			nft = docNFT
			break
		}
	}

	if nft == nil {
		nft = &coredocumentpb.NFT{
			CollectionId: encodedCollectionID,
		}

		ncd.Document.Nfts = append(ncd.Document.Nfts, nft)
	}

	nft.ItemId = encodedItemID

	if grantReadAccess {
		if err := ncd.addNFTToReadRules(nft.GetCollectionId(), nft.GetItemId()); err != nil {
			return nil, fmt.Errorf("couldn't add NFT to read rules: %w", err)
		}
	}

	cd.Modified = true
	return ncd, nil
}

// NFTs returns the list of NFTs created for this model
func (cd *CoreDocument) NFTs() []*coredocumentpb.NFT {
	return cd.Document.Nfts
}

// ConstructNFT checks the sizes of the encoded collection and item IDs and concatenates them.
func ConstructNFT(encodedCollectionID []byte, encodedItemID []byte) ([]byte, error) {
	switch {
	case len(encodedCollectionID) != nftCollectionIDByteCount:
		return nil, errors.NewTypedError(ErrNftByteLength, errors.New("provided length %d", len(encodedCollectionID)))
	case len(encodedItemID) != nftItemIDByteCount:
		return nil, errors.NewTypedError(ErrNftByteLength, errors.New("provided length %d", len(encodedItemID)))
	default:
		nft := append(encodedCollectionID, encodedItemID...)

		return nft, nil
	}
}

// isNFTInRole checks if the given nft is part of the core document role.
// If found, returns the index of the nft in the role and true
func isNFTInRole(role *coredocumentpb.Role, encodedCollectionID []byte, encodedItemID []byte) (nftIdx int, found bool) {
	enft, err := ConstructNFT(encodedCollectionID, encodedItemID)
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

func getStoredNFT(nfts []*coredocumentpb.NFT, encodedCollectionID []byte) *coredocumentpb.NFT {
	for _, nft := range nfts {
		if bytes.Equal(nft.GetCollectionId(), encodedCollectionID) {
			return nft
		}
	}

	return nil
}

func getReadAccessProofKeys(
	cd *coredocumentpb.CoreDocument,
	encodedCollectionID []byte,
	encodedItemID []byte,
) (pks []string, err error) {
	var rridx int  // index of the read rules which contain the role
	var ridx int   // index of the role
	var nftIdx int // index of the NFT in the above role
	var rk []byte  // role key of the above role

	found := findReadRole(
		cd,
		func(i, j int, role *coredocumentpb.Role) bool {
			z, found := isNFTInRole(role, encodedCollectionID, encodedItemID)
			if found {
				rridx = i
				ridx = j
				rk = role.RoleKey
				nftIdx = z
			}

			return found
		},
		coredocumentpb.Action_ACTION_READ,
	)

	if !found {
		return nil, ErrNFTRoleMissing
	}

	return []string{
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].roles[%d]", rridx, ridx),          // proof that a read rule exists with the nft role
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].action", rridx),                   // proof that this read rule has read access
		fmt.Sprintf(CDTreePrefix+".roles[%s].nfts[%d]", hexutil.Encode(rk), nftIdx), // proof that role with nft exists
	}, nil
}

// isAccountIDinRole returns the index of the collaborator and true if did is in the given role as collaborators.
func isAccountIDinRole(role *coredocumentpb.Role, accountID *types.AccountID) (idx int, found bool) {
	for i, id := range role.Collaborators {
		if bytes.Equal(id, accountID.ToBytes()) {
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

// validateAccessToken validates that given access token against its signature
func validateAccessToken(publicKey []byte, token *coredocumentpb.AccessToken, requesterID []byte) error {
	// assemble token message from the token for validation
	reqID, err := types.NewAccountID(requesterID)
	if err != nil {
		return ErrRequesterInvalidAccountID
	}
	granterID, err := types.NewAccountID(token.Granter)
	if err != nil {
		return ErrGranterInvalidAccountID
	}
	tm, err := AssembleTokenMessage(token.Identifier, granterID, reqID, token.RoleIdentifier, token.DocumentIdentifier, token.DocumentVersion)
	if err != nil {
		return err
	}
	validated := crypto.VerifyMessage(publicKey, tm, token.Signature, crypto.CurveEd25519)
	if !validated {
		return ErrAccessTokenInvalid
	}
	return nil
}

func (cd *CoreDocument) findAccessToken(tokenID []byte) (at *coredocumentpb.AccessToken, err error) {
	// check if the access token is present on the document indicated in the AT request
	for _, at := range cd.Document.AccessTokens {
		if bytes.Equal(tokenID, at.Identifier) {
			return at, nil
		}
	}
	return at, ErrAccessTokenNotFound
}

// ATGranteeCanRead checks that the grantee of the access token can read the document requested
func (cd *CoreDocument) ATGranteeCanRead(ctx context.Context, docService Service, identityService v2.Service, tokenID, docID []byte, requesterID *types.AccountID) (err error) {
	// find the access token
	at, err := cd.findAccessToken(tokenID)
	if err != nil {
		return err
	}
	granterID, err := types.NewAccountID(at.Granter)
	if err != nil {
		return ErrGranterInvalidAccountID
	}
	granteeID, err := types.NewAccountID(at.Grantee)
	if err != nil {
		return ErrGranteeInvalidAccountID
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
	if !bytes.Equal(at.GetDocumentIdentifier(), docID) {
		return ErrReqDocNotMatch
	}
	_, err = docService.GetVersion(ctx, cd.Document.DocumentIdentifier, at.DocumentVersion)
	if err != nil {
		return ErrDocumentRetrieval
	}

	timestamp, err := cd.Timestamp()

	if err != nil {
		return ErrDocumentTimestampRetrieval
	}

	// validate that the public key of the granter is the public key that has been used to sign the access token
	err = identityService.ValidateKey(granterID, at.Key, keystoreType.KeyPurposeP2PDocumentSigning, timestamp)
	if err != nil {
		return ErrDocumentSigningKeyValidation
	}

	return validateAccessToken(at.Key, at, granteeID.ToBytes())
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
func (cd *CoreDocument) DeleteAccessToken(granteeID *types.AccountID) (*CoreDocument, error) {
	ncd, err := cd.PrepareNewVersion(nil, CollaboratorsAccess{}, nil)
	if err != nil {
		return nil, err
	}

	accessTokens := ncd.Document.AccessTokens
	for i, t := range accessTokens {
		if bytes.Equal(t.Grantee, granteeID.ToBytes()) {
			ncd.Document.AccessTokens = removeTokenAtIndex(i, accessTokens)
			ncd.Modified = true
			return ncd, nil
		}
	}
	return nil, ErrAccessTokenNotFound
}

// RemoveTokenAtIndex removes the access token at index i from slice a and returns a new slice
// Note: changes the order of the slice elements
func removeTokenAtIndex(idx int, tokens []*coredocumentpb.AccessToken) []*coredocumentpb.AccessToken {
	result := make([]*coredocumentpb.AccessToken, len(tokens))

	copy(result, tokens)

	result[idx] = result[len(result)-1]
	result[len(result)-1] = nil
	result = result[:len(result)-1]

	return result
}

// assembleAccessToken assembles a Read Access Token from the payload received
func assembleAccessToken(ctx context.Context, params AccessTokenParams, docVersion []byte) (*coredocumentpb.AccessToken, error) {
	account, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}
	granterID := account.GetIdentity()

	// TODO: this roleID will be specified later with field level read access
	roleID := utils.RandomSlice(32)
	tokenIdentifier := utils.RandomSlice(32)

	granteeID, err := types.NewAccountIDFromHexString(params.Grantee)
	if err != nil {
		return nil, err
	}
	// assemble access token message to be signed
	docID, err := hexutil.Decode(params.DocumentIdentifier)
	if err != nil {
		return nil, err
	}

	tm, err := AssembleTokenMessage(tokenIdentifier, granterID, granteeID, roleID, docID, docVersion)
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
		Granter:            granterID.ToBytes(),
		Grantee:            granteeID.ToBytes(),
		RoleIdentifier:     roleID,
		DocumentIdentifier: docID,
		Signature:          sig.GetSignature(),
		Key:                sig.GetPublicKey(),
		DocumentVersion:    docVersion,
	}

	return at, nil
}

// AssembleTokenMessage assembles a token message
func AssembleTokenMessage(
	tokenIdentifier []byte,
	granterID *types.AccountID,
	granteeID *types.AccountID,
	roleID []byte,
	docID []byte,
	docVersion []byte,
) ([]byte, error) {
	ids := [][]byte{tokenIdentifier, roleID, docID}
	for _, id := range ids {
		if len(id) != idSize {
			return nil, ErrInvalidIDLength
		}
	}

	tm := append(tokenIdentifier, granterID.ToBytes()...)
	tm = append(tm, granteeID.ToBytes()...)
	tm = append(tm, roleID...)
	tm = append(tm, docID...)
	tm = append(tm, docVersion...)
	return tm, nil
}
