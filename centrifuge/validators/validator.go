package validators

// Validator defines function required for the Document validations
// This will be implemented by all types of documents.
// Ex: Core document, invoice, purchase order etc...
// valid: if true, document is valid
// errMsg: error message if invalid
// errors: sub-errors if there are any
type Validator interface {
	Validate() (valid bool, errMsg string, errors map[string]string)
}
