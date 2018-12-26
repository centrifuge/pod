// +build unit

package p2p

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/p2p/grpc"
	"github.com/paralin/go-libp2p-grpc"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/storage"

	"github.com/centrifuge/go-centrifuge/config"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

var (
	cfg        config.Configuration
	testClient *p2pServer
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		documents.Bootstrapper{},
	}
	ctx := make(map[string]interface{})
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	testClient = &p2pServer{config: cfg}
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestCentP2PServer_Start(t *testing.T) {

}

func TestCentP2PServer_StartContextCancel(t *testing.T) {
	cfg.Set("p2p.port", 38203)
	cp2p := &p2pServer{grpcSrvs: make(map[identity.CentID]*p2pgrpc.GRPCProtocol), config: cfg, grpcHandlerCreator: func() p2ppb.P2PServiceServer {
		return grpc.New(cfg, nil)
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
	cfg.Set("p2p.port", 100000000)
	cp2p := &p2pServer{config: cfg}
	ctx, _ := context.WithCancel(context.Background())
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go cp2p.Start(ctx, &wg, startErr)
	err := <-startErr
	wg.Wait()
	assert.NotNil(t, err, "Error should be not nil")
	assert.Equal(t, "failed to parse tcp: 100000000 failed to parse port addr: greater than 65536", err.Error())
}

func TestCentP2PServer_makeBasicHostNoExternalIP(t *testing.T) {
	listenPort := 38202
	cfg.Set("p2p.port", listenPort)
	cp2p := &p2pServer{config: cfg}

	h, err := cp2p.makeBasicHost(listenPort)
	assert.Nil(t, err)
	assert.NotNil(t, h)
}

func TestCentP2PServer_makeBasicHostWithExternalIP(t *testing.T) {
	externalIP := "100.100.100.100"
	listenPort := 38202
	cfg.Set("p2p.port", listenPort)
	cfg.Set("p2p.externalIP", externalIP)
	cp2p := &p2pServer{config: cfg}
	h, err := cp2p.makeBasicHost(listenPort)
	assert.Nil(t, err)
	assert.NotNil(t, h)
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", externalIP, listenPort))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(h.Addrs()))
	assert.Contains(t, h.Addrs(), addr)
}

func TestCentP2PServer_makeBasicHostWithWrongExternalIP(t *testing.T) {
	externalIP := "100.200.300.400"
	listenPort := 38202
	cfg.Set("p2p.port", listenPort)
	cfg.Set("p2p.externalIP", externalIP)
	cp2p := &p2pServer{config: cfg}
	h, err := cp2p.makeBasicHost(listenPort)
	assert.NotNil(t, err)
	assert.Nil(t, h)
}
