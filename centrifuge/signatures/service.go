package signatures

import "github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"

type SigningService struct {
	KnownKeys [][]byte
}


// VerifySignatures verifies all signatures on the current document
func (srv *SigningService) VerifySignatures (doc *coredocument.CoreDocument) (err error, verified bool) {
	return nil, false

}