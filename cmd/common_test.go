//go:build integration

package cmd

import (
	"fmt"
	"net"
	"os"
	"path"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/config/configstore"
	"github.com/centrifuge/pod/jobs"
	"github.com/centrifuge/pod/storage/leveldb"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&integration_test.Bootstrapper{},
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&configstore.Bootstrapper{},
	&jobs.Bootstrapper{},
}

func TestMain(m *testing.M) {
	_ = bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestCreateConfig(t *testing.T) {
	tempDir, err := testingcommons.GetRandomTestStoragePath("config-create-test")
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	network := "catalyst"
	apiHost := "127.0.0.1"
	apiPort := 8082
	p2pPort := 38202
	bootstrapPeers := []string{
		"/ip4/127.0.0.1/tcp/38202/ipfs/QmTQxbwkuZYYDfuzTbxEAReTNCLozyy558vQngVvPMjLYk",
		"/ip4/127.0.0.1/tcp/38203/ipfs/QmVf6EN6mkqWejWKW2qPu16XpdG3kJo1T3mhahPB5Se5n1",
	}
	centChainURL := "ws://127.0.0.1:9946"
	authenticationEnabled := true
	ipfsPinningServiceName := "pinata"
	ipfsPinningServiceURL := "https://pinata.com"
	ipfsPinnginServiceAuth := "test-auth"
	// Ferdie's secret seed
	podOperatorSecretSeed := "0x42438b7883391c05512a938e36c2df0131e088b3756d6aa7a755fbff19d2f842"
	// Eve's secret seed
	podAdminSecretSeed := "0x786ad0e2df456fe43dd1f91ebca22e235bc162e0bb8d53c633e8c85b2af68b7a"

	err = CreateConfig(
		tempDir,
		network,
		apiHost,
		apiPort,
		p2pPort,
		bootstrapPeers,
		"",
		centChainURL,
		authenticationEnabled,
		ipfsPinningServiceName,
		ipfsPinningServiceURL,
		ipfsPinnginServiceAuth,
		podOperatorSecretSeed,
		podAdminSecretSeed,
	)
	assert.NoError(t, err)

	cfg := config.LoadConfiguration(path.Join(tempDir, "config.yaml"))
	assert.Equal(t, authenticationEnabled, cfg.IsAuthenticationEnabled())
	assert.Equal(t, centChainURL, cfg.GetCentChainNodeURL())
	assert.Equal(t, network, cfg.GetNetworkString())

	configStoragePath := path.Join(tempDir, "db/centrifuge_config_data.leveldb")
	assert.Equal(t, configStoragePath, cfg.GetConfigStoragePath())
	assertFileExists(t, configStoragePath)

	storagePath := path.Join(tempDir, "db/centrifuge_data.leveldb")
	assert.Equal(t, storagePath, cfg.GetStoragePath())
	assertFileExists(t, storagePath)

	assert.Equal(t, bootstrapPeers, cfg.GetBootstrapPeers())
	assert.Equal(t, net.JoinHostPort(apiHost, fmt.Sprintf("%d", apiPort)), cfg.GetServerAddress())
	assert.Equal(t, p2pPort, cfg.GetP2PPort())

	assert.Equal(t, podOperatorSecretSeed, cfg.GetPodOperatorSecretSeed())

	// Initialize a config service that's using the storage from the config that
	// we create here.
	configDB, err := leveldb.NewLevelDBStorage(cfg.GetConfigStoragePath())
	assert.NoError(t, err)

	dbRepo := leveldb.NewLevelDBRepository(configDB)

	dbRepo.Register(new(configstore.PodAdmin))
	dbRepo.Register(new(configstore.PodOperator))

	cfgService := configstore.NewService(configstore.NewDBRepository(dbRepo))

	podAdminKeyPair, err := subkey.DeriveKeyPair(sr25519.Scheme{}, cfg.GetPodAdminSecretSeed())
	assert.NoError(t, err)

	podAdminAccountID, err := types.NewAccountID(podAdminKeyPair.AccountID())
	assert.NoError(t, err)

	podAdmin, err := cfgService.GetPodAdmin()
	assert.NoError(t, err)
	assert.Equal(t, podAdminAccountID, podAdmin.GetAccountID())

	podOperatorKeyPair, err := subkey.DeriveKeyPair(sr25519.Scheme{}, cfg.GetPodOperatorSecretSeed())
	assert.NoError(t, err)

	podOperatorAccountID, err := types.NewAccountID(podOperatorKeyPair.AccountID())
	assert.NoError(t, err)

	podOperator, err := cfgService.GetPodOperator()
	assert.NoError(t, err)
	assert.Equal(t, podOperatorAccountID, podOperator.GetAccountID())

	cfgP2pPubKey, cfgP2pPrivateKey := cfg.GetP2PKeyPair()

	expectedP2pPubKeyPath := path.Join(tempDir, "p2p.pub.pem")
	expectedP2pPrivateKeyPath := path.Join(tempDir, "p2p.key.pem")

	assert.Equal(t, expectedP2pPubKeyPath, cfgP2pPubKey)
	assert.Equal(t, expectedP2pPrivateKeyPath, cfgP2pPrivateKey)

	assertFileExists(t, cfgP2pPubKey)
	assertFileExists(t, cfgP2pPrivateKey)

	pubKey, err := utils.ReadKeyFromPemFile(cfgP2pPubKey, utils.PublicKey)
	assert.NoError(t, err)
	assert.Len(t, pubKey, 32)

	privateKey, err := utils.ReadKeyFromPemFile(cfgP2pPrivateKey, utils.PrivateKey)
	assert.NoError(t, err)
	assert.Len(t, privateKey, 64)
}

func assertFileExists(t *testing.T, path string) {
	fileInfo, err := os.Stat(path)
	assert.NoError(t, err)
	assert.NotNil(t, fileInfo)
}
