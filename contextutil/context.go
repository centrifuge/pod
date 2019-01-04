package contextutil

import (
	"context"

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
