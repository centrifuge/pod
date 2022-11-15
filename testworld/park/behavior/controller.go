//go:build testworld

package behavior

import (
	"context"
	"fmt"
	"testing"

	identityv2 "github.com/centrifuge/go-centrifuge/identity/v2"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	podBootstrap "github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/testworld/park/behavior/client"
	"github.com/centrifuge/go-centrifuge/testworld/park/behavior/webhook"
	"github.com/centrifuge/go-centrifuge/testworld/park/bootstrap"
	"github.com/centrifuge/go-centrifuge/testworld/park/factory"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-centrifuge/utils"
)

type Controller struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	hosts map[host.Name]*host.Host

	webhookReceiver *webhook.Receiver
}

func NewController() (*Controller, error) {
	ctx, cancel := context.WithCancel(context.Background())

	webhookReceiver, err := createWebhookReceiver()

	if err != nil {
		cancel()
		return nil, fmt.Errorf("couldn't create webhook receiver: %w", err)
	}

	return &Controller{
		ctx:             ctx,
		ctxCancel:       cancel,
		webhookReceiver: webhookReceiver,
	}, nil
}

func (h *Controller) GetWebhookReceiver() *webhook.Receiver {
	return h.webhookReceiver
}

func (h *Controller) GetClientForHost(t *testing.T, name host.Name) (*client.Client, error) {
	testHost, err := h.GetHost(name)

	if err != nil {
		return nil, err
	}

	hostJWT, err := testHost.GetMainAccount().GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])

	if err != nil {
		return nil, fmt.Errorf("couldn't get token for host: %w", err)
	}

	testClient := client.New(t, h.webhookReceiver, testHost.GetAPIURL(), hostJWT)

	return testClient, nil
}

func (h *Controller) CreateRandomAccountOnHost(name host.Name) (*host.Account, error) {
	testHost, err := h.GetHost(name)

	if err != nil {
		return nil, err
	}

	randomAccount, err := factory.CreateRandomHostAccount(
		testHost.GetServiceCtx(),
		h.webhookReceiver.GetURL(),
		testHost.GetMainAccount(),
	)

	if err != nil {
		return nil, fmt.Errorf("couldn't create random host account: %w", err)
	}

	if err := factory.AddTestHostAccountProxies(testHost.GetServiceCtx(), randomAccount); err != nil {
		return nil, fmt.Errorf("couldn't add test host account proxies: %w", err)
	}

	if err := identityv2.AddAccountKeysToStore(testHost.GetServiceCtx(), randomAccount.GetAccount()); err != nil {
		return nil, fmt.Errorf("couldn't add test account keys to store: %w", err)
	}

	return randomAccount, nil
}

func (h *Controller) GetHost(name host.Name) (*host.Host, error) {
	if host, ok := h.hosts[name]; ok {
		return host, nil
	}

	return nil, fmt.Errorf("host '%s' not found", name)
}

func (h *Controller) GetHosts() map[host.Name]*host.Host {
	return h.hosts
}

func (h *Controller) Start() error {
	_ = podBootstrap.RunTestBootstrappers([]podBootstrap.TestBootstrapper{&integration_test.Bootstrapper{}}, nil)

	go h.webhookReceiver.Start(h.ctx)

	testHosts, err := bootstrap.CreateTestHosts(h.webhookReceiver.GetURL())

	if err != nil {
		return fmt.Errorf("couldn't create test hosts: %w", err)
	}

	h.hosts = testHosts

	return nil
}

func (h *Controller) Stop() error {
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

func createWebhookReceiver() (*webhook.Receiver, error) {
	_, port, err := utils.GetFreeAddrPort()
	if err != nil {
		return nil, fmt.Errorf("couldn't get free port: %w", err)
	}

	return webhook.NewReceiver(port, webhookEndpoint), nil
}
