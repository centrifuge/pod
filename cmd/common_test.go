//go:build integration

package cmd

import (
	"fmt"
	"net"
	"os"
	"path"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	jobs.Bootstrapper{},
	&configstore.Bootstrapper{},
	&integration_test.Bootstrapper{},
}

func TestMain(m *testing.M) {
	_ = bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestCreateConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp(path.Join(os.TempDir(), "go-centrifuge"), "create-config-test-*")
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

	// Initialize a config service that's using the storage from the config that
	// we create here.
	configDB, err := leveldb.NewLevelDBStorage(cfg.GetConfigStoragePath())
	assert.NoError(t, err)

	dbRepo := leveldb.NewLevelDBRepository(configDB)

	dbRepo.Register(new(configstore.NodeAdmin))

	cfgService := configstore.NewService(configstore.NewDBRepository(dbRepo))

	cfgAdminPubKey, cfgAdminPrivateKey := cfg.GetNodeAdminKeyPair()

	expectedAdminPubKeyPath := path.Join(tempDir, "node_admin.pub.pem")
	expectedAdminPrivateKeyPath := path.Join(tempDir, "node_admin.key.pem")

	assert.Equal(t, expectedAdminPubKeyPath, cfgAdminPubKey)
	assert.Equal(t, expectedAdminPrivateKeyPath, cfgAdminPrivateKey)

	assertFileExists(t, cfgAdminPubKey)
	assertFileExists(t, cfgAdminPrivateKey)

	pubKey, err := utils.ReadKeyFromPemFile(cfgAdminPubKey, utils.PublicKey)
	assert.NoError(t, err)
	assert.Len(t, pubKey, 32)

	adminAccountID, err := types.NewAccountID(pubKey)
	assert.NoError(t, err)

	nodeAdmin, err := cfgService.GetNodeAdmin()
	assert.NoError(t, err)
	assert.Equal(t, adminAccountID, nodeAdmin.GetAccountID())

	privateKey, err := utils.ReadKeyFromPemFile(cfgAdminPrivateKey, utils.PrivateKey)
	assert.NoError(t, err)
	assert.Len(t, privateKey, 32)

	cfgP2pPubKey, cfgP2pPrivateKey := cfg.GetP2PKeyPair()

	expectedP2pPubKeyPath := path.Join(tempDir, "p2p.pub.pem")
	expectedP2pPrivateKeyPath := path.Join(tempDir, "p2p.key.pem")

	assert.Equal(t, expectedP2pPubKeyPath, cfgP2pPubKey)
	assert.Equal(t, expectedP2pPrivateKeyPath, cfgP2pPrivateKey)

	assertFileExists(t, cfgP2pPubKey)
	assertFileExists(t, cfgP2pPrivateKey)

	pubKey, err = utils.ReadKeyFromPemFile(cfgP2pPubKey, utils.PublicKey)
	assert.NoError(t, err)
	assert.Len(t, pubKey, 32)

	privateKey, err = utils.ReadKeyFromPemFile(cfgP2pPrivateKey, utils.PrivateKey)
	assert.NoError(t, err)
	assert.Len(t, privateKey, 64)

	cfgSigningPubKey, cfgSigningPrivateKey := cfg.GetSigningKeyPair()

	expectedSigningPubKeyPath := path.Join(tempDir, "signing.pub.pem")
	expectedSigningPrivateKeyPath := path.Join(tempDir, "signing.key.pem")

	assert.Equal(t, expectedSigningPubKeyPath, cfgSigningPubKey)
	assert.Equal(t, expectedSigningPrivateKeyPath, cfgSigningPrivateKey)

	assertFileExists(t, cfgSigningPubKey)
	assertFileExists(t, cfgSigningPrivateKey)

	pubKey, err = utils.ReadKeyFromPemFile(cfgSigningPubKey, utils.PublicKey)
	assert.NoError(t, err)
	assert.Len(t, pubKey, 32)

	privateKey, err = utils.ReadKeyFromPemFile(cfgSigningPrivateKey, utils.PrivateKey)
	assert.NoError(t, err)
	assert.Len(t, privateKey, 64)
}

func assertFileExists(t *testing.T, path string) {
	fileInfo, err := os.Stat(path)
	assert.NoError(t, err)
	assert.NotNil(t, fileInfo)
}
