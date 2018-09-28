package p2p

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

func TestCentP2PServer_Start(t *testing.T) {

}

func TestCentP2PServer_StartContextCancel(t *testing.T) {
	priv, pub := getKeys()
	cp2p := NewCentP2PServer(38203, []string{}, pub, priv)
	ctx, canc := context.WithCancel(context.Background())
	startErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go cp2p.Start(ctx, &wg, startErr)
	// TODO make some rpc calls to make sure the peer is up
	time.Sleep(1 * time.Second)
	// cancel the context to shutdown the server
	canc()
	wg.Wait()
	// TODO make some rpc calls to make sure the peer is down
}

func TestCentP2PServer_StartListenError(t *testing.T) {
	// cause an error by using an invalid port
	priv, pub := getKeys()
	cp2p := NewCentP2PServer(100000000, []string{}, pub, priv)
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

func getKeys() (ed25519.PrivateKey, ed25519.PublicKey) {
	return ed25519keys.GetPrivateSigningKey("../../example/resources/signingKey.key.pem"),
		ed25519keys.GetPublicSigningKey("../../example/resources/signingKey.pub.pem")
}
