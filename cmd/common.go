package cmd

import (
	"context"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("centrifuge-cmd")

func generateKeys(config config.Configuration) error {
	p2pPub, p2pPvt := config.GetP2PKeyPair()
	signPub, signPvt := config.GetSigningKeyPair()
	err := crypto.GenerateSigningKeyPair(p2pPub, p2pPvt, crypto.CurveEd25519)
	if err != nil {
		return err
	}

	return crypto.GenerateSigningKeyPair(signPub, signPvt, crypto.CurveSecp256K1)
}

// CreateConfig creates a config file using provide parameters and the default config
func CreateConfig(targetDataDir, ethNodeURL, accountKeyPath, accountPassword, network, apiHost string, apiPort, p2pPort int64, bootstraps []string, txPoolAccess, preCommitEnabled bool, p2pConnectionTimeout string, smartContractAddrs *config.SmartContractAddresses, webhookURL string) error {
	data := map[string]interface{}{
		"targetDataDir":     targetDataDir,
		"accountKeyPath":    accountKeyPath,
		"accountPassword":   accountPassword,
		"network":           network,
		"ethNodeURL":        ethNodeURL,
		"bootstraps":        bootstraps,
		"apiHost":           apiHost,
		"apiPort":           apiPort,
		"p2pPort":           p2pPort,
		"p2pConnectTimeout": p2pConnectionTimeout,
		"txpoolaccess":      txPoolAccess,
		"preCommitEnabled":  preCommitEnabled,
		"webhookURL":        webhookURL,
	}
	if smartContractAddrs != nil {
		data["smartContractAddresses"] = smartContractAddrs
	}

	configFile, err := config.CreateConfigFile(data)
	if err != nil {
		return err
	}
	log.Infof("Config File Created: %s\n", configFile.ConfigFileUsed())
	ctx, canc, _ := CommandBootstrap(configFile.ConfigFileUsed())
	cfg := ctx[bootstrap.BootstrappedConfig].(config.Configuration)

	idService, ok := ctx[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("bootstrapped identity service not initialized")
	}
	idFactory, ok := ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	if !ok {
		return errors.New("bootstrapped identity factory not initialized")
	}

	// create keys locally
	err = generateKeys(cfg)
	if err != nil {
		return errors.New("failed to generate keys: %v", err)
	}

	acc, err := configstore.TempAccount("main", cfg)
	if err != nil {
		return err
	}
	ctxh, err := contextutil.New(context.Background(), acc)
	if err != nil {
		return err
	}
	DID, err := idFactory.CreateIdentity(ctxh)
	if err != nil {
		return err
	}

	acci := acc.(*configstore.Account)
	acci.IdentityID = DID[:]

	configFile.Set("identityId", DID.String())
	err = configFile.WriteConfig()
	if err != nil {
		return err
	}
	cfg.Set("identityId", DID.String())
	log.Infof("Identity created [%s]", DID.String())

	err = idService.AddKeysForAccount(acci)
	if err != nil {
		return err
	}

	canc()
	db := ctx[storage.BootstrappedDB].(storage.Repository)
	dbCfg := ctx[storage.BootstrappedConfigDB].(storage.Repository)
	defer db.Close()
	defer dbCfg.Close()
	log.Infof("---------Centrifuge node configuration file successfully created!---------")
	log.Infof("Please run the Centrifuge node using the following command: centrifuge run -c %s\n", configFile.ConfigFileUsed())
	log.Infof("Your DID is: [%s]", DID.String())
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
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	// init node with only the queue server which is needed by commands
	n := node.New([]node.Server{queueSrv})
	cx, canc := context.WithCancel(context.Background())
	e := make(chan error)
	go n.Start(cx, e)
	return ctx, canc, nil
}
