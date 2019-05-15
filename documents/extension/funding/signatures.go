package funding

import (
	"context"
	"encoding/json"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func (s service) createSignAttrs(model documents.Model, idxFunding string, selfDID identity.DID, account config.Account) ([]documents.Attribute, error) {
	var attributes []documents.Attribute
	data, err := s.deriveFundingData(model, idxFunding)
	if err != nil {
		return nil, err
	}

	signMsg, err := json.Marshal(data)
	if err != nil {
		return nil, ErrJSON
	}

	// example "funding_agreement[2].signatures"
	sLabel := generateLabel(fundingFieldKey, idxFunding, fundingSignatures)
	attrIdx, err := incrementArrayAttrIDX(model, sLabel)
	if err != nil {
		return nil, err
	}
	attributes = append(attributes, attrIdx)

	// example: "funding_agreement[2].signatures[4]"
	sFieldLabel := generateLabel(generateLabel(fundingFieldKey, idxFunding, "")+fundingSignaturesFieldKey, attrIdx.Value.Int256.String(), "")

	attrSign, err := documents.NewSignedAttribute(sFieldLabel, selfDID, account, model, signMsg)
	if err != nil {
		return nil, err
	}
	attributes = append(attributes, attrSign)

	return attributes, nil
}

// Sign adds a signature to an existing document
func (s service) Sign(ctx context.Context, fundingID string, identifier []byte) (documents.Model, error) {
	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	account, err := contextutil.Account(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigAccountID, err)
	}

	model, err := s.Service.GetCurrentVersion(ctx, identifier)
	if err != nil {
		return nil, documents.ErrDocumentNotFound
	}

	idxFunding, err := s.findFundingIDX(model, fundingID)
	if err != nil {
		return nil, ErrFundingNotFound
	}

	attributes, err := s.createSignAttrs(model, idxFunding, selfDID, account)
	if err != nil {
		return nil, err
	}

	err = model.AddAttributes(nil, attributes...)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func (s service) validateValueOfSignAttr(funding *Data, signAttr documents.Attribute) (bool, error) {
	value, err := json.Marshal(funding)
	if err != nil {
		return false, ErrJSON
	}
	return utils.IsSameByteSlice(value, signAttr.Value.Signed.Value), nil
}

func (s service) validateSignedFundingVersion(ctx context.Context, identifier []byte, fundingID string, signAttr documents.Attribute) (*clientfundingpb.FundingSignature, error) {
	did := signAttr.Value.Signed.Identity
	signedDocVersion, err := s.Service.GetVersion(ctx, identifier, signAttr.Value.Signed.DocumentVersion)
	if err != nil {
		return nil, documents.ErrDocumentNotFound
	}

	signedFunding, err := s.findFunding(signedDocVersion, fundingID)
	if err != nil {
		return nil, ErrFundingNotFound
	}

	valid, err := s.validateValueOfSignAttr(signedFunding, signAttr)
	if err != nil {
		return nil, err
	}

	if valid {
		// the value of the older funding version signature is correct
		return &clientfundingpb.FundingSignature{Valid: true, SignedVersion: hexutil.Encode(identifier), Identity: did.String(), OutdatedSignature: true}, nil
	}

	return &clientfundingpb.FundingSignature{Valid: false, SignedVersion: hexutil.Encode(identifier), Identity: did.String(), OutdatedSignature: true}, nil
}

func (s service) signAttrToClientData(ctx context.Context, current documents.Model, funding *Data, signAttr documents.Attribute) (*clientfundingpb.FundingSignature, error) {
	if signAttr.Value.Type != documents.AttrSigned {
		return nil, ErrFundingSignature
	}

	did := signAttr.Value.Signed.Identity
	valid, err := s.validateValueOfSignAttr(funding, signAttr)
	if err != nil {
		return nil, err
	}

	// value correct (funding data didn't change since signing)
	if valid {
		return &clientfundingpb.FundingSignature{Valid: true, SignedVersion: hexutil.Encode(current.ID()), Identity: did.String(), OutdatedSignature: false}, nil
	}

	return s.validateSignedFundingVersion(ctx, current.ID(), funding.FundingId, signAttr)
}

func (s service) deriveFundingSignatures(ctx context.Context, model documents.Model, funding *Data, idxFunding string) ([]*clientfundingpb.FundingSignature, error) {
	var signatures []*clientfundingpb.FundingSignature
	sLabel := generateLabel(fundingFieldKey, idxFunding, fundingSignatures)
	key, err := documents.AttrKeyFromLabel(sLabel)
	if err != nil {
		return nil, err
	}

	if !model.AttributeExists(key) {
		return signatures, nil
	}

	lastIdx, err := getArrayLatestIDX(model, sLabel)
	if err != nil {
		return nil, err
	}

	i, err := documents.NewInt256("0")
	if err != nil {
		return nil, err
	}

	for i.Cmp(lastIdx) != 1 {
		sFieldLabel := generateLabel(generateLabel(fundingFieldKey, idxFunding, "")+fundingSignaturesFieldKey, i.String(), "")
		key, err := documents.AttrKeyFromLabel(sFieldLabel)
		if err != nil {
			return nil, err
		}
		attrSign, err := model.GetAttribute(key)
		if err != nil {
			return nil, err
		}

		clientSign, err := s.signAttrToClientData(ctx, model, funding, attrSign)
		if err != nil {
			continue
		}
		signatures = append(signatures, clientSign)
		i, err = i.Inc()

		if err != nil {
			return nil, err
		}

	}
	return signatures, nil
}
