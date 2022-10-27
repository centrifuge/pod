//go:build unit

package p2p

import (
	"context"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	protocolIDDispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/errors"
	ms "github.com/centrifuge/go-centrifuge/p2p/messenger"
	p2pMocks "github.com/centrifuge/go-centrifuge/p2p/mocks"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/libp2p/go-libp2p-core/crypto"
	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

const (
	testStoragePattern = "p2p-server-*"
)

func TestPeer_Server_Start(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx, cancel := context.WithCancel(context.Background())

	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(testStoragePattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	err = os.MkdirAll(randomStoragePath, os.ModePerm)
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PPort").
		Return(9080)

	p2pPublicKeyPath := path.Join(randomStoragePath, "p2p_public.pub.pem")
	p2pPrivateKeyPath := path.Join(randomStoragePath, "p2p_public.key.pem")

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PKeyPair").
		Return(p2pPublicKeyPath, p2pPrivateKeyPath)

	err = config.GenerateAndWriteP2PKeys(genericUtils.GetMock[*config.ConfigurationMock](mocks))
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PExternalIP").
		Return("")

	connectionTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(connectionTimeout)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock1 := config.NewAccountMock(t)
	accountMock1.On("GetIdentity").
		Return(accountID1)

	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock2 := config.NewAccountMock(t)
	accountMock2.On("GetIdentity").
		Return(accountID2)

	accounts := []config.Account{accountMock1, accountMock2}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccounts").
		Return(accounts, nil).
		Once()

	protocolIDChan := make(chan protocol.ID)

	genericUtils.GetMock[*protocolIDDispatcher.DispatcherMock[protocol.ID]](mocks).On("Subscribe", ctx).
		Return(protocolIDChan, nil).
		Once()

	bootstrapPeers := []string{
		"/ip4/127.0.0.1/tcp/38202/ipfs/QmTQxbwkuZYYDfuzTbxEAReTNCLozyy558vQngVvPMjLYk",
		"/ip4/127.0.0.1/tcp/38203/ipfs/QmVf6EN6mkqWejWKW2qPu16XpdG3kJo1T3mhahPB5Se5n1",
	}

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetBootstrapPeers").
		Return(bootstrapPeers).
		Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("IsDebugLogEnabled").
		Return(false).
		Once()

	var wg sync.WaitGroup

	wg.Add(1)

	startupErr := make(chan error, 1)

	go peer.Start(ctx, &wg, startupErr)

	select {
	case <-time.After(3 * time.Second):
	case err := <-startupErr:
		assert.Fail(t, "expected no error, got: %s", err)
	}

	cancel()

	doneChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected peer to be stopped")
	case <-doneChan:
		// Test successful
	}
}

func TestPeer_Server_Start_NoP2PPort(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx, cancel := context.WithCancel(context.Background())

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PPort").
		Return(0)

	var wg sync.WaitGroup

	wg.Add(1)

	startupErr := make(chan error, 1)

	go peer.Start(ctx, &wg, startupErr)

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected error")
	case err := <-startupErr:
		assert.NotNil(t, err)
	}

	cancel()

	doneChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected peer to be stopped")
	case <-doneChan:
		// Test successful
	}
}

func TestPeer_Server_Start_NoP2PKeys(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx, cancel := context.WithCancel(context.Background())

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PPort").
		Return(9080)

	p2pPublicKeyPath := "p2p_public.pub.pem"
	p2pPrivateKeyPath := "p2p_public.key.pem"

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PKeyPair").
		Return(p2pPublicKeyPath, p2pPrivateKeyPath)

	var wg sync.WaitGroup

	wg.Add(1)

	startupErr := make(chan error, 1)

	go peer.Start(ctx, &wg, startupErr)

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected error")
	case err := <-startupErr:
		assert.NotNil(t, err)
	}

	cancel()

	doneChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected peer to be stopped")
	case <-doneChan:
		// Test successful
	}
}

func TestPeer_Server_Start_BasicHostError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx, cancel := context.WithCancel(context.Background())

	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(testStoragePattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	err = os.MkdirAll(randomStoragePath, os.ModePerm)
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PPort").
		Return(9080)

	p2pPublicKeyPath := path.Join(randomStoragePath, "p2p_public.pub.pem")
	p2pPrivateKeyPath := path.Join(randomStoragePath, "p2p_public.key.pem")

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PKeyPair").
		Return(p2pPublicKeyPath, p2pPrivateKeyPath)

	err = config.GenerateAndWriteP2PKeys(genericUtils.GetMock[*config.ConfigurationMock](mocks))
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PExternalIP").
		Return("invalid-ip")

	var wg sync.WaitGroup

	wg.Add(1)

	startupErr := make(chan error, 1)

	go peer.Start(ctx, &wg, startupErr)

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected error")
	case err := <-startupErr:
		assert.NotNil(t, err)
	}

	cancel()

	doneChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected peer to be stopped")
	case <-doneChan:
		// Test successful
	}
}

