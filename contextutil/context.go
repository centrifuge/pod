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

// Context updates a context with account info using the configstore, must only be used for api handlers
func Context(ctx context.Context, cs config.Service) (context.Context, error) {
	ctxIdentity, ok := ctx.Value(config.AccountHeaderKey).(*types.AccountID)
	if !ok {
		return nil, errors.New("failed to get header %v", config.AccountHeaderKey)
	}

	acc, err := cs.GetAccount(ctxIdentity[:])
	if err != nil {
		return nil, errors.New("failed to get header: %v", err)
	}

	return WithAccount(ctx, acc), nil
}

// Identity returns the identity from the context.
func Identity(ctx context.Context) (*types.AccountID, error) {
	did, ok := ctx.Value(config.AccountHeaderKey).(*types.AccountID)
	if !ok {
		return did, ErrIdentityMissingFromContext
	}

	return did, nil
}
