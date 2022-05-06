package v3

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrAccountFromContextRetrieval = errors.Error("couldn't retrieve account from context")
	ErrKeyRingPairRetrieval        = errors.Error("couldn't retrieve key ring pair for account")
	ErrMetadataRetrieval           = errors.Error("couldn't retrieve latest metadata")
	ErrCallCreation                = errors.Error("couldn't create call")
	ErrSubmitAndWatchExtrinsic     = errors.Error("couldn't submit and watch extrinsic")
	ErrClassIDEncoding             = errors.Error("couldn't encode class ID")
	ErrInstanceIDEncoding          = errors.Error("couldn't encode instance ID")
	ErrClassIDDecoding             = errors.Error("couldn't decode class ID")
	ErrInstanceIDDecoding          = errors.Error("couldn't decode instance ID")
	ErrStorageKeyCreation          = errors.Error("couldn't create storage key")
	ErrClassDetailsRetrieval       = errors.Error("couldn't retrieve class details")
	ErrInstanceDetailsRetrieval    = errors.Error("couldn't retrieve instance details")
	ErrDocumentRetrieval           = errors.Error("couldn't retrieve document")
	ErrInstanceAlreadyMinted       = errors.Error("instance is already minted")
	ErrInstanceIDGeneration        = errors.Error("couldn't generate instance ID")
	ErrIdentityRetrieval           = errors.Error("couldn't retrieve identity")
	ErrMintJobDispatch             = errors.Error("couldn't dispatch NFT mint job")
	ErrInstanceDetailsNotFound     = errors.Error("instance details not found")
)
