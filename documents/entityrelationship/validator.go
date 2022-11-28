package entityrelationship

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
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

		if relationship.Data.OwnerIdentity == nil {
			return errors.NewTypedError(documents.ErrIdentityInvalid, errors.New("owner identity is nil"))
		}

		if relationship.Data.TargetIdentity == nil {
			return errors.NewTypedError(documents.ErrIdentityInvalid, errors.New("target identity is nil"))
		}

		identities := []*types.AccountID{relationship.Data.OwnerIdentity, relationship.Data.TargetIdentity}
		for _, identity := range identities {
			if err := identityService.ValidateAccount(identity); err != nil {
				return errors.NewTypedError(
					documents.ErrIdentityInvalid,
					fmt.Errorf("invalid account %s", identity.ToHexString()),
				)
			}
		}

		return nil
	})
}
