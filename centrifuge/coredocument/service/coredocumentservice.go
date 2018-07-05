package coredocumentservice

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
)

func isEmptyId(cdi []byte) bool {
	return tools.IsEmptyByteSlice(cdi)
}

func isSame(cdi []byte) bool {
	return tools.IsEmptyByteSlice(cdi)
}

func GenerateCoreDocumentIdentifier() (cdi []byte) {
	cdi32 := tools.RandomByte32()
	copy(cdi, cdi32[:])
	return
}

func AutoFillDocumentIdentifiers(document coredocumentpb.CoreDocument) (error) {
	if isEmptyId(document.DocumentIdentifier) && isEmptyId(document.CurrentIdentifier) && isEmptyId(document.NextIdentifier) {
		return nil
	} else if !isEmptyId(document.DocumentIdentifier) && isEmptyId(document.CurrentIdentifier) && isEmptyId(document.NextIdentifier) {
		return nil
	} else if !isEmptyId(document.DocumentIdentifier) && !isEmptyId(document.CurrentIdentifier) && isEmptyId(document.NextIdentifier) {
		return nil
	} else if isEmptyId(document.DocumentIdentifier) && !isEmptyId(document.CurrentIdentifier) {
		return coredocument.NewErrInconsistentState("No DocumentIdentifier but has CurrentIdentifier")
	} else if isEmptyId(document.CurrentIdentifier) && !isEmptyId(document.NextIdentifier) {
		return coredocument.NewErrInconsistentState("No CurrentIdentifier but has NextIdentifier")
	} else if !isEmptyId(document.DocumentIdentifier) && !isEmptyId(document.CurrentIdentifier) && !isEmptyId(document.NextIdentifier) {
		// Good: all identifiers are different
		if !tools.IsSameByteSlice(document.DocumentIdentifier, document.CurrentIdentifier) && !tools.IsSameByteSlice(document.DocumentIdentifier, document.NextIdentifier) && !tools.IsSameByteSlice(document.CurrentIdentifier, document.NextIdentifier){
			return nil
		}
		// Good: DocumentIdentifier == CurrentIdentifier but NextIdentifier is different
		if tools.IsSameByteSlice(document.DocumentIdentifier, document.CurrentIdentifier) && !tools.IsSameByteSlice(document.DocumentIdentifier, document.NextIdentifier){
			return nil
		}
		// Problem (re-using an old identifier for NextIdentifier): CurrentIdentifier or DocumentIdentifier same as NextIdentifier
		if tools.IsSameByteSlice(document.NextIdentifier, document.DocumentIdentifier) || tools.IsSameByteSlice(document.NextIdentifier, document.CurrentIdentifier) {
			return coredocument.NewErrInconsistentState("Reusing old Identifier")
		}
	}
	return coredocument.NewErrInconsistentState("Unexpected state")
}

// FillCoreDocumentIdentifiers fills in empty document identifiers so that a CoreDocument is usable for sending, anchoring, etc
//func FillCoreDocumentIdentifiers(document coredocumentpb.CoreDocument) (coredocumentpb.CoreDocument, error) {
//	if isEmptyId(document.DocumentIdentifier) && (!isEmptyId(document.CurrentIdentifier) || !isEmptyId(document.NextIdentifier)) {
//		return nil, errors.New("Document state is inconsistent. No original document identifier found but document has current or next identifier.")
//	}
//
//	if (isEmptyId(document.DocumentIdentifier)) {
//		document.DocumentIdentifier = GenerateCoreDocumentIdentifier()
//	}
//	if (isEmptyId(document.CurrentIdentifier)) {
//		document.DocumentIdentifier = GenerateCoreDocumentIdentifier()
//	}
//}
