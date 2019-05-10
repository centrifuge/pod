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

	idxFunding, err := s.findFunding(model, fundingID)
	if err != nil {
		return nil, ErrFundingNotFound
	}

	attributes, err := s.createSignAttrs(model, idxFunding, selfDID, account)
	if err != nil {
		return nil, err
	}

	err = model.AddAttributes(attributes...)
	if err != nil {
		return nil, err
	}

	return model, nil
}


//func verifySignature()

func (s service) signAttrToClientData(ctx context.Context, current documents.Model, signAttr documents.Attribute) (*clientfundingpb.FundingSignature, error) {
	if signAttr.Value.Type != documents.AttrSigned {
		return nil, ErrFundingSignature
	}


	//docSigned, err := s.Service.GetVersion(ctx,current.ID(),signAttr.Value.Signed.DocumentVersion)

	return &clientfundingpb.FundingSignature{Valid:true,SignedVersion:hexutil.Encode(current.ID())},nil
}

func (s service) deriveFundingSignatures(ctx context.Context, model documents.Model, idxFunding string) ([]*clientfundingpb.FundingSignature, error) {

	var signatures []*clientfundingpb.FundingSignature
	// example "funding_agreement[2].signatures"
	sLabel := generateLabel(fundingFieldKey, idxFunding, fundingSignatures)

	lastIdx, err := getArrayLatestIDX(model, sLabel)
	if err != nil {
		return nil, err
	}

	i, err := documents.NewInt256("0")
	if err != nil {
		return nil, err
	}

	for i.Cmp(lastIdx) != 1 {
		// example: "funding_agreement[2].signatures[4]"
		sFieldLabel := generateLabel(generateLabel(fundingFieldKey, idxFunding, "")+fundingSignaturesFieldKey, i.String(), "")
		key, err := documents.AttrKeyFromLabel(sFieldLabel)
		if err != nil {
			return nil, err
		}
		attrSign, err := model.GetAttribute(key)
		if err != nil {
			return nil, err
		}

		clientSign, err := s.signAttrToClientData(ctx, model, attrSign)
		if err != nil {
			continue
		}
		signatures= append(signatures, clientSign)
		i, err = i.Inc()

		if err != nil {
			return nil, err
		}

	}
	return signatures, nil
}




