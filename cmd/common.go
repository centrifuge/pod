package cmd

import (
	"context"

	logging "github.com/ipfs/go-log"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	c "github.com/centrifuge/go-centrifuge/context"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/keytools"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/syndtr/goleveldb/leveldb"
)

var log = logging.Logger("centrifuge-cmd")

func createIdentity(idService identity.Service) (identity.CentID, error) {
	centID := identity.RandomCentID()
	_, confirmations, err := idService.CreateIdentity(centID)
	if err != nil {
		return [identity.CentIDLength]byte{}, err
	}
	_ = <-confirmations

	return centID, nil
}

func generateKeys(config config.Configuration) {
	p2pPub, p2pPvt := config.GetSigningKeyPair()
	ethAuthPub, ethAuthPvt := config.GetEthAuthKeyPair()
	keytools.GenerateSigningKeyPair(p2pPub, p2pPvt, "ed25519")
	keytools.GenerateSigningKeyPair(p2pPub, p2pPvt, "ed25519")
	keytools.GenerateSigningKeyPair(ethAuthPub, ethAuthPvt, "secp256k1")
}

func addKeys(idService identity.Service) error {
	err := idService.AddKeyFromConfig(identity.KeyPurposeP2P)
	if err != nil {
		panic(err)
	}
	err = idService.AddKeyFromConfig(identity.KeyPurposeSigning)
	if err != nil {
		panic(err)
	}
	err = idService.AddKeyFromConfig(identity.KeyPurposeEthMsgAuth)
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
	cfg := ctx[config.BootstrappedConfig].(config.Configuration)
	generateKeys(cfg)

	idService := ctx[identity.BootstrappedIDService].(identity.Service)
	id, err := createIdentity(idService)
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
	err = addKeys(idService)
	if err != nil {
		return err
	}
	canc()
	db := ctx[config.BootstrappedLevelDB].(*leveldb.DB)
	db.Close()
	return nil
}

// RunBootstrap bootstraps the node for running
func RunBootstrap(cfgFile string) {
	mb := c.MainBootstrapper{}
	mb.PopulateRunBootstrappers()
	ctx := map[string]interface{}{}
	ctx[config.BootstrappedConfigFile] = cfgFile
	err := mb.Bootstrap(ctx)
	if err != nil {
		// application must not continue to run
		panic(err)
	}
}

// BaseBootstrap bootstraps the node for testing purposes mainly
func BaseBootstrap(cfgFile string) map[string]interface{} {
	mb := c.MainBootstrapper{}
	mb.PopulateBaseBootstrappers()
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
	ctx := BaseBootstrap(cfgFile)
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	// init node with only the queue server which is needed by commands
	n := node.New([]node.Server{queueSrv})
	cx, canc := context.WithCancel(context.Background())
	e := make(chan error)
	go n.Start(cx, e)
	return ctx, canc, nil
}
