package v3

import "github.com/centrifuge/pod/errors"

const (
	ErrItemIDGeneration            = errors.Error("couldn't generate item ID")
	ErrMintJobDispatch             = errors.Error("couldn't dispatch NFT mint job")
	ErrDocumentRetrieval           = errors.Error("couldn't retrieve document")
	ErrCollectionIDDecoding        = errors.Error("couldn't decode collection ID")
	ErrItemIDDecoding              = errors.Error("couldn't decode item ID")
	ErrItemAlreadyMinted           = errors.Error("instance is already minted")
	ErrOwnerNotFound               = errors.Error("owner not found")
	ErrOwnerRetrieval              = errors.Error("couldn't retrieve owner")
	ErrCollectionCheck             = errors.Error("couldn't check if collection exists")
	ErrCollectionAlreadyExists     = errors.Error("collection already exists")
	ErrCreateCollectionJobDispatch = errors.Error("couldn't dispatch create collection job")
	ErrItemMetadataNotFound        = errors.Error("item metadata not found")
	ErrItemMetadataRetrieval       = errors.Error("couldn't retrieve item metadata")
	ErrItemAttributeNotFound       = errors.Error("item attribute not found")
	ErrItemAttributeRetrieval      = errors.Error("couldn't retrieve item attribute")
)
