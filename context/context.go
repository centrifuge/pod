package context

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

// Header holds custom request objects to pass down the pipeline.
type Header struct {
	context context.Context
	self    *identity.IDConfig
}

// NewContextHeader creates new instance of the request headers.
func NewContextHeader(context context.Context, config config.Configuration) (*Header, error) {
	idConfig, err := identity.GetIdentityConfig(config.(identity.Config))
	if err != nil {
		return nil, errors.New("failed to get id config: %v", err)
	}

	return &Header{self: idConfig, context: context}, nil
}

// Self returns Self CentID.
func (h *Header) Self() *identity.IDConfig {
	return h.self
}

// Context returns context.Context of the request.
func (h *Header) Context() context.Context {
	return h.context
}
