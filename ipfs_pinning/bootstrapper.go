package ipfs_pinning

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, ok := ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	if !ok {
		return errors.New("couldn't find config")
	}

	ctx[bootstrap.BootstrappedIPFSPinningService] = NewPinataServiceClient(
		cfg.GetIPFSPinningServiceURL(),
		cfg.GetIPFSPinningServiceJWT(),
	)

	return nil
}
