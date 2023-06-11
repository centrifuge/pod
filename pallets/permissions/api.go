package permissions

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/validation"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("permissions_api")
)

const (
	ErrPermissionRolesNotFound  = errors.Error("permission roles not found")
	ErrPermissionRolesRetrieval = errors.Error("permission roles retrieval")
	ErrAccountIDEncoding        = errors.Error("account ID encoding")
	ErrPermissionScopeEncoding  = errors.Error("permission scope encoding")
)

const (
	PalletName            = "Permissions"
	PermissionStorageName = "Permission"
)

//go:generate mockery --name API --structname APIMock --filename api_mock.go --inpackage

type API interface {
	GetPermissionRoles(accountID *types.AccountID, poolID types.U64) (*PermissionRoles, error)
}

type api struct {
	centAPI centchain.API
}

func NewAPI(centAPI centchain.API) API {
	return &api{centAPI: centAPI}
}

func (a *api) GetPermissionRoles(accountID *types.AccountID, poolID types.U64) (*PermissionRoles, error) {
	err := validation.Validate(
		validation.NewValidator(accountID, validation.AccountIDValidationFn),
		validation.NewValidator(poolID, validation.U64ValidationFn),
	)

	if err != nil {
		log.Errorf("Validation error: %s", err)

		return nil, err
	}

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	encodedAccountID, err := codec.Encode(accountID)

	if err != nil {
		log.Errorf("Couldn't encode account ID: %s", err)

		return nil, ErrAccountIDEncoding
	}

	scope := PermissionScope{
		IsPool: true,
		AsPool: poolID,
	}

	encodedScope, err := codec.Encode(scope)

	if err != nil {
		log.Errorf("Couldn't encode permission scope: %s", err)

		return nil, ErrPermissionScopeEncoding
	}

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		PermissionStorageName,
		encodedAccountID,
		encodedScope,
	)

	if err != nil {
		log.Errorf("Couldn't create storage key: %s", err)

		return nil, errors.ErrStorageKeyCreation
	}

	var permissionRoles PermissionRoles

	ok, err := a.centAPI.GetStorageLatest(storageKey, &permissionRoles)

	if err != nil {
		log.Errorf("Couldn't retrieve permission roles from storage: %s", err)

		return nil, ErrPermissionRolesRetrieval
	}

	if !ok {
		log.Errorf("Permission roles not found for account ID %s and pool ID %d", accountID.ToHexString(), poolID)

		return nil, ErrPermissionRolesNotFound
	}

	return &permissionRoles, nil

}