func TestPeer_Server_Start_InitProtocolsError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx, cancel := context.WithCancel(context.Background())

	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(testStoragePattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	err = os.MkdirAll(randomStoragePath, os.ModePerm)
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PPort").
		Return(9080)

	p2pPublicKeyPath := path.Join(randomStoragePath, "p2p_public.pub.pem")
	p2pPrivateKeyPath := path.Join(randomStoragePath, "p2p_public.key.pem")

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PKeyPair").
		Return(p2pPublicKeyPath, p2pPrivateKeyPath)

	err = config.GenerateAndWriteP2PKeys(genericUtils.GetMock[*config.ConfigurationMock](mocks))
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PExternalIP").
		Return("")

	connectionTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(connectionTimeout)

	cfgServiceErr := errors.New("error")

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccounts").
		Return(nil, cfgServiceErr).
		Once()

	var wg sync.WaitGroup

	wg.Add(1)

	startupErr := make(chan error, 1)

	go peer.Start(ctx, &wg, startupErr)

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected error")
	case err := <-startupErr:
		assert.Equal(t, cfgServiceErr, err)
	}

	cancel()

	doneChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected peer to be stopped")
	case <-doneChan:
		// Test successful
	}
}

func TestPeer_Server_Start_ProtocolIDDispatcherError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx, cancel := context.WithCancel(context.Background())

	randomStoragePath, err := testingcommons.GetRandomTestStoragePath(testStoragePattern)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(randomStoragePath)
	}()

	err = os.MkdirAll(randomStoragePath, os.ModePerm)
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PPort").
		Return(9080)

	p2pPublicKeyPath := path.Join(randomStoragePath, "p2p_public.pub.pem")
	p2pPrivateKeyPath := path.Join(randomStoragePath, "p2p_public.key.pem")

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PKeyPair").
		Return(p2pPublicKeyPath, p2pPrivateKeyPath)

	err = config.GenerateAndWriteP2PKeys(genericUtils.GetMock[*config.ConfigurationMock](mocks))
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PExternalIP").
		Return("")

	connectionTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(connectionTimeout)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock1 := config.NewAccountMock(t)
	accountMock1.On("GetIdentity").
		Return(accountID1)

	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock2 := config.NewAccountMock(t)
	accountMock2.On("GetIdentity").
		Return(accountID2)

	accounts := []config.Account{accountMock1, accountMock2}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccounts").
		Return(accounts, nil).
		Once()

	protocolIDDispatcherErr := errors.New("error")

	genericUtils.GetMock[*protocolIDDispatcher.DispatcherMock[protocol.ID]](mocks).On("Subscribe", ctx).
		Return(nil, protocolIDDispatcherErr).
		Once()

	var wg sync.WaitGroup

	wg.Add(1)

	startupErr := make(chan error, 1)

	go peer.Start(ctx, &wg, startupErr)

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected error")
	case err := <-startupErr:
		assert.Equal(t, protocolIDDispatcherErr, err)
	}

	cancel()

	doneChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected peer to be stopped")
	case <-doneChan:
		// Test successful
	}
}

func TestPeer_Server_processProtocolIDs(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	protocolID1 := protocol.ID("protocol-id-1")
	protocolID2 := protocol.ID("protocol-id-2")

	genericUtils.GetMock[*ms.MessengerMock](mocks).On("Init", protocolID1).
		Once()

	genericUtils.GetMock[*ms.MessengerMock](mocks).On("Init", protocolID2).
		Once()

	c := make(chan protocol.ID)

	ctx, cancel := context.WithCancel(context.Background())

	go peer.processProtocolIDs(ctx, c)

	sendDone := make(chan struct{})

	go func() {
		defer close(sendDone)

		c <- protocolID1
		c <- protocolID2
	}()

	select {
	case <-time.After(3 * time.Second):
		assert.Fail(t, "expected that protocol IDs are sent")
	case <-sendDone:
	}

	cancel()

	time.Sleep(1 * time.Second)

	// Context is canceled, no protocol ID should be processed now.

	select {
	case c <- protocolID1:
		assert.Fail(t, "protocol ID should not be processed")
	case <-time.After(1 * time.Second):
		// Test successful
	}
}

