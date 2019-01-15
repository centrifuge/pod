// +build unit

package p2p

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"

	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

var (
	cfg       config.Service
	idService identity.Service
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		&queue.Bootstrapper{},
		transactions.Bootstrapper{},
		documents.Bootstrapper{},
	}
	ctx := make(map[string]interface{})
	idService = &testingcommons.MockIDService{}
	ctx[identity.BootstrappedIDService] = idService
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[config.BootstrappedConfigStorage].(config.Service)
	c, _ := cfg.GetConfig()
	n := c.(*configstore.NodeConfig)
	n.MainIdentity.SigningKeyPair.Pub = "../build/resources/signingKey.pub.pem"
	n.MainIdentity.SigningKeyPair.Priv = "../build/resources/signingKey.key.pem"
	n.MainIdentity.EthAuthKeyPair.Pub = "../build/resources/ethauth.pub.pem"
	n.MainIdentity.EthAuthKeyPair.Priv = "../build/resources/ethauth.key.pem"
	cfg.UpdateConfig(c)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestCentP2PServer_StartContextCancel(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	n := c.(*configstore.NodeConfig)
	n.P2PPort = 38203
	_, err = cfg.UpdateConfig(c)
	assert.NoError(t, err)
	cp2p := &peer{config: cfg, handlerCreator: func() *receiver.Handler {
		return receiver.New(cfg, nil, receiver.HandshakeValidator(n.NetworkID, idService))
	}}
	ctx, canc := context.WithCancel(context.Background())
	startErr := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go cp2p.Start(ctx, &wg, startErr)
	time.Sleep(1 * time.Second)
	// cancel the context to shutdown the server
	canc()
	wg.Wait()
	assert.Equal(t, 0, len(startErr), "should not error out")
}

func TestCentP2PServer_StartListenError(t *testing.T) {
	// cause an error by using an invalid port
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	n := c.(*configstore.NodeConfig)
	n.P2PPort = 100000000
	_, err = cfg.UpdateConfig(n)
	assert.NoError(t, err)
	cp2p := &peer{config: cfg}
	ctx, _ := context.WithCancel(context.Background())
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go cp2p.Start(ctx, &wg, startErr)
	err = <-startErr
	wg.Wait()
	assert.NotNil(t, err, "Error should be not nil")
	assert.Equal(t, "failed to parse tcp: 100000000 failed to parse port addr: greater than 65536", err.Error())
}

func TestCentP2PServer_makeBasicHostNoExternalIP(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	listenPort := 38202
	cp2p := &peer{config: cfg}
	pu, pr := c.GetSigningKeyPair()
	priv, pub, err := cp2p.createSigningKey(pu, pr)
	h, err := makeBasicHost(priv, pub, "", listenPort)
	assert.Nil(t, err)
	assert.NotNil(t, h)
}

func TestCentP2PServer_makeBasicHostWithExternalIP(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	externalIP := "100.100.100.100"
	listenPort := 38202
	cp2p := &peer{config: cfg}
	pu, pr := c.GetSigningKeyPair()
	priv, pub, err := cp2p.createSigningKey(pu, pr)
	h, err := makeBasicHost(priv, pub, externalIP, listenPort)
	assert.Nil(t, err)
	assert.NotNil(t, h)
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", externalIP, listenPort))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(h.Addrs()))
	assert.Contains(t, h.Addrs(), addr)
}

func TestCentP2PServer_makeBasicHostWithWrongExternalIP(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	externalIP := "100.200.300.400"
	listenPort := 38202
	cp2p := &peer{config: cfg}
	pu, pr := c.GetSigningKeyPair()
	priv, pub, err := cp2p.createSigningKey(pu, pr)
	h, err := makeBasicHost(priv, pub, externalIP, listenPort)
	assert.NotNil(t, err)
	assert.Nil(t, h)
}
