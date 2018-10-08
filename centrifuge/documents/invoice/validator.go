package invoice

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
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
			err = documents.AppendError(err, fmt.Errorf("currency is invalid"))
		}

		return err
	})
}
