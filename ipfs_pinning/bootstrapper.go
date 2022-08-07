package ipfs_pinning

import (
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

const (
	BootstrappedIPFSPinningService = "BootstrappedIPFSPinningService"
)

var (
	supportedPinningServices = map[string]struct{}{
		"pinata": {},
	}
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, ok := ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	if !ok {
		return errors.New("couldn't find config")
	}

	pinningServiceName := cfg.GetIPFSPinningServiceName()

	if _, ok := supportedPinningServices[pinningServiceName]; !ok {
		return errors.New("unsupported pinning service")
	}

	pinningService, err := NewPinataServiceClient(
		cfg.GetIPFSPinningServiceURL(),
		cfg.GetIPFSPinningServiceAuth(),
	)

	if err != nil {
		return fmt.Errorf("couldn't create pinning service client: %w", err)
	}

	ctx[BootstrappedIPFSPinningService] = pinningService

	return nil
}
