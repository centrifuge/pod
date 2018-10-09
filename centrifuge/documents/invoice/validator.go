package invoice

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
)

// fieldValidateFunc validates the fields of the invoice model
func fieldValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, new documents.Model) error {
		if new == nil {
			return fmt.Errorf("nil document")
		}

		inv, ok := new.(*InvoiceModel)
		if !ok {
			return fmt.Errorf("unknown document type")
		}

		var err error
		if !documents.IsCurrencyValid(inv.Currency) {
			err = documents.AppendError(err, documents.NewError("inv_currency", "currency is invalid"))
		}

		return err
	})
}

// dataRootValidator calculates the data root and checks if it matches with the one on core document
func dataRootValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, model documents.Model) (err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("data root validation failed: %v", err)
			}
		}()

		if model == nil {
			return fmt.Errorf("nil document")
		}

		coreDoc, err := model.PackCoreDocument()
		if err != nil {
			return fmt.Errorf("failed to pack coredocument: %v", err)
		}

		if tools.IsEmptyByteSlice(coreDoc.DataRoot) {
			return fmt.Errorf("data root missing")
		}

		inv, ok := model.(*InvoiceModel)
		if !ok {
			return fmt.Errorf("unknown document type: %T", model)
		}

		if err = inv.calculateDataRoot(); err != nil {
			return fmt.Errorf("failed to calculate data root: %v", err)
		}

		if !tools.IsSameByteSlice(inv.CoreDocument.DataRoot, coreDoc.DataRoot) {
			return fmt.Errorf("mismatched data root")
		}

		return nil
	})
}

// CreateValidator is a validator group that should be run before creating the invoice and persisting it to DB
func CreateValidator() documents.ValidatorGroup {
	return documents.ValidatorGroup{
		fieldValidator(),
		dataRootValidator(),
	}
}
