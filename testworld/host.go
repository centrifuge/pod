// +build testworld

package tests

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/centrifuge/go-centrifuge/config"
	ctx "github.com/centrifuge/go-centrifuge/context"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/node"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("host")

type host struct {
	name, dir, ethNodeUrl, accountKeyPath, accountPassword, network,
	identityFactoryAddr, identityRegistryAddr, anchorRepositoryAddr, paymentObligationAddr string
	apiPort, p2pPort   int64
	bootstrapNodes     []string
	bootstrappedCtx    map[string]interface{}
	txPoolAccess       bool
	smartContractAddrs *config.SmartContractAddresses
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

func (h *host) Init() error {
	err := cmd.CreateConfig(h.dir, h.ethNodeUrl, h.accountKeyPath, h.accountPassword, h.network, h.apiPort, h.p2pPort, h.bootstrapNodes, h.txPoolAccess, h.smartContractAddrs)
	if err != nil {
		return err
	}
	m := ctx.MainBootstrapper{}
	m.PopulateBaseBootstrappers()
	h.bootstrappedCtx = map[string]interface{}{
		config.BootstrappedConfigFile: h.dir + "/config.yaml",
	}
	return m.Bootstrap(h.bootstrappedCtx)
}

func (h *host) Start(c context.Context) error {
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

func (h *host) CreateInvoice(inv invoice.Invoice, collaborators []string) {
	// TODO work
}
