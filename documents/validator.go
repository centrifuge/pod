package documents

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Validator is an interface every Validator (atomic or group) should implement
type Validator interface {
	// Validate validates the updates to the model in newState.
	Validate(oldState Model, newState Model) error
}

// ValidatorGroup implements Validator for validating a set of validators.
type ValidatorGroup []Validator

//Validate will execute all group specific atomic validations
func (group ValidatorGroup) Validate(oldState Model, newState Model) (errs error) {

	for _, v := range group {
		if err := v.Validate(oldState, newState); err != nil {
			errs = errors.AppendError(errs, err)
		}
	}
	return errs
}

// IsCurrencyValid checks if the currency is of length 3
func IsCurrencyValid(cur string) bool {
	return utils.IsStringOfLength(cur, 3)
}

// ValidatorFunc implements Validator and can be used as a adaptor for functions
// with specific function signature
type ValidatorFunc func(old, new Model) error

// Validate passes the arguments to the underlying validator
// function and returns the results
func (vf ValidatorFunc) Validate(old, new Model) error {
	return vf(old, new)
}

// UpdateVersionValidator validates if the new core document is properly derived from old one
func UpdateVersionValidator() Validator {
	return ValidatorFunc(func(old, new Model) error {
		if old == nil || new == nil {
			return errors.New("need both the old and new model")
		}

		dr, err := old.CalculateDocumentRoot()
		if err != nil {
			return errors.New("failed to get previous version document root: %v", err)
		}
		checks := []struct {
			name string
			a, b []byte
		}{
			{
				name: "cd_document_identifier",
				a:    old.ID(),
				b:    new.ID(),
			},

			{
				name: "cd_previous_version",
				a:    old.CurrentVersion(),
				b:    new.PreviousVersion(),
			},

			{
				name: "cd_current_version",
				a:    old.NextVersion(),
				b:    new.CurrentVersion(),
			},

			{
				name: "cd_document_root",
				a:    dr,
				b:    new.PreviousDocumentRoot(),
			},
		}

		for _, c := range checks {
			if !utils.CheckMultiple32BytesFilled(c.a, c.b) {
				err = errors.AppendError(err, errors.New("missing identifiers"))
				continue
			}

			if !utils.IsSameByteSlice(c.a, c.b) {
				err = errors.AppendError(err, errors.New("%s: mismatch", c.name))
			}
		}

		if utils.IsEmptyByteSlice(new.NextVersion()) {
			err = errors.AppendError(err, errors.New("cd_next_version: not set"))
		}

		return err
	})
}

// baseValidator validates the core document basic fields like identifier, versions, and salts
func baseValidator() Validator {
	return ValidatorFunc(func(_, model Model) (err error) {
		if utils.IsEmptyByteSlice(model.ID()) {
			err = errors.AppendError(err, errors.New("document identifier not set"))
		}

		if utils.IsEmptyByteSlice(model.CurrentVersion()) {
			err = errors.AppendError(err, errors.New("current version not set"))
		}

		if utils.IsEmptyByteSlice(model.NextVersion()) {
			err = errors.AppendError(err, errors.New("next version not set"))
		}

		// double check the identifiers
		isSameBytes := utils.IsSameByteSlice
		// Problem (re-using an old identifier for NextVersion): CurrentVersion or DocumentIdentifier same as NextVersion
		if isSameBytes(model.NextVersion(), model.ID()) ||
			isSameBytes(model.NextVersion(), model.CurrentVersion()) {
			err = errors.AppendError(err, errors.New("identifiers re-used"))
		}

		return err
	})
}

// signingRootValidator checks the existence of signing root
func signingRootValidator() Validator {
	return ValidatorFunc(func(_, model Model) error {
		sr, err := model.CalculateSigningRoot()
		if err != nil {
			return errors.New("failed to get signing root: %v", err)
		}

		if len(sr) != idSize {
			return errors.New("signing root is invalid")
		}

		return nil
	})
}

// documentRootValidator checks the existence of document root
// recalculates the document root and compares with existing one
func documentRootValidator() Validator {
	return ValidatorFunc(func(_, model Model) error {
		dr, err := model.CalculateDocumentRoot()
		if err != nil {
			return errors.New("failed to get document root: %v", err)
		}

		if len(dr) != idSize {
			return errors.New("document root is invalid")
		}

		return nil
	})
}

