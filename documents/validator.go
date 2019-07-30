package documents

import (
	"bytes"
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// MaxAuthoredToCommitDuration is the maximum allowed time period for a document to be anchored after a authoring it based on document timestamp.
// I.E. This is basically the maximum time period allowed for document consensus to complete as well.
const MaxAuthoredToCommitDuration = 120 * time.Minute

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

// versionIDsValidator checks if the versions are properly set for new document update
func versionIDsValidator() Validator {
	return ValidatorFunc(func(old, new Model) error {
		if old == nil || new == nil {
			return errors.New("need both the old and new model")
		}

		var err error
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

// CreateVersionValidator validates if the new core document is properly derived from old one
func CreateVersionValidator(repo anchors.AnchorRepository) Validator {
	return ValidatorGroup{
		baseValidator(),
		currentVersionValidator(repo),
		LatestVersionValidator(repo),
	}
}

// UpdateVersionValidator validates if the new core document is properly derived from old one
func UpdateVersionValidator(repo anchors.AnchorRepository) Validator {
	return ValidatorGroup{
		versionIDsValidator(),
		currentVersionValidator(repo),
		LatestVersionValidator(repo),
	}
}

// baseValidator validates the core document basic fields like identifier, versions, and salts
func baseValidator() Validator {
	return ValidatorFunc(func(_, model Model) (err error) {
		if model == nil {
			return ErrModelNil
		}

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
		if model == nil {
			return ErrModelNil
		}

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
		if model == nil {
			return ErrModelNil
		}

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

// documentAuthorValidator checks if a given sender DID is the document author
func documentAuthorValidator(sender identity.DID) Validator {
	return ValidatorFunc(func(_, model Model) error {
		if model == nil {
			return ErrModelNil
		}

		author, err := model.Author()
		if err != nil {
			return err
		}
		if !author.Equal(sender) {
			return errors.New("document sender is not the author")
		}

		return nil
	})
}

// documentTimestampForSigningValidator checks if a given document has a timestamp recent enough to be signed
func documentTimestampForSigningValidator() Validator {
	return ValidatorFunc(func(_, model Model) error {
		if model == nil {
			return ErrModelNil
		}

		tm, err := model.Timestamp()
		if err != nil {
			return errors.New("failed to get document timestamp: %v", err)
		}

		if tm.Before(time.Now().UTC().Add(-MaxAuthoredToCommitDuration)) {
			return errors.New("document is too old to be signed")
		}
		return nil
	})
}

// signaturesValidator validates all the signatures in the core document
// assumes signing root is verified
// Note: can be used when during the signature request on collaborator side and post signature collection on sender side
// Note: this will break the current flow where we proceed to anchor even signatures verification fails
func signaturesValidator(idService identity.Service) Validator {
	return ValidatorFunc(func(_, model Model) error {
		if model == nil {
			return ErrModelNil
		}

		sr, err := model.CalculateSigningRoot()
		if err != nil {
			return errors.New("failed to get signing root: %v", err)
		}

		signatures := model.Signatures()
		if len(signatures) < 1 {
			return errors.New("atleast one signature expected")
		}
		author, err := model.Author()
		if err != nil {
			return err
		}
		collaborators, err := model.GetSignerCollaborators(author)
		if err != nil {
			return errors.New("could not get signer collaborators")
		}

		authorFound := false
		for _, sig := range signatures {
			sigDID, _ := identity.NewDIDFromBytes(sig.SignerId)
			if author.Equal(sigDID) {
				authorFound = true
			}

			// we only care about validating that signer is part of signing collaborators and not the other way around
			// since a collaborator can decide to not sign a document and the protocol still defines it as a valid state for a model.
			collaboratorFound := false
			for _, cb := range collaborators {
				if sigDID.Equal(cb) {
					collaboratorFound = true
				}
			}

			// signer is not found in signing collaborators and he is not the author either
			if !collaboratorFound && !author.Equal(sigDID) {
				err = errors.AppendError(
					err,
					errors.New("signature_%s verification failed: signer is not part of the signing collaborators", hexutil.Encode(sig.SignerId)))
				continue
			}

			tm, terr := model.Timestamp()
			if terr != nil {
				err = errors.AppendError(
					err,
					errors.New("signature_%s verification failed: %v", hexutil.Encode(sig.SignerId), terr))
				continue
			}

			if erri := idService.ValidateSignature(sigDID, sig.PublicKey, sig.Signature, sr, tm); erri != nil {
				err = errors.AppendError(
					err,
					errors.New("signature_%s verification failed: %v", hexutil.Encode(sig.SignerId), erri))
			}
		}
		if !authorFound {
			err = errors.AppendError(
				err,
				errors.New("signature verification failed: author's signature missing on document"))
		}
		return err
	})
}

// anchoredValidator checks if the document root matches the one on chain with specific anchorID
// assumes document root is generated and verified
func anchoredValidator(repo anchors.AnchorRepository) Validator {
	return ValidatorFunc(func(_, model Model) error {
		if model == nil {
			return ErrModelNil
		}

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

		gotRoot, anchoredAt, err := repo.GetAnchorData(anchorID)
		if err != nil {
			return errors.New("failed to get document root for anchor %s from chain: %v", anchorID.String(), err)
		}

		if !utils.IsSameByteSlice(docRoot[:], gotRoot[:]) {
			return errors.New("mismatched document roots")
		}

		tm, err := model.Timestamp()
		if err != nil {
			return errors.New("failed to get model update time: %v", err)
		}

		if tm.Add(MaxAuthoredToCommitDuration).Before(anchoredAt) {
			return errors.New("document was anchored after max allowed time for anchor %s", anchorID.String())
		}

		return nil
	})
}

// versionNotAnchoredValidator checks if the given version is not anchored on the chain.
// returns error if the version id is already anchored.
func versionNotAnchoredValidator(repo anchors.AnchorRepository, id []byte) error {
	anchorID, err := anchors.ToAnchorID(id)
	if err != nil {
		return errors.NewTypedError(ErrDocumentIdentifier, err)
	}

	_, _, err = repo.GetAnchorData(anchorID)
	if err == nil {
		return ErrDocumentIDReused
	}

	return nil
}

// LatestVersionValidator checks if the document is the latest version
func LatestVersionValidator(repo anchors.AnchorRepository) Validator {
	return ValidatorFunc(func(_, model Model) error {
		if model == nil {
			return ErrModelNil
		}

		err := versionNotAnchoredValidator(repo, model.NextVersion())
		if err != nil {
			return errors.NewTypedError(ErrDocumentNotLatest, err)
		}

		return nil
	})
}

// currentVersionValidator checks if the current version of the document has been anchored.
// returns an error if the current version has been anchored already.
func currentVersionValidator(repo anchors.AnchorRepository) Validator {
	return ValidatorFunc(func(_, model Model) error {
		if model == nil {
			return ErrModelNil
		}

		err := versionNotAnchoredValidator(repo, model.CurrentVersion())
		if err != nil {
			return errors.NewTypedError(ErrDocumentNotLatest, err)
		}

		return nil
	})
}

// anchorRepoAddressValidator validates if the model is using the configured anchor repository address.
func anchorRepoAddressValidator(anchoredRepoAddr common.Address) Validator {
	return ValidatorFunc(func(_, model Model) error {
		addr := model.AnchorRepoAddress()
		if !bytes.Equal(addr.Bytes(), anchoredRepoAddr.Bytes()) {
			return ErrDifferentAnchoredAddress
		}

		return nil
	})
}

// attributeValidator validates the signed attributes.
func attributeValidator(repo anchors.AnchorRepository, idSrv identity.Service) Validator {
	return ValidatorFunc(func(_, model Model) (err error) {
		ts, err := model.Timestamp()
		if err != nil {
			return err
		}

		attrs := model.GetAttributes()
		for _, attr := range attrs {
			if attr.Value.Type != AttrSigned {
				continue
			}

			signed := attr.Value.Signed
			var aid anchors.AnchorID
			aid, err = anchors.ToAnchorID(signed.DocumentVersion)
			if err != nil {
				return err
			}

			_, ats, erro := repo.GetAnchorData(aid)
			if erro != nil {
				// the attribute was added in this update itself.
				// pick the update time from the model itself
				ats = ts
			}

			payload := attributeSignaturePayload(signed.Identity[:], model.ID(), signed.DocumentVersion, signed.Value)
			erri := idSrv.ValidateSignature(signed.Identity, signed.PublicKey, signed.Signature, payload, ats)
			if erri != nil {
				err = errors.AppendError(err, errors.New("failed to validate signed attribute %s: %v", attr.KeyLabel, erri))
			}
		}

		return err
	})
}

// transitionValidator checks that the document changes are within the transition_rule capability of the
// collaborator making the changes
func transitionValidator(collaborator identity.DID) Validator {
	return ValidatorFunc(func(old, new Model) error {
		if old == nil {
			return nil
		}

		err := old.CollaboratorCanUpdate(new, collaborator)
		if err != nil {
			return errors.New("invalid document state transition: %v", err)
		}

		return nil
	})
}

// PreAnchorValidator is a validator group with following validators
// base validator
// signing root validator
// document root validator
// signatures validator
// should be called before pre anchoring
func PreAnchorValidator(idService identity.Service, repo anchors.AnchorRepository) ValidatorGroup {
	return ValidatorGroup{
		SignatureValidator(idService, repo),
		documentRootValidator(),
	}
}

// PostAnchoredValidator is a validator group with following validators
// PreAnchorValidator
// anchoredValidator
// should be called after anchoring the document/when received anchored document
func PostAnchoredValidator(idService identity.Service, repo anchors.AnchorRepository) ValidatorGroup {
	return ValidatorGroup{
		PreAnchorValidator(idService, repo),
		anchoredValidator(repo),
		LatestVersionValidator(repo),
	}
}

// ReceivedAnchoredDocumentValidator is a validator group with following validators
// transitionValidator
// PostAnchoredValidator
func ReceivedAnchoredDocumentValidator(
	idService identity.Service,
	repo anchors.AnchorRepository,
	collaborator identity.DID) ValidatorGroup {
	return ValidatorGroup{
		transitionValidator(collaborator),
		PostAnchoredValidator(idService, repo),
	}
}

// RequestDocumentSignatureValidator is a validator group with the following validators
// SignatureValidator
// transitionsValidator
// it should be called when a document is received over the p2p layer before signing
func RequestDocumentSignatureValidator(
	repo anchors.AnchorRepository,
	idService identity.Service,
	collaborator identity.DID,
	anchorRepoAddress common.Address) ValidatorGroup {
	return ValidatorGroup{
		documentTimestampForSigningValidator(),
		documentAuthorValidator(collaborator),
		currentVersionValidator(repo),
		LatestVersionValidator(repo),
		anchorRepoAddressValidator(anchorRepoAddress),
		transitionValidator(collaborator),
		SignatureValidator(idService, repo),
	}
}

// SignatureValidator is a validator group with following validators
// baseValidator
// signingRootValidator
// signaturesValidator
// should be called after sender signing the document, before requesting the document and after signature collection
func SignatureValidator(idService identity.Service, repo anchors.AnchorRepository) ValidatorGroup {
	return ValidatorGroup{
		baseValidator(),
		signingRootValidator(),
		signaturesValidator(idService),
		attributeValidator(repo, idService),
	}
}
