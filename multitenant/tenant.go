package multitenant

import (
	"context"
	"sync"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("tenant")

// Config is configuration for a single tenant
type Config interface {
	GetEthereumAccount(accountName string) (account *config.AccountConfig, err error)
	GetEthereumDefaultAccountName() string
	GetReceiveEventNotificationEndpoint() string
	GetIdentityID() ([]byte, error)
	GetSigningKeyPair() (pub, priv string)
	GetEthAuthKeyPair() (pub, priv string)
}

// Server interface must be implemented by all tenant specific background servers on Cent Node
type Server interface {

	// Name is the unique name given to the service within the Cent Node
	Name() string

	// Start starts the service, expectation is that this would always be called in a separate go routine.
	// WaitGroup contract must always be honoured by calling `defer wg.Done()`
	Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error)
}

// tenant is a tenant in the multi tenant centrifuge node
type tenant struct {

	// iocContext is the inversion of control context used for the tenant at the bootstrap stage, its a copy of the map with references to,
	// objects initialised with MainBootstrapper + objects initialised using TenantBootstrapper
	iocContext map[string]interface{}

	// config is the Tenants config
	config Config

	// servers are the long running services running per tenant which must be kept to a minimum
	servers []Server
}

// New creates a new tenant in the cent node given the config and the servers to initiate
func New(iocContext map[string]interface{}, config Config, servers []Server) *tenant {
	return &tenant{iocContext: iocContext, config: config, servers: servers}
}

func (t *tenant) ID() (identity.CentID, error) {
	idBytes, err := t.config.GetIdentityID()
	if err != nil {
		return identity.CentID{}, err
	}
	return identity.ToCentID(idBytes)
}

func (t *tenant) Config() Config {
	return t.config
}

func (t *tenant) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()
	// check and get ID
	tid, err := t.ID()
	if err != nil {
		startupErr <- err
		return
	}
	ctxCh, cancel := context.WithCancel(ctx)
	defer cancel()
	childErr := make(chan error)
	var twg sync.WaitGroup
	twg.Add(len(t.servers))
	for _, s := range t.servers {
		go s.Start(ctxCh, &twg, childErr)
	}

	select {
	case errOut := <-childErr:
		log.Errorf("Tenant %s received error from child service, stopping all child servers %v", tid, errOut)
		// if one of the children fails to start all should stop
		cancel()
		// send the error upstream
		startupErr <- errOut
		twg.Wait()
		return
	case <-ctx.Done():
		log.Infof("Tenant %s received context.done signal, stopping all child servers", tid)
		// Note that in this case the children will also receive the done signal via the passed on context
		twg.Wait()
		log.Infof("Tenant %s stopped all child servers", tid)
		// special case to make the caller wait until servers are shutdown
		startupErr <- nil
		return
	}
}

// TODO create methods in Tenant to access specific methods for document services and others using iocContext
