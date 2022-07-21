package entityrelationship

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// fieldValidateFunc validates the fields of the entity relationship model
func fieldValidator(identityService v2.Service) documents.Validator {
	return documents.ValidatorFunc(func(_, new documents.Document) error {
		if new == nil {
			return documents.ErrDocumentNil
		}

		relationship, ok := new.(*EntityRelationship)
		if !ok {
			return documents.ErrDocumentInvalidType
		}

		identities := []*types.AccountID{relationship.Data.OwnerIdentity, relationship.Data.TargetIdentity}
		for _, identity := range identities {
			ctx := context.Background()
			err := identityService.ValidateAccount(ctx, identity)
			if err != nil {
				return documents.ErrIdentityInvalid
			}
		}

		return nil
	})
}
