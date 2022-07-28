package contextutil

import (
	"context"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
)

type contextKey string

const (
	// ErrSelfNotFound must be used when self value is not found in the context
	ErrSelfNotFound = errors.Error("self value not found in the context")

	// ErrIdentityMissingFromContext sentinel error when identity is missing from the context.
	ErrIdentityMissingFromContext = errors.Error("failed to extract identity from context")

	self = contextKey("self")
)

// WithAccount sets config to the context and returns it
func WithAccount(ctx context.Context, cfg config.Account) context.Context {
	return context.WithValue(ctx, self, cfg)
}

// Account extracts the TenantConfig from the given context value
func Account(ctx context.Context) (config.Account, error) {
	tc, ok := ctx.Value(self).(config.Account)
	if !ok {
		return nil, ErrSelfNotFound
	}
	return tc, nil
}

// Identity returns the identity from the context.
func Identity(ctx context.Context) (*types.AccountID, error) {
	acc, err := Account(ctx)

	if err != nil {
		return nil, err
	}

	return acc.GetIdentity(), nil
}
