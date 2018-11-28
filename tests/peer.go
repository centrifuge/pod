// +build integration

package tests

import (
	"github.com/centrifuge/go-centrifuge/node"
		"github.com/centrifuge/go-centrifuge/config"
	ctx "github.com/centrifuge/go-centrifuge/context"
	"fmt"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"os"
	"os/signal"
	"context"
	logging "github.com/ipfs/go-log"
	"github.com/centrifuge/go-centrifuge/cmd"
)

var log = logging.Logger("peer")

type peer struct {
	name, dir, ethNodeUrl, accountKeyPath, accountPassword, network string
	apiPort, p2pPort int64
	bootstrapNodes []string
	bootstrappedCtx map[string]interface{}
	txPoolAccess bool
	node *node.Node
}

func NewPeer(name, ethNodeUrl, accountKeyPath, accountPassword, network string,
	apiPort,
	p2pPort int64,
	bootstraps []string,
	txPoolAccess bool) *peer {
	return &peer{
		name: name,
		ethNodeUrl: ethNodeUrl,
		accountKeyPath: accountKeyPath,
		accountPassword: accountPassword,
		network: network,
		apiPort: apiPort,
		p2pPort: p2pPort,
		bootstrapNodes: bootstraps,
		txPoolAccess: txPoolAccess,
		dir: "peerconfigs/" + name,
	}
}

func (p *peer) Init() error {
	err := cmd.CreateConfig(p.dir, p.ethNodeUrl, p.accountKeyPath, p.accountPassword, p.network, p.apiPort, p.p2pPort, p.bootstrapNodes, p.txPoolAccess)
	if err != nil {
		return err
	}
	m := ctx.MainBootstrapper{}
	m.PopulateBaseBootstrappers()
	p.bootstrappedCtx = map[string]interface{}{
		config.BootstrappedConfigFile: "<conf filename>",
	}
	return m.Bootstrap(p.bootstrappedCtx)
}

func (p *peer) Start(c context.Context) error {
	srvs, err := node.GetServers(p.bootstrappedCtx)
	if err != nil {
		return fmt.Errorf("failed to load servers: %v", err)
	}

	p.node = node.New(srvs)
	feedback := make(chan error)
	// may be we can pass a context that exists in c here
	cancCtx, canc := context.WithCancel(context.WithValue(c, bootstrap.NodeObjRegistry, p.bootstrappedCtx))
	go p.node.Start(cancCtx, feedback)
	controlC := make(chan os.Signal, 1)
	signal.Notify(controlC, os.Interrupt)
	for {
		select {
		case err := <-feedback:
			log.Error(p.name + " panicking because of ", err)
			panic(err)
		case sig := <-controlC:
			log.Info(p.name + " shutting down because of ", sig)
			canc()
			err := <-feedback
			return err
		}
	}
}
