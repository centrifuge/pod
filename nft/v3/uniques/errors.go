package uniques

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrCollectionIDEncoding       = errors.Error("couldn't encode collection ID")
	ErrItemIDEncoding             = errors.Error("couldn't encode item ID")
	ErrCollectionDetailsRetrieval = errors.Error("couldn't retrieve collection details")
	ErrCollectionDetailsNotFound  = errors.Error("collection details not found")
	ErrItemMetadataRetrieval      = errors.Error("couldn't retrieve item metadata")
	ErrItemMetadataNotFound       = errors.Error("item metadata not found")
	ErrAdminMultiAddressCreation  = errors.Error("couldn't create admin multi address")
	ErrOwnerMultiAddressCreation  = errors.Error("couldn't create owner multi address")
	ErrItemDetailsRetrieval       = errors.Error("couldn't retrieve item details")
	ErrItemDetailsNotFound        = errors.Error("item details not found")
)
