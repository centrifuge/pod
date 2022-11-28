//go:build testworld

package host

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/storage"
)

type ControlUnit struct {
	cfg           config.Configuration
	serviceCtx    map[string]any
	bootstrappers []bootstrap.Bootstrapper

	p2pAddress string

	nodeCtx       context.Context
	nodeCtxCancel context.CancelFunc
	nodeErrChan   chan error
}

func NewControlUnit(
	cfg config.Configuration,
	serviceCtx map[string]any,
	bootstrappers []bootstrap.Bootstrapper,
	p2pAddress string,
) *ControlUnit {
	return &ControlUnit{
		cfg:           cfg,
		serviceCtx:    serviceCtx,
		bootstrappers: bootstrappers,
		p2pAddress:    p2pAddress,
	}
}

func (c *ControlUnit) GetPodCfg() config.Configuration {
	return c.cfg
}

func (c *ControlUnit) GetP2PAddress() string {
	return c.p2pAddress
}

func (c *ControlUnit) GetServiceCtx() map[string]any {
	return c.serviceCtx
}

func (c *ControlUnit) Start() error {
	for _, bootstrapper := range c.bootstrappers {
		if err := bootstrapper.Bootstrap(c.serviceCtx); err != nil {
			return fmt.Errorf("couldn't bootstrap %T: %w", bootstrapper, err)
		}
	}

	return c.startNode()
}

const (
	nodeStartErrorTimeout = 5 * time.Second
)

func (c *ControlUnit) startNode() error {
	nodeServers, err := node.GetServers(c.serviceCtx)

	if err != nil {
		return fmt.Errorf("couldn't get node servers: %w", err)
	}

	node := node.New(nodeServers)

	valueCtx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, c.serviceCtx)

	nodeCtx, nodeCtxCancel := context.WithCancel(valueCtx)

	nodeErrChan := make(chan error)

	go node.Start(nodeCtx, nodeErrChan)

	select {
	case err = <-nodeErrChan:
		nodeCtxCancel()
		return fmt.Errorf("couldn't start node: %w", err)
	case <-time.After(nodeStartErrorTimeout):
		// Node started successfully.
	}

	c.nodeCtx = nodeCtx
	c.nodeCtxCancel = nodeCtxCancel
	c.nodeErrChan = nodeErrChan

	return nil
}

const (
	nodeStopTimeout = 5 * time.Second
)

func (c *ControlUnit) Stop() error {
	c.nodeCtxCancel()

	c.serviceCtx[storage.BootstrappedDB].(storage.Repository).Close()
	c.serviceCtx[storage.BootstrappedConfigDB].(storage.Repository).Close()

	select {
	case err := <-c.nodeErrChan:
		if err != nil {
			return fmt.Errorf("node stop error: %w", err)
		}

		return nil
	case <-time.After(nodeStopTimeout):
		return errors.New("timeout reached when stopping the node")
	}
}
