//go:build unit
// +build unit

package p2p

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

var (
	cfg       config.Service
	idService identity.Service
)

func TestMain(m *testing.M) {
	ctx := make(map[string]interface{})
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	centChainClient := &centchain.MockAPI{}
	ctx[centchain.BootstrappedCentChainClient] = centChainClient
	ibootstrappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobs.Bootstrapper{},
		&configstore.Bootstrapper{},
		&anchors.Bootstrapper{},
		documents.Bootstrapper{},
	}
	idService = &testingcommons.MockIdentityService{}
	ctx[identity.BootstrappedDIDService] = idService
	ctx[identity.BootstrappedDIDFactory] = &identity.MockFactory{}
	bootstrap.RunTestBootstrappers(ibootstrappers, ctx)
	cfg = ctx[config.BootstrappedConfigStorage].(config.Service)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestCentP2PServer_StartContextCancel(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	c = updateKeys(c)
	n := c.(*configstore.NodeConfig)
	n.P2PPort = 38203
	cfgMock := mockmockConfigStore(n)
	assert.NoError(t, err)
	cp2p := &peer{config: cfgMock, handlerCreator: func() *receiver.Handler {
		return receiver.New(cfgMock, receiver.HandshakeValidator(n.NetworkID, idService), nil, new(testingdocuments.MockRegistry), idService)
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
	c = updateKeys(c)
	n := c.(*configstore.NodeConfig)
	n.P2PPort = 100000000
	cfgMock := mockmockConfigStore(n)
	assert.NoError(t, err)
	cp2p := &peer{config: cfgMock}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go cp2p.Start(ctx, &wg, startErr)
	err = <-startErr
	wg.Wait()
	assert.NotNil(t, err, "Error should be not nil")
	assert.Equal(t, "failed to parse multiaddr \"/ip4/0.0.0.0/tcp/100000000\": invalid value \"100000000\" for protocol tcp: failed to parse port addr: greater than 65536", err.Error())
}

func TestCentP2PServer_makeBasicHostNoExternalIP(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	c = updateKeys(c)
	listenPort := 38202
	pu, pr := c.GetP2PKeyPair()
	priv, pub, err := crypto.ObtainP2PKeypair(pu, pr)
	assert.NoError(t, err)
	h, _, err := makeBasicHost(context.Background(), priv, pub, "", listenPort)
	assert.Nil(t, err)
	assert.NotNil(t, h)
}

func TestCentP2PServer_makeBasicHostWithExternalIP(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	c = updateKeys(c)
	externalIP := "100.100.100.100"
	listenPort := 38202
	pu, pr := c.GetP2PKeyPair()
	priv, pub, err := crypto.ObtainP2PKeypair(pu, pr)
	assert.NoError(t, err)
	h, _, err := makeBasicHost(context.Background(), priv, pub, externalIP, listenPort)
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
	c = updateKeys(c)
	externalIP := "100.200.300.400"
	listenPort := 38202
	pu, pr := c.GetP2PKeyPair()
	priv, pub, err := crypto.ObtainP2PKeypair(pu, pr)
	assert.NoError(t, err)
	h, _, err := makeBasicHost(context.Background(), priv, pub, externalIP, listenPort)
	assert.NotNil(t, err)
	assert.Nil(t, h)
}

func updateKeys(c config.Configuration) config.Configuration {
	n := c.(*configstore.NodeConfig)
	n.MainIdentity.P2PKeyPair.Pub = "../build/resources/p2pKey.pub.pem"
	n.MainIdentity.P2PKeyPair.Pvt = "../build/resources/p2pKey.key.pem"
	n.MainIdentity.SigningKeyPair.Pub = "../build/resources/signingKey.pub.pem"
	n.MainIdentity.SigningKeyPair.Pvt = "../build/resources/signingKey.key.pem"
	return c
}

func mockmockConfigStore(n config.Configuration) *configstore.MockService {
	mockConfigStore := &configstore.MockService{}
	mockConfigStore.On("GetConfig").Return(n, nil)
	mockConfigStore.On("GetAccounts").Return([]config.Account{&configstore.Account{IdentityID: utils.RandomSlice(identity.DIDLength)}}, nil)
	return mockConfigStore
}
