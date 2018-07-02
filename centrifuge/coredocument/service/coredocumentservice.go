package coredocumentservice

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/go-errors/errors"
)

func isEmptyCoreDocumentIdentifier(cdi []byte) bool {
	return tools.IsEmptyByteSlice(cdi)
}

func GenerateCoreDocumentIdentifier() (cdi []byte) {
	cdi32 := tools.RandomByte32()
	copy(cdi, cdi32[:])
	return
}

func checkForValidDocumentIdentifiers(document coredocumentpb.CoreDocument) ([]error) {
	errs := []error{}
	if isEmptyCoreDocumentIdentifier(document.DocumentIdentifier) && (!isEmptyCoreDocumentIdentifier(document.CurrentIdentifier) || !isEmptyCoreDocumentIdentifier(document.NextIdentifier)) {
		errs = append(errs, errors.New("No original document identifier found but document has current or next identifier."))
	}
	return errs
}

// FillCoreDocumentIdentifiers fills in empty document identifiers so that a CoreDocument is usable for sending, anchoring, etc
//func FillCoreDocumentIdentifiers(document coredocumentpb.CoreDocument) (coredocumentpb.CoreDocument, error) {
//	if isEmptyCoreDocumentIdentifier(document.DocumentIdentifier) && (!isEmptyCoreDocumentIdentifier(document.CurrentIdentifier) || !isEmptyCoreDocumentIdentifier(document.NextIdentifier)) {
//		return nil, errors.New("Document state is inconsistent. No original document identifier found but document has current or next identifier.")
//	}
//
//	if (isEmptyCoreDocumentIdentifier(document.DocumentIdentifier)) {
//		document.DocumentIdentifier = GenerateCoreDocumentIdentifier()
//	}
//	if (isEmptyCoreDocumentIdentifier(document.CurrentIdentifier)) {
//		document.DocumentIdentifier = GenerateCoreDocumentIdentifier()
//	}
//}
