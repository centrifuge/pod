package coreapi

import (
	logging "github.com/ipfs/go-log"
)

const (
	// DocumentIDParam for document_id in api path.
	DocumentIDParam = "document_id"

	// VersionIDParam for version_id in api path.
	VersionIDParam = "version_id"

	// AccountIDParam for accounts api.
	AccountIDParam = "account_id"

	// CollectionIDParam for NFT V3 api.
	CollectionIDParam = "collection_id"

	// ItemIDParam for NFT V3 api.
	ItemIDParam = "item_id"

	// AttributeNameParam for NFT V3 api.
	AttributeNameParam = "attribute_name"

	// AssetIDNameParam for Investor api.
	AssetIDNameParam = "asset_id"

	// LoanIDNameParam for Investor api.
	LoanIDNameParam = "loan_id"

	// PoolIDNameParam for Investor api.
	PoolIDNameParam = "pool_id"
)

var log = logging.Logger("core_api")
