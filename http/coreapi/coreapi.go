package coreapi

import (
	logging "github.com/ipfs/go-log"
)

const (
	// DocumentIDParam for document_id in api path.
	DocumentIDParam = "document_id"

	// VersionIDParam for version_id in api path.
	VersionIDParam = "version_id"

	// TokenIDParam for nft tokenID
	TokenIDParam = "token_id"

	// RegistryAddressParam for nft registry path
	RegistryAddressParam = "registry_address"

	// AccountIDParam for accounts api
	AccountIDParam = "account_id"
)

var log = logging.Logger("core_api")
