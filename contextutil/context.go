package contextutil

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/ethereum/go-ethereum/common/hexutil"

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

// New creates new instance of the request headers.
// TODO(ved): this doesn't deserve to be here. Causes too many implicit requirements
func New(ctx context.Context, cfg config.Account) (context.Context, error) {
	return context.WithValue(ctx, self, cfg), nil
}

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
	tcIDHex, ok := ctx.Value(config.AccountHeaderKey).(string)
	if !ok {
		return nil, errors.New("failed to get header %v", config.AccountHeaderKey)
	}

	tcID, err := hexutil.Decode(tcIDHex)
	if err != nil {
		return nil, errors.New("failed to get header: %v", err)
	}

	tc, err := cs.GetAccount(tcID)
	if err != nil {
		return nil, errors.New("failed to get header: %v", err)
	}

	ctxHeader, err := New(ctx, tc)
	if err != nil {
		return nil, errors.New("failed to get header: %v", err)
	}
	return ctxHeader, nil
}

// DIDFromContext returns did from the context.
func DIDFromContext(ctx context.Context) (did identity.DID, err error) {
	didStr, ok := ctx.Value(config.AccountHeaderKey).(string)
	if !ok {
		return did, ErrDIDMissingFromContext
	}

	didBytes, err := hexutil.Decode(didStr)
	if err != nil {
		return did, err
	}

	return identity.NewDIDFromBytes(didBytes)
}
