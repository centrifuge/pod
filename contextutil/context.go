package contextutil

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

type contextKey string

const (
	// ErrSelfNotFound must be used when self value is not found in the context
	ErrSelfNotFound = errors.Error("self value not found in the context")

	// ErrDIDMissingFromContext sentinel error when did is missing from the context.
	ErrDIDMissingFromContext = errors.Error("failed to extract did from context")

	self = contextKey("self")
)

// WithAccount sets config to the context and returns it
func WithAccount(ctx context.Context, cfg config.Account) context.Context {
	return context.WithValue(ctx, self, cfg)
}

// AccountDID extracts the AccountConfig DID from the given context value
func AccountDID(ctx context.Context) (identity.DID, error) {
	acc, err := Account(ctx)
	if err != nil {
		return identity.DID{}, err
	}
	did := acc.GetIdentity()

	return did, nil
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
	ctxIdentity, ok := ctx.Value(config.AccountHeaderKey).(identity.DID)
	if !ok {
		return nil, errors.New("failed to get header %v", config.AccountHeaderKey)
	}

	acc, err := cs.GetAccount(ctxIdentity[:])
	if err != nil {
		return nil, errors.New("failed to get header: %v", err)
	}

	return WithAccount(ctx, acc), nil
}

// DIDFromContext returns did from the context.
func DIDFromContext(ctx context.Context) (did identity.DID, err error) {
	did, ok := ctx.Value(config.AccountHeaderKey).(identity.DID)
	if !ok {
		return did, ErrDIDMissingFromContext
	}

	return did, nil
}
