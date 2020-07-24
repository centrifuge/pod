package entityrelationship

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

// fieldValidateFunc validates the fields of the entity relationship model
func fieldValidator(factory identity.Factory) documents.Validator {
	return documents.ValidatorFunc(func(_, new documents.Model) error {
		if new == nil {
			return documents.ErrDocumentNil
		}

		relationship, ok := new.(*EntityRelationship)
		if !ok {
			return documents.ErrDocumentInvalidType
		}

		identities := []*identity.DID{relationship.Data.OwnerIdentity, relationship.Data.TargetIdentity}
		for _, i := range identities {
			valid, err := factory.IdentityExists(i)
			if err != nil || !valid {
				return errors.New("identity not created from identity factory")
			}
		}

		return nil
	})
}

// CreateValidator returns a validator group that should be run before creating the entity and persisting it to DB
func CreateValidator(factory identity.Factory) documents.ValidatorGroup {
	return documents.ValidatorGroup{
		fieldValidator(factory),
	}
}

// UpdateValidator returns a validator group that should be run before updating the entity
func UpdateValidator(factory identity.Factory, anchorSrv anchors.Service) documents.ValidatorGroup {
	return documents.ValidatorGroup{
		fieldValidator(factory),
		documents.UpdateVersionValidator(anchorSrv),
	}
}
