package context

import (
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/config"
	"fmt"
	"context"
)

// Placeholder to pass custom request objects down the pipeline
type ContextHeader struct {
	context context.Context
	self *identity.IdentityConfig
}

// NewContextHeader creates new instance of the request headers needed
func NewContextHeader(context context.Context, config *config.Configuration) (*ContextHeader, error) {
	idConfig, err := identity.GetIdentityConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get id config: %v", err)
	}

	return &ContextHeader{self: idConfig, context: context}, nil
}

// Self returns Self CentID
func (h *ContextHeader) Self() *identity.IdentityConfig {
	return h.self
}

// Context returns context.Context of the request
func (h *ContextHeader) Context() context.Context {
	return h.context
}
