package contextutil

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/errors"
)

type contextKey string

const (
	// ErrSelfNotFound must be used when self value is not found in the context
	ErrSelfNotFound = errors.Error("self value not found in the context")

	self = contextKey("self")
)

// NewCentrifugeContext creates new instance of the request headers.
func NewCentrifugeContext(ctx context.Context, cfg *configstore.TenantConfig) (context.Context, error) {
	return context.WithValue(ctx, self, cfg), nil
}

// Self returns Self CentID.
func Self(ctx context.Context) (*identity.IDConfig, error) {
	tc, ok := ctx.Value(self).(*configstore.TenantConfig)
	if !ok {
		return nil, ErrSelfNotFound
	}
	return identity.GetIdentityConfig(tc)
}

func Tenant(ctx context.Context) (*configstore.TenantConfig, error) {
	tc, ok := ctx.Value(self).(*configstore.TenantConfig)
	if !ok {
		return nil, ErrSelfNotFound
	}
	return tc, nil
}

// CentContext updates a context with tenant info using the configstore, must only be used for api handlers
func CentContext(config configstore.Service, ctx context.Context) (context.Context, error) {
	// TODO [multi-tenancy] remove following and read the tenantID from the context
	tc, err := config.GetAllTenants()
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}
	ctxHeader, err := NewCentrifugeContext(ctx, tc[0])
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}
	return ctxHeader, nil
}
