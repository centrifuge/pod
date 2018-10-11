package documents

import (
	"fmt"
	"strings"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
)

// CreateProofs util function that takes document data tree, coreDocument and a list fo fields and generates proofs
func CreateProofs(dataTree *proofs.DocumentTree, coreDoc *coredocumentpb.CoreDocument, fields []string) (proofs []*proofspb.Proof, err error) {
	dataRootHashes, err := coredocument.GetDataProofHashes(coreDoc)
	if err != nil {
		return nil, fmt.Errorf("createProofs error %v", err)
	}

	signingRootHashes, err := coredocument.GetSigningProofHashes(coreDoc)
	if err != nil {
		return nil, fmt.Errorf("createProofs error %v", err)
	}

	cdtree, err := coredocument.GetDocumentSigningTree(coreDoc)
	if err != nil {
		return nil, fmt.Errorf("createProofs error %v", err)
	}

	// We support fields that belong to different document trees, as we do not prepend a tree prefix to the field, the approach
	// is to try in both trees to find the field and create the proof accordingly
	for _, field := range fields {
		rootHashes := dataRootHashes
		proof, err := dataTree.CreateProof(field)
		if err != nil {
			if strings.Contains(err.Error(), "No such field") {
				proof, err = cdtree.CreateProof(field)
				if err != nil {
					return nil, fmt.Errorf("createProofs error %v", err)
				}
				rootHashes = signingRootHashes
			} else {
				return nil, fmt.Errorf("createProofs error %v", err)
			}
		}
		proof.SortedHashes = append(proof.SortedHashes, rootHashes...)
		proofs = append(proofs, &proof)
	}
	return
}
