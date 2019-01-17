package purchaseorder

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
)

// fieldValidateFunc validates the fields of the purchase order model
func fieldValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, new documents.Model) error {
		if new == nil {
			return errors.New("nil document")
		}

		po, ok := new.(*PurchaseOrder)
		if !ok {
			return errors.New("unknown document type")
		}

		var err error
		if !documents.IsCurrencyValid(po.Currency) {
			err = errors.AppendError(err, documents.NewError("po_currency", "currency is invalid"))
		}

		return err
	})
}

// dataRootValidator calculates the data root and checks if it matches with the one on core document
func dataRootValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, model documents.Model) (err error) {
		defer func() {
			if err != nil {
				err = errors.New("data root validation failed: %v", err)
			}
		}()

		if model == nil {
			return errors.New("nil document")
		}

		coreDoc, err := model.PackCoreDocument()
		if err != nil {
			return errors.New("failed to pack coredocument: %v", err)
		}

		if utils.IsEmptyByteSlice(coreDoc.DataRoot) {
			return errors.New("data root missing")
		}

		inv, ok := model.(*PurchaseOrder)
		if !ok {
			return errors.New("unknown document type: %T", model)
		}

		if err = inv.calculateDataRoot(); err != nil {
			return errors.New("failed to calculate data root: %v", err)
		}

		if !utils.IsSameByteSlice(inv.CoreDocument.DataRoot, coreDoc.DataRoot) {
			return errors.New("mismatched data root")
		}

		return nil
	})
}

// CreateValidator returns a validator group that should be run before creating the purchase order and persisting it to DB
func CreateValidator() documents.ValidatorGroup {
	return documents.ValidatorGroup{
		fieldValidator(),
		dataRootValidator(),
	}
}

// UpdateValidator returns a validator group that should be run before updating the purchase order
func UpdateValidator() documents.ValidatorGroup {
	return documents.ValidatorGroup{
		fieldValidator(),
		dataRootValidator(),
		documents.UpdateVersionValidator(),
	}
}
