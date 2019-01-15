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
func NewCentrifugeContext(ctx context.Context, cfg config.TenantConfiguration) (context.Context, error) {
	return context.WithValue(ctx, self, cfg), nil
}

// Self returns Self CentID.
func Self(ctx context.Context) (*identity.IDConfig, error) {
	tc, ok := ctx.Value(self).(config.TenantConfiguration)
	if !ok {
		return nil, ErrSelfNotFound
	}
	return identity.GetIdentityConfig(tc)
}

// Tenant extracts the TenanConfig from the given context value
func Tenant(ctx context.Context) (config.TenantConfiguration, error) {
	tc, ok := ctx.Value(self).(config.TenantConfiguration)
	if !ok {
		return nil, ErrSelfNotFound
	}
	return tc, nil
}

// NodeContext updates a context with tenant info using the configstore, must only be used for api handlers
func NodeContext(ctx context.Context, cs config.Service) (context.Context, error) {
	tcIDHex := ctx.Value(config.TenantKey).(string)
	tcID, err := hexutil.Decode(tcIDHex)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}

	tc, err := cs.GetTenant(tcID)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}

	ctxHeader, err := NewCentrifugeContext(ctx, tc)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}
	return ctxHeader, nil
}
