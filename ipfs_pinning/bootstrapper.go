package ipfs_pinning

import (
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, ok := ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	if !ok {
		return errors.New("couldn't find config")
	}

	pinningService, err := NewPinataServiceClient(
		cfg.GetIPFSPinningServiceURL(),
		cfg.GetIPFSPinningServiceJWT(),
	)

	if err != nil {
		return fmt.Errorf("couldn't create pinning service client: %w", err)
	}

	ctx[bootstrap.BootstrappedIPFSPinningService] = pinningService

	return nil
}
