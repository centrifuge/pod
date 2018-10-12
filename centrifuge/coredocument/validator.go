package coredocument

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
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

		isEqual := tools.IsSameByteSlice
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
			if !isEqual(c.a, c.b) {
				err = documents.AppendError(err, documents.NewError(c.name, "mismatched"))
			}
		}

		if tools.IsEmptyByteSlice(newCD.NextVersion) {
			err = documents.AppendError(err, documents.NewError("cd_next_version", centerrors.RequiredField))
		}

		return err
	})
}
