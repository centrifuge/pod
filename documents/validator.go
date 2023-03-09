package documents

import (
	"reflect"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/pallets/anchors"
	"github.com/centrifuge/pod/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// MaxAuthoredToCommitDuration is the maximum allowed time period for a document to be anchored after a authoring it based on document timestamp.
// I.E. This is basically the maximum time period allowed for document consensus to complete as well.
const MaxAuthoredToCommitDuration = 120 * time.Minute

//go:generate mockery --name Validator --structname ValidatorMock --filename validator_mock.go --inpackage

// Validator is an interface every Validator (atomic or group) should implement
type Validator interface {
	// Validate validates the updates to the model in newState.
	Validate(oldState Document, newState Document) error
}

// ValidatorGroup implements Validator for validating a set of validators.
type ValidatorGroup []Validator

// Validate will execute all group specific atomic validations
func (group ValidatorGroup) Validate(oldState Document, newState Document) (errs error) {
	for _, v := range group {
		if err := v.Validate(oldState, newState); err != nil {
			errs = errors.AppendError(errs, err)
		}
	}
	return errs
}

const (
	currencyLength = 3
)

// IsCurrencyValid checks if the currency is of length 3
func IsCurrencyValid(cur string) bool {
	return utils.IsStringOfLength(cur, currencyLength)
}

// ValidatorFunc implements Validator and can be used as an adaptor for functions
// with specific function signature
type ValidatorFunc func(old, new Document) error

// Validate passes the arguments to the underlying validator
// function and returns the results
func (vf ValidatorFunc) Validate(old, new Document) error {
	return vf(old, new)
}

// versionIDsValidator checks if the versions are properly set for new document update
func versionIDsValidator() Validator {
	return ValidatorFunc(func(old, new Document) error {
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

// baseValidator validates the core document basic fields like identifier, versions, and salts
func baseValidator() Validator {
	return ValidatorFunc(func(_, model Document) (err error) {
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
	return ValidatorFunc(func(_, model Document) error {
		if model == nil {
			return ErrModelNil
		}

		sr, err := model.CalculateSigningRoot()
		if err != nil {
			return errors.NewTypedError(ErrDocumentCalculateSigningRoot, err)
		}

		if len(sr) != idSize {
			return ErrInvalidSigningRoot
		}

		return nil
	})
}

// documentRootValidator checks the existence of document root
// recalculates the document root and compares with existing one
func documentRootValidator() Validator {
	return ValidatorFunc(func(_, model Document) error {
		if model == nil {
			return ErrModelNil
		}

		dr, err := model.CalculateDocumentRoot()
		if err != nil {
			return errors.NewTypedError(ErrDocumentCalculateDocumentRoot, err)
		}

		if len(dr) != idSize {
			return ErrInvalidDocumentRoot
		}

		return nil
	})
}

// documentAuthorValidator checks if a given sender DID is the document author
func documentAuthorValidator(sender *types.AccountID) Validator {
	return ValidatorFunc(func(_, model Document) error {
		if model == nil {
			return ErrModelNil
		}

		author, err := model.Author()
		if err != nil {
			return ErrDocumentAuthorRetrieval
		}

		if !author.Equal(sender) {
			return ErrDocumentSenderNotAuthor
		}

		return nil
	})
}

// documentTimestampForSigningValidator checks if a given document has a timestamp recent enough to be signed
func documentTimestampForSigningValidator() Validator {
	return ValidatorFunc(func(_, model Document) error {
		if model == nil {
			return ErrModelNil
		}

		tm, err := model.Timestamp()
		if err != nil {
			return errors.NewTypedError(ErrDocumentTimestampRetrieval, err)
		}

		if tm.Before(time.Now().UTC().Add(-MaxAuthoredToCommitDuration)) {
			return ErrDocumentTooOldToSign
		}
		return nil
	})
}

// signaturesValidator validates all the signatures in the core document
// assumes signing root is verified
// Note: can be used when during the signature request on collaborator side and post signature collection on sender side
// Note: this will break the current flow where we proceed to anchor even signatures verification fails
func signaturesValidator(identityService v2.Service) Validator {
	return ValidatorFunc(func(_, model Document) error {
		if model == nil {
			return ErrModelNil
		}

		sr, err := model.CalculateSigningRoot()
		if err != nil {
			return errors.NewTypedError(ErrDocumentCalculateSigningRoot, err)
		}

		signatures := model.Signatures()
		if len(signatures) < 1 {
			return ErrDocumentNoSignatures
		}

		author, err := model.Author()
		if err != nil {
			return ErrDocumentAuthorRetrieval
		}

		collaborators, err := model.GetSignerCollaborators(author)
		if err != nil {
			return ErrDocumentSignerCollaboratorsRetrieval
		}

		authorFound := false
		for _, sig := range signatures {
			signerAccountID, accErr := types.NewAccountID(sig.SignerId)

			if accErr != nil {
				err = errors.AppendError(
					err,
					errors.New("signature_%s verification failed: couldn't parse signer account ID", hexutil.Encode(sig.SignerId)),
				)
				continue
			}

			if author.Equal(signerAccountID) {
				authorFound = true
			}

			// we only care about validating that signer is part of signing collaborators and not the other way around
			// since a collaborator can decide to not sign a document and the protocol still defines it as a valid state for a model.
			collaboratorFound := false
			for _, cb := range collaborators {
				if signerAccountID.Equal(cb) {
					collaboratorFound = true
				}
			}

			// signer is not found in signing collaborators, and they are not the author either
			if !collaboratorFound && !author.Equal(signerAccountID) {
				err = errors.AppendError(
					err,
					errors.New("signature_%s verification failed: signer is not part of the signing collaborators", hexutil.Encode(sig.SignerId)))
				continue
			}

			timestamp, timestampErr := model.Timestamp()

			if timestampErr != nil {
				err = errors.AppendError(err, errors.New("couldn't retrieve document timestamp: %s", timestampErr))
				continue
			}

			validationError := identityService.ValidateDocumentSignature(
				signerAccountID,
				sig.PublicKey,
				ConsensusSignaturePayload(sr, sig.TransitionValidated),
				sig.Signature,
				timestamp,
			)

			if validationError != nil {
				err = errors.AppendError(
					err,
					errors.New("signature_%s verification failed: %v", hexutil.Encode(sig.SignerId), validationError),
				)
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
func anchoredValidator(anchorSrv anchors.API) Validator {
	return ValidatorFunc(func(_, model Document) error {
		if model == nil {
			return ErrModelNil
		}

		anchorID, err := anchors.ToAnchorID(model.CurrentVersion())
		if err != nil {
			return errors.NewTypedError(ErrAnchorIDCreation, err)
		}

		dr, err := model.CalculateDocumentRoot()
		if err != nil {
			return errors.NewTypedError(ErrDocumentCalculateDocumentRoot, err)
		}

		docRoot, err := anchors.ToDocumentRoot(dr)
		if err != nil {
			return errors.NewTypedError(ErrDocumentRootCreation, err)
		}

		gotRoot, anchoredAt, err := anchorSrv.GetAnchorData(anchorID)
		if err != nil {
			return errors.NewTypedError(
				ErrDocumentAnchorDataRetrieval,
				errors.New("failed to get document root for anchor %s from chain: %v", anchorID.String(), err),
			)
		}

		if !utils.IsSameByteSlice(docRoot[:], gotRoot[:]) {
			return ErrDocumentRootsMismatch
		}

		tm, err := model.Timestamp()
		if err != nil {
			return errors.NewTypedError(ErrDocumentTimestampRetrieval, err)
		}

		if tm.Add(MaxAuthoredToCommitDuration).Before(anchoredAt) {
			return errors.NewTypedError(
				ErrDocumentInvalidAnchorTime,
				errors.New("document was anchored after max allowed time for anchor %s", anchorID.String()),
			)
		}

		return nil
	})
}

// versionNotAnchoredValidator checks if the given version is not anchored on the chain.
// returns error if the version id is already anchored.
func versionNotAnchoredValidator(anchorSrv anchors.API, id []byte) error {
	anchorID, err := anchors.ToAnchorID(id)
	if err != nil {
		return errors.NewTypedError(ErrAnchorIDCreation, err)
	}

	_, _, err = anchorSrv.GetAnchorData(anchorID)
	if err == nil {
		return ErrDocumentIDReused
	}

	return nil
}

// LatestVersionValidator checks if the document is the latest version
func LatestVersionValidator(anchorSrv anchors.API) Validator {
	return ValidatorFunc(func(_, model Document) error {
		if model == nil {
			return ErrModelNil
		}

		err := versionNotAnchoredValidator(anchorSrv, model.NextVersion())
		if err != nil {
			return errors.NewTypedError(ErrDocumentNotLatest, err)
		}

		return nil
	})
}

// currentVersionValidator checks if the current version of the document has been anchored.
// returns an error if the current version has been anchored already.
func currentVersionValidator(anchorSrv anchors.API) Validator {
	return ValidatorFunc(func(_, model Document) error {
		if model == nil {
			return ErrModelNil
		}

		err := versionNotAnchoredValidator(anchorSrv, model.CurrentVersion())
		if err != nil {
			return errors.NewTypedError(ErrDocumentNotLatest, err)
		}

		return nil
	})
}

// attributeValidator validates the signed attributes.
func attributeValidator(identityService v2.Service) Validator {
	return ValidatorFunc(func(_, model Document) (err error) {
		if model == nil {
			return ErrModelNil
		}

		attrs := model.GetAttributes()
		for _, attr := range attrs {
			if attr.Value.Type != AttrSigned {
				continue
			}

			signed := attr.Value.Signed

			payload := attributeSignaturePayload(signed.Identity.ToBytes(), model.ID(), signed.DocumentVersion, signed.Value)

			timestamp, timestampErr := model.Timestamp()

			if timestampErr != nil {
				err = errors.AppendError(err, errors.New("couldn't get model timestamp: %s", timestampErr))
				continue
			}

			validationError := identityService.ValidateDocumentSignature(
				signed.Identity,
				signed.PublicKey,
				payload,
				signed.Signature,
				timestamp,
			)

			if validationError != nil {
				err = errors.AppendError(err, errors.New("failed to validate signature for attribute %s: %v", attr.KeyLabel, validationError))
			}
		}

		return err
	})
}

// transitionValidator checks that the document changes are within the transition_rule capability of the
// collaborator making the changes
func transitionValidator(collaborator *types.AccountID) Validator {
	return ValidatorFunc(func(old, new Document) error {
		if old == nil {
			return nil
		}

		err := old.CollaboratorCanUpdate(new, collaborator)
		if err != nil {
			return errors.NewTypedError(ErrInvalidDocumentStateTransition, err)
		}

		return nil
	})
}

// computeFieldsValidator verifies the execution of each compute field by re executing the WASM and checking the result
// is same as the one that is stored in the document.
func computeFieldsValidator(timeout time.Duration) Validator {
	return ValidatorFunc(func(_, new Document) error {
		computeFields := new.GetComputeFieldsRules()
		attributes := func() map[AttrKey]Attribute {
			attrMap := make(map[AttrKey]Attribute)
			for _, attr := range new.GetAttributes() {
				attrMap[attr.Key] = attr
			}
			return attrMap
		}()

		for _, computeField := range computeFields {
			// execute compute fields
			targetAttr, err := executeComputeField(computeField, attributes, timeout)
			if err != nil {
				return err
			}

			// verify the targetAttr is same as the one already stored
			attr := attributes[targetAttr.Key]
			if !reflect.DeepEqual(attr, targetAttr) {
				return errors.New("compute fields[%s] validation failed", hexutil.Encode(computeField.RuleKey))
			}
		}

		return nil
	})
}

// CreateVersionValidator validates if the new core document is properly derived from old one
func CreateVersionValidator(anchorSrv anchors.API) Validator {
	return ValidatorGroup{
		baseValidator(),
		currentVersionValidator(anchorSrv),
		LatestVersionValidator(anchorSrv),
	}
}

// UpdateVersionValidator validates if the new core document is properly derived from old one
func UpdateVersionValidator(anchorSrv anchors.API) Validator {
	return ValidatorGroup{
		versionIDsValidator(),
		currentVersionValidator(anchorSrv),
		LatestVersionValidator(anchorSrv),
	}
}

// PreAnchorValidator is a validator group with following validators
// base validator
// signing root validator
// document root validator
// signatures validator
// should be called before pre anchoring
func PreAnchorValidator(identityService v2.Service) Validator {
	return ValidatorGroup{
		SignatureValidator(identityService),
		documentRootValidator(),
	}
}

// PostAnchoredValidator is a validator group with following validators
// PreAnchorValidator
// anchoredValidator
// should be called after anchoring the document/when received anchored document
func PostAnchoredValidator(identityService v2.Service, anchorSrv anchors.API) Validator {
	return ValidatorGroup{
		PreAnchorValidator(identityService),
		anchoredValidator(anchorSrv),
		LatestVersionValidator(anchorSrv),
	}
}

// ReceivedAnchoredDocumentValidator is a validator group with following validators
// transitionValidator
// PostAnchoredValidator
func ReceivedAnchoredDocumentValidator(
	identityService v2.Service,
	anchorSrv anchors.API,
	collaborator *types.AccountID,
) Validator {
	return ValidatorGroup{
		transitionValidator(collaborator),
		PostAnchoredValidator(identityService, anchorSrv),
	}
}

// RequestDocumentSignatureValidator is a validator group with the following validators
// SignatureValidator
// transitionsValidator
// it should be called when a document is received over the p2p layer before signing
func RequestDocumentSignatureValidator(
	anchorSrv anchors.API,
	identityService v2.Service,
	collaborator *types.AccountID,
) Validator {
	return ValidatorGroup{
		documentTimestampForSigningValidator(),
		documentAuthorValidator(collaborator),
		currentVersionValidator(anchorSrv),
		LatestVersionValidator(anchorSrv),
		transitionValidator(collaborator),
		SignatureValidator(identityService),
	}
}

// SignatureValidator is a validator group with following validators
// baseValidator
// signingRootValidator
// signaturesValidator
// should be called after sender signing the document, before requesting the document and after signature collection
func SignatureValidator(identityService v2.Service) Validator {
	return ValidatorGroup{
		baseValidator(),
		signingRootValidator(),
		signaturesValidator(identityService),
		attributeValidator(identityService),
		computeFieldsValidator(computeFieldsTimeout),
	}
}
