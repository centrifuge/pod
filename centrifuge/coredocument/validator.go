package coredocument

import (
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// UpdateVersionValidator validates if the new core document is properly derived from old one
func UpdateVersionValidator() documents.Validator {
	return documents.ValidatorFunc(func(old, new documents.Model) error {
		if old == nil || new == nil {
			return fmt.Errorf("need both the old and new model")
		}

		oldCD, err := old.PackCoreDocument()
		if err != nil {
			return fmt.Errorf("failed to fetch old core document: %v", err)
		}

		newCD, err := new.PackCoreDocument()
		if err != nil {
			return fmt.Errorf("failed to fetch new core document: %v", err)
		}

		fmt.Println(oldCD)
		fmt.Println("====")
		fmt.Println(newCD)
		checks := []struct {
			name string
			a, b []byte
		}{
			{
				name: "cd_document_identifier",
				a:    oldCD.DocumentIdentifier,
				b:    newCD.DocumentIdentifier,
			},

			{
				name: "cd_previous_version",
				a:    oldCD.CurrentVersion,
				b:    newCD.PreviousVersion,
			},

			{
				name: "cd_current_version",
				a:    oldCD.NextVersion,
				b:    newCD.CurrentVersion,
			},

			{
				name: "cd_previous_version",
				a:    oldCD.DocumentRoot,
				b:    newCD.PreviousRoot,
			},
		}

		for _, c := range checks {
			if !tools.CheckMultiple32BytesFilled(c.a, c.b) {
				err = documents.AppendError(err, documents.NewError(c.name, "missing identifiers"))
				continue
			}

			if !tools.IsSameByteSlice(c.a, c.b) {
				err = documents.AppendError(err, documents.NewError(c.name, "mismatched"))
			}
		}

		if tools.IsEmptyByteSlice(newCD.NextVersion) {
			err = documents.AppendError(err, documents.NewError("cd_next_version", centerrors.RequiredField))
		}

		return err
	})
}

// getCoreDocument takes an model and returns the core document of the model
func getCoreDocument(model documents.Model) (*coredocumentpb.CoreDocument, error) {
	if model == nil {
		return nil, fmt.Errorf("nil model")
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, fmt.Errorf("failed to get core document: %v", err)
	}

	return cd, nil
}

// baseValidator validates the core document basic fields like identifier, versions, and salts
func baseValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, model documents.Model) error {
		cd, err := getCoreDocument(model)
		if err != nil {
			return err
		}

		if err := Validate(cd); err != nil {
			return err
		}

		return nil
	})
}

// signingRootValidator checks the existence of signing root
// recalculates the signing root and compares with existing one
func signingRootValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, model documents.Model) error {
		cd, err := getCoreDocument(model)
		if err != nil {
			return err
		}

		if tools.IsEmptyByteSlice(cd.SigningRoot) {
			return fmt.Errorf("signing root missing")
		}

		tree, err := GetDocumentSigningTree(cd)
		if err != nil {
			return fmt.Errorf("failed to calculate signing root: %v", err)
		}

		if !tools.IsSameByteSlice(cd.SigningRoot, tree.RootHash()) {
			return fmt.Errorf("signing root mismatch")
		}

		return nil
	})
}

// documentRootValidator checks the existence of document root
// recalculates the document root and compares with existing one
func documentRootValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, model documents.Model) error {
		cd, err := getCoreDocument(model)
		if err != nil {
			return err
		}

		if tools.IsEmptyByteSlice(cd.DocumentRoot) {
			return fmt.Errorf("document root missing")
		}

		tree, err := GetDocumentRootTree(cd)
		if err != nil {
			return fmt.Errorf("failed to calculate document root: %v", err)
		}

		if !tools.IsSameByteSlice(cd.DocumentRoot, tree.RootHash()) {
			return fmt.Errorf("document root mismatch")
		}

		return nil
	})
}

// readyForSignaturesValidator validates self signature
// re-calculates the signature and compares with existing one
// assumes signing_root is already generated and verified
// Note: this needs to used only before document is sent for signatures from the collaborators
func readyForSignaturesValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, model documents.Model) error {
		cd, err := getCoreDocument(model)
		if err != nil {
			return err
		}

		if len(cd.Signatures) != 1 {
			return fmt.Errorf("expecting only one signature")
		}

		c, err := ed25519keys.GetIDConfig()
		if err != nil {
			return fmt.Errorf("failed to get keys for signature calculation: %v", err)
		}

		s := signatures.Sign(c, cd.SigningRoot)
		sh := cd.Signatures[0]
		if !tools.IsSameByteSlice(s.EntityId, sh.EntityId) {
			err = documents.AppendError(err, documents.NewError("cd_entity_id", "entity ID mismatch"))
		}

		if !tools.IsSameByteSlice(s.PublicKey, sh.PublicKey) {
			err = documents.AppendError(err, documents.NewError("cd_public_key", "public key mismatch"))
		}

		if !tools.IsSameByteSlice(s.Signature, sh.Signature) {
			err = documents.AppendError(err, documents.NewError("cd_signature", "signature mismatch"))
		}

		return err
	})
}

// signaturesValidator validates all the signatures in the core document
// assumes signing root is verified
// Note: can be used when during the signature request on collaborator side and post signature collection on sender side
// Note: this will break the current flow where we proceed to anchor even signatures verification fails
func signaturesValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, model documents.Model) error {
		cd, err := getCoreDocument(model)
		if err != nil {
			return err
		}

		if len(cd.Signatures) < 1 {
			return fmt.Errorf("atleast one signature expected")
		}

		for _, sig := range cd.Signatures {
			if errI := signatures.ValidateSignature(sig, cd.SigningRoot); errI != nil {
				err = documents.AppendError(
					err,
					documents.NewError(fmt.Sprintf("signature_%s", hexutil.Encode(sig.EntityId)), "signature verification failed"))
			}
		}

		return err
	})
}