func TestPeer_Server_runDHT(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx := context.Background()

	bootstrapPeers := []string{
		"/ip4/127.0.0.1/tcp/38202/ipfs/QmTQxbwkuZYYDfuzTbxEAReTNCLozyy558vQngVvPMjLYk",
		"/ip4/127.0.0.1/tcp/38203/ipfs/QmVf6EN6mkqWejWKW2qPu16XpdG3kJo1T3mhahPB5Se5n1",
	}

	multiAddress1, err := ma.NewMultiaddr(bootstrapPeers[0])
	assert.NoError(t, err)
	peer1, err := libp2ppeer.AddrInfoFromP2pAddr(multiAddress1)
	assert.NoError(t, err)

	multiAddress2, err := ma.NewMultiaddr(bootstrapPeers[1])
	assert.NoError(t, err)
	peer2, err := libp2ppeer.AddrInfoFromP2pAddr(multiAddress2)
	assert.NoError(t, err)

	peerstoreMock := p2pMocks.NewPeerstoreMock(t)

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).On("Peerstore").
		Return(peerstoreMock, nil).Times(len(bootstrapPeers))

	peerstoreMock.On("AddAddrs", peer1.ID, peer1.Addrs, time.Duration(pstore.PermanentAddrTTL)).
		Once()

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).On("Connect", ctx, *peer1).
		Return(nil).Once()

	peerstoreMock.On("AddAddrs", peer2.ID, peer2.Addrs, time.Duration(pstore.PermanentAddrTTL)).
		Once()

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).On("Connect", ctx, *peer2).
		Return(nil).Once()

	genericUtils.GetMock[*p2pMocks.IpfsDHTMock](mocks).On("Bootstrap", ctx).
		Return(nil).Once()

	err = peer.runDHT(ctx, bootstrapPeers)
	assert.NoError(t, err)
}

func TestPeer_Server_runDHT_Errors(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx := context.Background()

	bootstrapPeers := []string{
		"invalid-peer",
		"/ip4/127.0.0.1/tcp/38203/ipfs/QmVf6EN6mkqWejWKW2qPu16XpdG3kJo1T3mhahPB5Se5n1",
	}

	multiAddress, err := ma.NewMultiaddr(bootstrapPeers[1])
	assert.NoError(t, err)
	addrInfo, err := libp2ppeer.AddrInfoFromP2pAddr(multiAddress)
	assert.NoError(t, err)

	peerstoreMock := p2pMocks.NewPeerstoreMock(t)

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).On("Peerstore").
		Return(peerstoreMock, nil).Once()

	peerstoreMock.On("AddAddrs", addrInfo.ID, addrInfo.Addrs, time.Duration(pstore.PermanentAddrTTL)).
		Once()

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).On("Connect", ctx, *addrInfo).
		Return(nil).Once()

	genericUtils.GetMock[*p2pMocks.IpfsDHTMock](mocks).On("Bootstrap", ctx).
		Return(nil).Once()

	err = peer.runDHT(ctx, bootstrapPeers)
	assert.NoError(t, err)

	// No peers

	genericUtils.GetMock[*p2pMocks.IpfsDHTMock](mocks).On("Bootstrap", ctx).
		Return(nil).Once()

	err = peer.runDHT(ctx, nil)
	assert.NoError(t, err)

	// Bootstrap error

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).On("Peerstore").
		Return(peerstoreMock, nil).Once()

	peerstoreMock.On("AddAddrs", addrInfo.ID, addrInfo.Addrs, time.Duration(pstore.PermanentAddrTTL)).
		Once()

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).On("Connect", ctx, *addrInfo).
		Return(nil).Once()

	bootstrapErr := errors.New("error")

	genericUtils.GetMock[*p2pMocks.IpfsDHTMock](mocks).On("Bootstrap", ctx).
		Return(bootstrapErr).Once()

	err = peer.runDHT(ctx, bootstrapPeers)
	assert.ErrorIs(t, err, bootstrapErr)
}

func Test_makeBasicHost(t *testing.T) {
	ctx := context.Background()

	_, privateKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	priv, err := crypto.UnmarshalEd25519PrivateKey(privateKey)
	assert.NoError(t, err)

	externalIP := "127.0.0.1"
	externalPort := 9080

	host, dht, err := makeBasicHost(ctx, priv, externalIP, externalPort)
	assert.NoError(t, err)
	assert.NotNil(t, host)
	assert.NotNil(t, dht)
}

func Test_makeBasicHost_NoExternalIP(t *testing.T) {
	ctx := context.Background()

	_, privateKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	priv, err := crypto.UnmarshalEd25519PrivateKey(privateKey)
	assert.NoError(t, err)

	externalIP := ""
	externalPort := 9080

	host, dht, err := makeBasicHost(ctx, priv, externalIP, externalPort)
	assert.NoError(t, err)
	assert.NotNil(t, host)
	assert.NotNil(t, dht)
}

func Test_makeBasicHost_InvalidIP(t *testing.T) {
	ctx := context.Background()

	_, privateKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	priv, err := crypto.UnmarshalEd25519PrivateKey(privateKey)
	assert.NoError(t, err)

	externalIP := "invalid-ip"
	externalPort := 9080

	host, dht, err := makeBasicHost(ctx, priv, externalIP, externalPort)
	assert.NotNil(t, err)
	assert.Nil(t, host)
	assert.Nil(t, dht)
}

func Test_makeBasicHost_InvalidPort(t *testing.T) {
	ctx := context.Background()

	_, privateKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	priv, err := crypto.UnmarshalEd25519PrivateKey(privateKey)
	assert.NoError(t, err)

	externalIP := "127.0.0.1"
	externalPort := 99999

	host, dht, err := makeBasicHost(ctx, priv, externalIP, externalPort)
	assert.NotNil(t, err)
	assert.Nil(t, host)
	assert.Nil(t, dht)
}
