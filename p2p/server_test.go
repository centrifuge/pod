// +build unit

package p2p

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

func TestCentP2PServer_Start(t *testing.T) {

}

func TestCentP2PServer_StartContextCancel(t *testing.T) {
	cfg.Set("p2p.port", 38203)
	cp2p := &p2pServer{config: cfg, handler: GRPCHandler(cfg, nil)}
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
