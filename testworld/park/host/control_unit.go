package test_host

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

type TestHostControlUnit struct {
	serviceCtx    map[string]any
	bootstrappers []bootstrap.TestBootstrapper
}

func NewTestHostControlUnit(
	serviceCtx map[string]any,
	bootstrappers []bootstrap.TestBootstrapper,
) *TestHostControlUnit {
	return &TestHostControlUnit{
		serviceCtx,
		bootstrappers,
	}
}

func (c *TestHostControlUnit) GetServiceCtx() map[string]any {
	return c.serviceCtx
}

func (c *TestHostControlUnit) Start() error {
	for _, bootstrapper := range c.bootstrappers {
		if err := bootstrapper.TestBootstrap(c.serviceCtx); err != nil {
			return fmt.Errorf("couldn't start test host control unit: %w", err)
		}
	}

	return nil
}

func (c *TestHostControlUnit) Stop() error {
	for _, bootstrapper := range c.bootstrappers {
		if err := bootstrapper.TestTearDown(); err != nil {
			return fmt.Errorf("couldn't stop test host control unit: %w", err)
		}
	}

	return nil
}
