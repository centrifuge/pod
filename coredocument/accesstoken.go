package coredocument

import (
	"context"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)


	// UpdateCD request handling:

	// check if this is the document that should be updated to contain the AT (from request)
	// check if requesterId has permissions to update CD? or should this just be done in validation?

	// create AT with params
	// sign created AT

	// append to CD AT array
	// update new version of CD, anchor CD


	// AT validation/GetDocReq

	// check if AT grantee matches AT signature
	// check if AT grantee has sharing permissions for requested CD? (delegating_doc and requested doc)
		// how? in Roles or in ReadRules? only collaborators have sharing permissions?
		// is this synonymous with ReadRule.Action.Read?
	// check if requesterId is in Role for delegating document:
		// iterate through roles on CD indicated in DocReq.AccessTokenRequest.delegating_document_identifier
		// in every role, check if AT indicated by the DocReq.AccessTokenRequest.access_token_id is in role, if yes, return role_identifier of AT (and AT.document_identifier for later use?)
		// iterate through rules on CD, check the corresponding read_rule using previous role_identifier, if the associated action is Read/1, if yes, go to next check
	// if previous check has passed, go through the same process for the document identifier indicated in the access token

func IssueAccessToken(ctx context.Context, payload documentpb.AccessTokenParams, cd *coredocumentpb.CoreDocument,) (*coredocumentpb.AccessToken, error) {
	tokenIdentifier := utils.RandomSlice(32)
	id, err := contextutil.Self(ctx)
	if err != nil {
		return nil, err
	}
	granterId := id.ID.Bytes()
	// think about if roleId should be derived here or one step up
	roleId := byte(len(cd.Roles))
	if err != nil {
		return nil, err
	}
	grantee, err := hexutil.Decode(payload.GetGrantee())
	if err != nil {
		return nil, err
	}
	docId, err := hexutil.Decode(payload.GetDocumentIdentifier())
	if err != nil {
		return nil, err
	}
	// assemble access token message to be signed
	tm := append(tokenIdentifier, granterId...,)
	tm = append(grantee, roleId)
	tm = append(docId)
	// fetch key pair from identity
	privateKey := id.Keys[identity.KeyPurposeSigning].PrivateKey
	pubKey := id.Keys[identity.KeyPurposeSigning].PublicKey
	// sign the assembled access token message, return signature
	sig, err := crypto.SignMessage(privateKey, tm, "CurveSecp256K1", true)
	if err != nil {
		return nil, err
	}
	// assemble the access token, appending the signature and public keys
	at := new(coredocumentpb.AccessToken)
	at.Identifier = tokenIdentifier
    at.Granter = granterId
    at.Grantee = grantee
    at.RoleIdentifier = []byte{roleId}
    at.DocumentIdentifier = docId
    at.Signature = sig
    at.Key = pubKey

	return at, nil
}
