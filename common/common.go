// Package common holds the common types used across the node
package common

import "github.com/ethereum/go-ethereum/common"

var (
	// TenantKey is used as key for the tenant identity in the context.ContextWithValue.
	TenantKey struct{}

	// DummyIdentity to be used until we have identity coming from auth header
	// TODO(ved): get rid of this once we have multitenancy enabled
	DummyIdentity = common.Address([20]byte{1})
)
