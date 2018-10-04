package invoice

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
)

// validator implements documents.Validator
type validator func(documents.Model, documents.Model) []error

func (v validator) Validate(old, new documents.Model) []error {
	return v(old, new)
}

// fieldValidateFunc validates the fields of the invoice model
func fieldValidator() documents.Validator {
	return validator(func(_, new documents.Model) []error {
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

		if !identity.IsCentIDValid(inv.Payee) {
			errs = append(errs, fmt.Errorf("payee is invalid"))
		}

		return errs
	})
}
