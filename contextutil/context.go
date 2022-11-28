package contextutil

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type contextKey string

const (
	// ErrSelfNotFound must be used when self value is not found in the context
	ErrSelfNotFound = errors.Error("self value not found in the context")

	self = contextKey("self")
)

// WithAccount returns a new context with the provided account and account identity as values.
func WithAccount(ctx context.Context, acc config.Account) context.Context {
	return context.WithValue(ctx, self, acc)
}

// Account extracts the TenantConfig from the given context value
func Account(ctx context.Context) (config.Account, error) {
	acc, ok := ctx.Value(self).(config.Account)
	if !ok {
		return nil, ErrSelfNotFound
	}
	return acc, nil
}

// Identity returns the identity from the context.
func Identity(ctx context.Context) (*types.AccountID, error) {
	acc, err := Account(ctx)

	if err != nil {
		return nil, err
	}

	return acc.GetIdentity(), nil
}
