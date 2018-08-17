package coredocument

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
)

// FillIdentifiers fills in missing identifiers for the given CoreDocument.
// It does checks on document consistency (e.g. re-using an old identifier).
// In the case of an error, it returns the error and an empty CoreDocument.
func FillIdentifiers(document coredocumentpb.CoreDocument) (coredocumentpb.CoreDocument, error) {
	isEmptyId := tools.IsEmptyByteSlice

	// check if the document identifier is empty
	if isEmptyId(document.DocumentIdentifier) {

		// check if current and next identifier are empty
		if !isEmptyId(document.CurrentIdentifier) {
			return document, errors.New(code.DocumentInvalid, "No DocumentIdentifier but has CurrentIdentifier")
		}

		// check if the next identifier is empty
		if !isEmptyId(document.NextIdentifier) {
			return document, errors.New(code.DocumentInvalid, "No CurrentIdentifier but has NextIdentifier")
		}

		// fill the identifiers
		document.DocumentIdentifier = tools.RandomSlice32()
		document.CurrentIdentifier = document.DocumentIdentifier
		document.NextIdentifier = tools.RandomSlice32()
		return document, nil
	}

	// check and fill current and next identifiers
	if isEmptyId(document.CurrentIdentifier) {
		document.CurrentIdentifier = document.DocumentIdentifier
	}

	if isEmptyId(document.NextIdentifier) {
		document.NextIdentifier = tools.RandomSlice32()
	}

	// double check the identifiers
	isSameBytes := tools.IsSameByteSlice

	// Problem (re-using an old identifier for NextIdentifier): CurrentIdentifier or DocumentIdentifier same as NextIdentifier
	if isSameBytes(document.NextIdentifier, document.DocumentIdentifier) ||
		isSameBytes(document.NextIdentifier, document.CurrentIdentifier) {
		return document, errors.New(code.DocumentInvalid, "Reusing old Identifier")
	}

	return document, nil
}
