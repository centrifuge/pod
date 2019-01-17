package contextutil

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/errors"
)

type contextKey string

const (
	// ErrSelfNotFound must be used when self value is not found in the context
	ErrSelfNotFound = errors.Error("self value not found in the context")

	self = contextKey("self")
)

// NewCentrifugeContext creates new instance of the request headers.
func NewCentrifugeContext(ctx context.Context, cfg config.Account) (context.Context, error) {
	return context.WithValue(ctx, self, cfg), nil
}

// Self returns Self CentID.
func Self(ctx context.Context) (*identity.IDConfig, error) {
	tc, ok := ctx.Value(self).(config.Account)
	if !ok {
		return nil, ErrSelfNotFound
	}
	return identity.GetIdentityConfig(tc)
}

// Account extracts the TenanConfig from the given context value
func Account(ctx context.Context) (config.Account, error) {
	tc, ok := ctx.Value(self).(config.Account)
	if !ok {
		return nil, ErrSelfNotFound
	}
	return tc, nil
}

// Context updates a context with account info using the configstore, must only be used for api handlers
func Context(ctx context.Context, cs config.Service) (context.Context, error) {
	tcIDHex, ok := ctx.Value(config.AccountHeaderKey).(string)
	if !ok {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header %v", config.AccountHeaderKey))
	}

	tcID, err := hexutil.Decode(tcIDHex)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}

	tc, err := cs.GetAccount(tcID)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}

	ctxHeader, err := NewCentrifugeContext(ctx, tc)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}
	return ctxHeader, nil
}