// readyForSignaturesValidator validates self signature
// re-calculates the signature and compares with existing one
// assumes signing_root is already generated and verified
// Note: this needs to used only before document is sent for signatures from the collaborators
func readyForSignaturesValidator(id, priv, pub []byte) Validator {
	return ValidatorFunc(func(_, model Model) error {
		sr, err := model.CalculateSigningRoot()
		if err != nil {
			return errors.New("failed to generate signing root: %v", err)
		}

		signatures := model.Signatures()
		if len(signatures) != 1 {
			return errors.New("expecting only one signature")
		}

		s := crypto.Sign(id, priv, pub, sr)
		sh := signatures[0]
		if !utils.IsSameByteSlice(s.EntityId, sh.EntityId) {
			err = errors.AppendError(err, errors.New("entity ID mismatch"))
		}

		if !utils.IsSameByteSlice(s.PublicKey, sh.PublicKey) {
			err = errors.AppendError(err, errors.New("public key mismatch"))
		}

		if !utils.IsSameByteSlice(s.Signature, sh.Signature) {
			err = errors.AppendError(err, errors.New("signature mismatch"))
		}

		return err
	})
}

// signaturesValidator validates all the signatures in the core document
// assumes signing root is verified
// Note: can be used when during the signature request on collaborator side and post signature collection on sender side
// Note: this will break the current flow where we proceed to anchor even signatures verification fails
func signaturesValidator(idService identity.ServiceDID) Validator {
	return ValidatorFunc(func(_, model Model) error {
		sr, err := model.CalculateSigningRoot()
		if err != nil {
			return errors.New("failed to get signing root: %v", err)
		}

		signatures := model.Signatures()
		if len(signatures) < 1 {
			return errors.New("atleast one signature expected")
		}

		for _, sig := range signatures {
			if erri := idService.ValidateSignature(&sig, sr); erri != nil {
				err = errors.AppendError(
					err,
					errors.New("signature_%s verification failed: %v", hexutil.Encode(sig.EntityId), erri))
			}
		}
		return err
	})
}

// anchoredValidator checks if the document root matches the one on chain with specific anchorID
// assumes document root is generated and verified
func anchoredValidator(repo anchors.AnchorRepository) Validator {
	return ValidatorFunc(func(_, model Model) error {
		anchorID, err := anchors.ToAnchorID(model.CurrentVersion())
		if err != nil {
			return errors.New("failed to get anchorID: %v", err)
		}

		dr, err := model.CalculateDocumentRoot()
		if err != nil {
			return errors.New("failed to get document root: %v", err)
		}

		docRoot, err := anchors.ToDocumentRoot(dr)
		if err != nil {
			return errors.New("failed to get document root: %v", err)
		}

		gotRoot, err := repo.GetDocumentRootOf(anchorID)
		if err != nil {
			return errors.New("failed to get document root from chain: %v", err)
		}

		if !utils.IsSameByteSlice(docRoot[:], gotRoot[:]) {
			return errors.New("mismatched document roots")
		}

		return nil
	})
}

// SignatureRequestValidator returns a validator group with following validators
// base validator
// signing root validator
// signatures validator
// should be used when node receives a document requesting for signature
func SignatureRequestValidator(idService identity.ServiceDID) ValidatorGroup {
	return PostSignatureRequestValidator(idService)
}

// PreAnchorValidator is a validator group with following validators
// base validator
// signing root validator
// document root validator
// signatures validator
// should be called before pre anchoring
func PreAnchorValidator(idService identity.ServiceDID) ValidatorGroup {
	return ValidatorGroup{
		PostSignatureRequestValidator(idService),
		documentRootValidator(),
	}
}

// PostAnchoredValidator is a validator group with following validators
// PreAnchorValidator
// anchoredValidator
// should be called after anchoring the document/when received anchored document
func PostAnchoredValidator(idService identity.ServiceDID, repo anchors.AnchorRepository) ValidatorGroup {
	return ValidatorGroup{
		PreAnchorValidator(idService),
		anchoredValidator(repo),
	}
}

// PreSignatureRequestValidator is a validator group with following validators
// baseValidator
// signingRootValidator
// readyForSignaturesValidator
// should be called after sender signing the document and before requesting the document
func PreSignatureRequestValidator(id, priv, pub []byte) ValidatorGroup {
	return ValidatorGroup{
		baseValidator(),
		signingRootValidator(),
		readyForSignaturesValidator(id, priv, pub),
	}
}

// PostSignatureRequestValidator is a validator group with following validators
// baseValidator
// signingRootValidator
// signaturesValidator
// should be called after the signature collection/before preparing for anchoring
func PostSignatureRequestValidator(idService identity.ServiceDID) ValidatorGroup {
	return ValidatorGroup{
		baseValidator(),
		signingRootValidator(),
		signaturesValidator(idService),
	}
}
