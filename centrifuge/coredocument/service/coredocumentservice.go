package coredocumentservice

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
)

func isEmptyId(cdi []byte) bool {
	return tools.IsEmptyByteSlice(cdi)
}

// GenerateCoreDocumentIdentifier generates a random identifier that is compatible to be used with the CoreDocument
// identifiers like DocumentIdentifier, CurrentIdentifier, NextIdentifier
func GenerateCoreDocumentIdentifier() []byte {
	cdi := make([]byte, 32)
	cdi32 := tools.RandomByte32()
	copy(cdi, cdi32[:32])
	return cdi
}

// AutoFillDocumentIdentifiers fills in missing identifiers for the given CoreDocument.
// It does checks on document consistency (e.g. re-using an old identifier).
// In the case of an error, it returns the error and an empty CoreDocument.
func AutoFillDocumentIdentifiers(document coredocumentpb.CoreDocument) (coredocumentpb.CoreDocument, error) {
	if isEmptyId(document.DocumentIdentifier) && isEmptyId(document.CurrentIdentifier) && isEmptyId(document.NextIdentifier) {
		document.DocumentIdentifier = GenerateCoreDocumentIdentifier()
		document.CurrentIdentifier = document.DocumentIdentifier
		document.NextIdentifier = GenerateCoreDocumentIdentifier()
		return document, nil
	} else if !isEmptyId(document.DocumentIdentifier) && isEmptyId(document.CurrentIdentifier) && isEmptyId(document.NextIdentifier) {
		document.CurrentIdentifier = document.DocumentIdentifier
		document.NextIdentifier = GenerateCoreDocumentIdentifier()
		return document, nil
	} else if !isEmptyId(document.DocumentIdentifier) && !isEmptyId(document.CurrentIdentifier) && isEmptyId(document.NextIdentifier) {
		document.NextIdentifier = GenerateCoreDocumentIdentifier()
		return document, nil
	} else if isEmptyId(document.DocumentIdentifier) && !isEmptyId(document.CurrentIdentifier) {
		return coredocumentpb.CoreDocument{}, coredocument.NewErrInconsistentState("No DocumentIdentifier but has CurrentIdentifier")
	} else if isEmptyId(document.CurrentIdentifier) && !isEmptyId(document.NextIdentifier) {
		return coredocumentpb.CoreDocument{}, coredocument.NewErrInconsistentState("No CurrentIdentifier but has NextIdentifier")
	} else if !isEmptyId(document.DocumentIdentifier) && !isEmptyId(document.CurrentIdentifier) && !isEmptyId(document.NextIdentifier) {
		// Good: all identifiers are different
		if !tools.IsSameByteSlice(document.DocumentIdentifier, document.CurrentIdentifier) && !tools.IsSameByteSlice(document.DocumentIdentifier, document.NextIdentifier) && !tools.IsSameByteSlice(document.CurrentIdentifier, document.NextIdentifier) {
			return document, nil
		}
		// Good: DocumentIdentifier == CurrentIdentifier but NextIdentifier is different
		if tools.IsSameByteSlice(document.DocumentIdentifier, document.CurrentIdentifier) && !tools.IsSameByteSlice(document.DocumentIdentifier, document.NextIdentifier) {
			return document, nil
		}
		// Problem (re-using an old identifier for NextIdentifier): CurrentIdentifier or DocumentIdentifier same as NextIdentifier
		if tools.IsSameByteSlice(document.NextIdentifier, document.DocumentIdentifier) || tools.IsSameByteSlice(document.NextIdentifier, document.CurrentIdentifier) {
			return coredocumentpb.CoreDocument{}, coredocument.NewErrInconsistentState("Reusing old Identifier")
		}
	}
	return coredocumentpb.CoreDocument{}, coredocument.NewErrInconsistentState("Unexpected state")
}
