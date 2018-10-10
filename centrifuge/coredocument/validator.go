package coredocument

import (
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
)

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

// fieldValidator validates the core document basic fields like identifier, versions, and salts
func fieldValidator() documents.Validator {
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
			return fmt.Errorf("failed to generate docuemnt root: %v", err)
		}

		if !tools.IsSameByteSlice(cd.DocumentRoot, tree.RootHash()) {
			return fmt.Errorf("document root mismatch")
		}

		return nil
	})
}
