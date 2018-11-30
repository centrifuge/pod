// +build testworld

package testworld

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/centrifuge/go-centrifuge/config"
	ctx "github.com/centrifuge/go-centrifuge/context"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/gavv/httpexpect"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("host")

type robert struct {
	hosts map[string]*host
}

type host struct {
	name, dir, ethNodeUrl, accountKeyPath, accountPassword, network,
	identityFactoryAddr, identityRegistryAddr, anchorRepositoryAddr, paymentObligationAddr string
	apiPort, p2pPort   int64
	bootstrapNodes     []string
	bootstrappedCtx    map[string]interface{}
	txPoolAccess       bool
	smartContractAddrs *config.SmartContractAddresses
	config             config.Config
	identity           identity.Identity
	node               *node.Node
}

func newHost(
	name, ethNodeUrl, accountKeyPath, accountPassword, network string,
	apiPort, p2pPort int64,
	bootstraps []string,
	txPoolAccess bool,
	smartContractAddrs *config.SmartContractAddresses,
) *host {
	return &host{
		name:               name,
		ethNodeUrl:         ethNodeUrl,
		accountKeyPath:     accountKeyPath,
		accountPassword:    accountPassword,
		network:            network,
		apiPort:            apiPort,
		p2pPort:            p2pPort,
		bootstrapNodes:     bootstraps,
		txPoolAccess:       txPoolAccess,
		smartContractAddrs: smartContractAddrs,
		dir:                "peerconfigs/" + name,
	}
}

func (h *host) init() error {
	err := cmd.CreateConfig(h.dir, h.ethNodeUrl, h.accountKeyPath, h.accountPassword, h.network, h.apiPort, h.p2pPort, h.bootstrapNodes, h.txPoolAccess, h.smartContractAddrs)
	if err != nil {
		return err
	}
	m := ctx.MainBootstrapper{}
	m.PopulateBaseBootstrappers()
	h.bootstrappedCtx = map[string]interface{}{
		config.BootstrappedConfigFile: h.dir + "/config.yaml",
	}
	err = m.Bootstrap(h.bootstrappedCtx)
	if err != nil {
		return err
	}
	h.config = h.bootstrappedCtx[config.BootstrappedConfig].(config.Config)
	idService := h.bootstrappedCtx[identity.BootstrappedIDService].(identity.Service)
	idBytes, err := h.config.GetIdentityID()
	if err != nil {
		return err
	}
	id, err := identity.ToCentID(idBytes)
	if err != nil {
		return err
	}
	h.identity, err = idService.LookupIdentityForID(id)
	if err != nil {
		return err
	}
	return nil
}

func (h *host) start(c context.Context) error {
	srvs, err := node.GetServers(h.bootstrappedCtx)
	if err != nil {
		return fmt.Errorf("failed to load servers: %v", err)
	}

	h.node = node.New(srvs)
	feedback := make(chan error)
	// may be we can pass a context that exists in c here
	cancCtx, canc := context.WithCancel(context.WithValue(c, bootstrap.NodeObjRegistry, h.bootstrappedCtx))
	go h.node.Start(cancCtx, feedback)
	controlC := make(chan os.Signal, 1)
	signal.Notify(controlC, os.Interrupt)
	for {
		select {
		case err := <-feedback:
			log.Error(h.name+" panicking because of ", err)
			panic(err)
		case sig := <-controlC:
			log.Info(h.name+" shutting down because of ", sig)
			canc()
			err := <-feedback
			return err
		}
	}
}

func (h *host) createInvoice(e *httpexpect.Expect, inv map[string]interface{}) (*httpexpect.Object, error) {
	return createInvoice(e, inv), nil
}

func (h *host) createHttpExpectation(t *testing.T) *httpexpect.Expect {
	return CreateInsecureClient(t, fmt.Sprintf("https://localhost:%d", h.config.GetServerPort()))
}

func (h *host) id() (identity.CentID, error) {
	return h.identity.CentID(), nil
}

func (h *host) p2pURL() (string, error) {
	lastB58Key, err := h.identity.CurrentP2PKey()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/ipfs/%s", h.p2pPort, lastB58Key), nil
}
