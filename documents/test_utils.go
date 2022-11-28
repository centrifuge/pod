//go:build unit

package documents

import (
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func getTestCollaborators(count int) ([]*types.AccountID, error) {
	var collaborators []*types.AccountID

	for i := 0; i < count; i++ {
		accountID, err := testingcommons.GetRandomAccountID()
		if err != nil {
			return nil, err
		}

		collaborators = append(collaborators, accountID)
	}

	return collaborators, nil
}

func getTestSignatures(author *types.AccountID, collaborators []*types.AccountID) []*coredocumentpb.Signature {
	var signatures []*coredocumentpb.Signature

	authorSignature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            author.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: true,
	}

	signatures = append(signatures, authorSignature)

	for _, collaborator := range collaborators {
		collaboratorSignature := &coredocumentpb.Signature{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            collaborator.ToBytes(),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: false,
		}

		signatures = append(signatures, collaboratorSignature)
	}

	return signatures
}
