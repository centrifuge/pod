package multitenant

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context"
	"github.com/centrifuge/go-centrifuge/errors"
)

// BootstrappedTenants is a slice of bootstrapped tenants in the system
const BootstrappedTenants = "BootstrappedTenants"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap runs the servers.
// Note: this is a blocking call.
func (*Bootstrapper) Bootstrap(c map[string]interface{}) error {
	// TODO get tenant configs and for each tenant config(config.Repository#GetAllTenants) run tenantBootstrapper with the context map and add create the tenant using the resulting contexts
	srvs, err := GetServers(c)
	if err != nil {
		return errors.New("failed to load servers: %v", err)
	}

	cfg, ok := c[bootstrap.BootstrappedConfig].(config.Configuration)
	if !ok {
		return errors.New("config not initialised")
	}

	// TODO this is just a single tenant example
	tb := context.TenantBootstrapper{}
	tb.Populate()
	err = tb.Bootstrap(c)
	if err != nil {
		return errors.New("failed to bootstrap tenants: %v", err)
	}
	t := New(tb.IOCContext, cfg, srvs)

	ts := []*tenant{t}
	c[BootstrappedTenants] = ts
	return nil
}

// GetServers gets the long running background services in the node as a list
func GetServers(ctx map[string]interface{}) ([]Server, error) {
	p2pSrv, ok := ctx[bootstrap.BootstrappedP2PServer]
	if !ok {
		return nil, errors.New("p2p server not initialized")
	}

	var servers []Server
	servers = append(servers, p2pSrv.(Server))
	return servers, nil
}
