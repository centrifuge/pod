package invoice

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
)

// fieldValidateFunc validates the fields of the invoice model
func fieldValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, new documents.Model) []error {
		if new == nil {
			return []error{fmt.Errorf("nil document")}
		}

		inv, ok := new.(*InvoiceModel)
		if !ok {
			return []error{fmt.Errorf("unknown document type")}
		}

		var errs []error
		if !documents.IsCurrencyValid(inv.Currency) {
			errs = append(errs, fmt.Errorf("currency is invalid"))
		}

		return errs
	})
}
