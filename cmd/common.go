package cmd

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers"

	"github.com/centrifuge/go-centrifuge/storage"

	logging "github.com/ipfs/go-log"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/queue"
)

var log = logging.Logger("centrifuge-cmd")

func createIdentity(ctx context.Context, idService identity.Service) (identity.CentID, error) {
	centID := identity.RandomCentID()
	_, confirmations, err := idService.CreateIdentity(ctx, centID)
	if err != nil {
		return [identity.CentIDLength]byte{}, err
	}
	_ = <-confirmations

	return centID, nil
}

func generateKeys(config config.Configuration) {
	p2pPub, p2pPvt := config.GetSigningKeyPair()
	ethAuthPub, ethAuthPvt := config.GetEthAuthKeyPair()
	crypto.GenerateSigningKeyPair(p2pPub, p2pPvt, "ed25519")
	crypto.GenerateSigningKeyPair(ethAuthPub, ethAuthPvt, "secp256k1")
}

func addKeys(config config.Configuration, idService identity.Service) error {
	err := idService.AddKeyFromConfig(config, identity.KeyPurposeP2P)
	if err != nil {
		panic(err)
	}
	err = idService.AddKeyFromConfig(config, identity.KeyPurposeSigning)
	if err != nil {
		panic(err)
	}
	err = idService.AddKeyFromConfig(config, identity.KeyPurposeEthMsgAuth)
	if err != nil {
		panic(err)
	}
	return nil
}

// CreateConfig creates a config file using provide parameters and the default config
func CreateConfig(
	targetDataDir, ethNodeURL, accountKeyPath, accountPassword, network string,
	apiPort, p2pPort int64,
	bootstraps []string,
	txPoolAccess bool,
	p2pConnectionTimeout string,
	smartContractAddrs *config.SmartContractAddresses) error {

	data := map[string]interface{}{
		"targetDataDir":     targetDataDir,
		"accountKeyPath":    accountKeyPath,
		"accountPassword":   accountPassword,
		"network":           network,
		"ethNodeURL":        ethNodeURL,
		"bootstraps":        bootstraps,
		"apiPort":           apiPort,
		"p2pPort":           p2pPort,
		"p2pConnectTimeout": p2pConnectionTimeout,
		"txpoolaccess":      txPoolAccess,
	}
	if smartContractAddrs != nil {
		data["smartContractAddresses"] = smartContractAddrs
	}
	v, err := config.CreateConfigFile(data)
	if err != nil {
		return err
	}
	log.Infof("Config File Created: %s\n", v.ConfigFileUsed())
	ctx, canc, _ := CommandBootstrap(v.ConfigFileUsed())
	cfg := ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	generateKeys(cfg)

	tc, err := configstore.TempAccount(cfg.GetEthereumDefaultAccountName(), cfg)
	if err != nil {
		return err
	}

	tctx, err := contextutil.NewCentrifugeContext(context.Background(), tc)
	if err != nil {
		return err
	}

	idService := ctx[identity.BootstrappedIDService].(identity.Service)
	id, err := createIdentity(tctx, idService)
	if err != nil {
		return err
	}
	v.Set("identityId", id.String())
	err = v.WriteConfig()
	if err != nil {
		return err
	}
	cfg.Set("identityId", id.String())
	log.Infof("Identity created [%s] [%x]", id.String(), id)
	err = addKeys(cfg, idService)
	if err != nil {
		return err
	}
	canc()
	db := ctx[storage.BootstrappedDB].(storage.Repository)
	dbCfg := ctx[storage.BootstrappedConfigDB].(storage.Repository)
	db.Close()
	dbCfg.Close()
	log.Infof("---------Centrifuge node configuration file successfully created!---------")
	log.Infof("Please run the Centrifuge node using the following command: centrifuge run -c %s\n", v.ConfigFileUsed())
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
