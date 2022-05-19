package ipfs

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, ok := ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	if !ok {
		return errors.New("couldn't find config")
	}

	service, err := New(context.Background(), cfg)

	if err != nil {
		return fmt.Errorf("couldn't create IPFS service: %w", err)
	}

	ctx[bootstrap.BootstrappedIPFSService] = service

	return nil
}
