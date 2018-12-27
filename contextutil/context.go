package contextutil

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

const ErrIDConfigNotFound = errors.Error("id configuration wasn't found")
const ErrSelfNotFound = errors.Error("self value not found in the context")

// NewCentrifugeContext creates new instance of the request headers.
func NewCentrifugeContext(ctx context.Context, config config.Configuration) (context.Context, error) {
	idConfig, err := identity.GetIdentityConfig(config.(identity.Config))
	if err != nil {
		return nil, errors.NewTypedError(ErrIDConfigNotFound, errors.New("%v", err))
	}
	return context.WithValue(ctx, "self", idConfig), nil
}

// Self returns Self CentID.
func Self(ctx context.Context) (*identity.IDConfig, error) {
	self, ok := ctx.Value("self").(*identity.IDConfig)
	if ok {
		return self, nil
	}
	return nil, ErrSelfNotFound
}
