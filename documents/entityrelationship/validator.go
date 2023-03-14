package entityrelationship

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
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
