package cmd

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/vedhavyas/go-subkey/sr25519"
)

var log = logging.Logger("centrifuge-cmd")

// CreateConfig creates a config file using provide parameters and the default config
func CreateConfig(
	targetDataDir, network, apiHost string,
	apiPort, p2pPort int,
	bootstraps []string,
	p2pConnectionTimeout string,
	centChainURL string,
	authenticationEnabled bool,
	ipfsPinningServiceName, ipfsPinningServiceURL, ipfsPinningServiceAuth string,
	podOperatorSecretSeed string,
) error {
	if podOperatorSecretSeed == "" {
		seed, err := generatePodOperatorSeed()

		if err != nil {
			return err
		}

		podOperatorSecretSeed = seed
	}

	data := map[string]interface{}{
		"targetDataDir":          targetDataDir,
		"network":                network,
		"bootstraps":             bootstraps,
		"apiHost":                apiHost,
		"apiPort":                apiPort,
		"p2pPort":                p2pPort,
		"p2pConnectTimeout":      p2pConnectionTimeout,
		"centChainURL":           centChainURL,
		"authenticationEnabled":  authenticationEnabled,
		"ipfsPinningServiceName": ipfsPinningServiceName,
		"ipfsPinningServiceURL":  ipfsPinningServiceURL,
		"ipfsPinningServiceAuth": ipfsPinningServiceAuth,
		"podOperatorSecretSeed":  podOperatorSecretSeed,
	}

	configFile, err := config.CreateConfigFile(data)

	if err != nil {
		return err
	}

	log.Infof("Config File Created: %s\n", configFile.ConfigFileUsed())

	cfg := config.LoadConfiguration(configFile.ConfigFileUsed())

	err = GenerateNodeKeys(cfg)

	if err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}

	ctx, canc, err := CommandBootstrap(configFile.ConfigFileUsed())

	if err != nil {
		return fmt.Errorf("failed to create bootstraps: %w", err)
	}

	defer canc()
	err = configFile.WriteConfig()
	if err != nil {
		return fmt.Errorf("couldn't write config: %w", err)
	}

	db := ctx[storage.BootstrappedDB].(storage.Repository)
	dbCfg := ctx[storage.BootstrappedConfigDB].(storage.Repository)
	defer db.Close()
	defer dbCfg.Close()
	log.Infof("---------Centrifuge node configuration file successfully created!---------")
	log.Infof("Please run the Centrifuge node using the following command: centrifuge run -c %s\n", configFile.ConfigFileUsed())
	return nil
}

// RunBootstrap bootstraps the node for running
func RunBootstrap(cfgFile string) {
	mb := bootstrappers.MainBootstrapper{}
	mb.PopulateRunBootstrappers()
	ctx := map[string]interface{}{}
	ctx[config.BootstrappedConfigFile] = cfgFile
	err := mb.Bootstrap(ctx)
	if err != nil {
		// application must not continue to run
		panic(err)
	}
}

// ExecCmdBootstrap bootstraps the node for command line and testing purposes
func ExecCmdBootstrap(cfgFile string) map[string]interface{} {
	mb := bootstrappers.MainBootstrapper{}
	mb.PopulateCommandBootstrappers()
	ctx := map[string]interface{}{}
	ctx[config.BootstrappedConfigFile] = cfgFile
	err := mb.Bootstrap(ctx)
	if err != nil {
		// application must not continue to run
		panic(err)
	}
	return ctx
}

// CommandBootstrap bootstraps the node for one time commands
func CommandBootstrap(cfgFile string) (map[string]interface{}, context.CancelFunc, error) {
	ctx := ExecCmdBootstrap(cfgFile)
	dispatcher := ctx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)
	n := node.New([]node.Server{dispatcher})
	cx, canc := context.WithCancel(context.Background())
	e := make(chan error)
	go n.Start(cx, e)
	return ctx, canc, nil
}

// GenerateNodeKeys generates the key pairs used for p2p, document signing and node admin.
func GenerateNodeKeys(config config.Configuration) error {
	p2pPub, p2pPvt := config.GetP2PKeyPair()
	nodeAdminPub, nodeAdminPvt := config.GetNodeAdminKeyPair()

	if err := crypto.GenerateSigningKeyPair(p2pPub, p2pPvt, crypto.CurveEd25519); err != nil {
		return err
	}

	return crypto.GenerateSigningKeyPair(nodeAdminPub, nodeAdminPvt, crypto.CurveSr25519)
}

func generatePodOperatorSeed() (string, error) {
	var scheme sr25519.Scheme

	kp, err := scheme.Generate()

	if err != nil {
		return "", fmt.Errorf("couldn't generate seed: %w", err)
	}

	return hexutil.Encode(kp.Seed()), nil
}
