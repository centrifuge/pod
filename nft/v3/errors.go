package v3

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrItemIDGeneration            = errors.Error("couldn't generate item ID")
	ErrMintJobDispatch             = errors.Error("couldn't dispatch NFT mint job")
	ErrDocumentRetrieval           = errors.Error("couldn't retrieve document")
	ErrCollectionIDDecoding        = errors.Error("couldn't decode collection ID")
	ErrItemIDDecoding              = errors.Error("couldn't decode item ID")
	ErrItemAlreadyMinted           = errors.Error("instance is already minted")
	ErrOwnerNotFound               = errors.Error("owner not found")
	ErrCollectionCheck             = errors.Error("couldn't check if collection exists")
	ErrCollectionAlreadyExists     = errors.Error("collection already exists")
	ErrCreateCollectionJobDispatch = errors.Error("couldn't dispatch create collection job")
	ErrPendingDocumentCommit       = errors.Error("couldn't commit pending document")
)
