//go:build testworld

package host

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

type ControlUnit struct {
	cfg           config.Configuration
	serviceCtx    map[string]any
	bootstrappers []bootstrap.TestBootstrapper
}

func NewControlUnit(
	cfg config.Configuration,
	serviceCtx map[string]any,
	bootstrappers []bootstrap.TestBootstrapper,
) *ControlUnit {
	return &ControlUnit{
		cfg,
		serviceCtx,
		bootstrappers,
	}
}

func (c *ControlUnit) GetServiceCtx() map[string]any {
	return c.serviceCtx
}

func (c *ControlUnit) Start() error {
	for _, bootstrapper := range c.bootstrappers {
		if err := bootstrapper.TestBootstrap(c.serviceCtx); err != nil {
			return fmt.Errorf("couldn't start test host control unit: %w", err)
		}
	}

	return nil
}

func (c *ControlUnit) Stop() error {
	for _, bootstrapper := range c.bootstrappers {
		if err := bootstrapper.TestTearDown(); err != nil {
			return fmt.Errorf("couldn't stop test host control unit: %w", err)
		}
	}

	return nil
}
