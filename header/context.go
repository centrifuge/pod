package header

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
)

// ContextHeader holds custom request objects to pass down the pipeline.
type ContextHeader struct {
	context context.Context
	self    *identity.IDConfig
}

// NewContextHeader creates new instance of the request headers.
func NewContextHeader(context context.Context, config config.Config) (*ContextHeader, error) {
	idConfig, err := identity.GetIdentityConfig(config.(identity.Config))
	if err != nil {
		return nil, fmt.Errorf("failed to get id config: %v", err)
	}

	return &ContextHeader{self: idConfig, context: context}, nil
}

// Self returns Self CentID.
func (h *ContextHeader) Self() *identity.IDConfig {
	return h.self
}

// Context returns context.Context of the request.
func (h *ContextHeader) Context() context.Context {
	return h.context
}
