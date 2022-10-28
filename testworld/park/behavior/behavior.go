//go:build testworld

package behavior

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/testworld/park/bootstrap"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/testworld/park/host"
)

type Head struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	hosts map[host.Name]*host.Host

	webhookReceiver *webhookReceiver
}

func NewHead() (*Head, error) {
	ctx, cancel := context.WithCancel(context.Background())

	webhookReceiver, err := createWebhookReceiver()

	if err != nil {
		cancel()
		return nil, fmt.Errorf("couldn't create webhook receiver: %w", err)
	}

	return &Head{
		ctx:             ctx,
		ctxCancel:       cancel,
		webhookReceiver: webhookReceiver,
	}, nil
}

func (h *Head) Start() error {
	go h.webhookReceiver.start(h.ctx)

	testHosts, err := bootstrap.CreateTestHosts(h.webhookReceiver.url())

	if err != nil {
		return fmt.Errorf("couldn't create test hosts: %w", err)
	}

	h.hosts = testHosts

	return nil
}

func (h *Head) Stop() error {
	h.ctxCancel()

	for hostName, host := range h.hosts {
		if err := host.Stop(); err != nil {
			return fmt.Errorf("couldn't stop host %s: %w", hostName, err)
		}
	}

	return nil
}

const (
	webhookEndpoint = "/webhook"
)

func createWebhookReceiver() (*webhookReceiver, error) {
	_, port, err := utils.GetFreeAddrPort()
	if err != nil {
		return nil, fmt.Errorf("couldn't get free port: %w", err)
	}

	return newWebhookReceiver(port, webhookEndpoint), nil
}
