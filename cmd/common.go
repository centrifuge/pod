package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/gocelery/v2"
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
func CreateConfig(cfgVals *config.ConfigVals) error {
	configFile, err := config.CreateConfigFile(cfgVals)
	if err != nil {
		return err
	}
	log.Infof("Config File Created: %s\n", configFile.ConfigFileUsed())
	ctx, canc, err := CommandBootstrap(configFile.ConfigFileUsed())
	if err != nil {
		return fmt.Errorf("failed to create bootstraps: %w", err)
	}
	defer canc()

	cfg := ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	idFactory := ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	dispatcher := ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	// create keys locally
	err = generateKeys(cfg)
	if err != nil {
		return errors.New("failed to generate keys: %v", err)
	}

	acc, err := configstore.TempAccount("main", cfg)
	if err != nil {
		return err
	}

	did, err := idFactory.NextIdentityAddress()
	if err != nil {
		return fmt.Errorf("failed to fetch the next did from factory: %w", err)
	}
	acc.(*configstore.Account).IdentityID = did[:]
	keys, err := acc.GetKeys()
	if err != nil {
		return fmt.Errorf("failed to fetch keys from the account: %w", err)
	}

	idKeys, err := identity.ConvertAccountKeysToKeyDID(keys)
	if err != nil {
		return fmt.Errorf("failed to convert keys: %w", err)
	}

	tx, err := idFactory.CreateIdentity("main", idKeys)
	if err != nil {
		return fmt.Errorf("failed to send ethereum txn: %w", err)
	}

	ok := dispatcher.RegisterRunnerFunc("ethWaitTxn", func([]interface{}, map[string]interface{}) (interface{}, error) {
		return ethereum.IsTxnSuccessful(context.Background(), client.GetEthClient(), tx.Hash())
	})
	if !ok {
		return errors.New("failed to register worker")
	}

	job := gocelery.NewRunnerFuncJob("Wait for Identity creation", "ethWaitTxn", nil, nil, time.Time{})
	res, err := dispatcher.Dispatch(did, job)
	if err != nil {
		return fmt.Errorf("failed to dispatch identity create job: %w", err)
	}

	_, err = res.Await(context.Background())
	if err != nil {
		return fmt.Errorf("identity creation failed: %w", err)
	}

	configFile.Set("identityId", did.String())
	err = configFile.WriteConfig()
	if err != nil {
		return err
	}

	db := ctx[storage.BootstrappedDB].(storage.Repository)
	dbCfg := ctx[storage.BootstrappedConfigDB].(storage.Repository)
	defer db.Close()
	defer dbCfg.Close()
	log.Infof("---------Centrifuge node configuration file successfully created!---------")
	log.Infof("Please run the Centrifuge node using the following command: centrifuge run -c %s\n", configFile.ConfigFileUsed())
	log.Infof("Your DID is: [%s]", did.String())
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
	dispatcher := ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	n := node.New([]node.Server{dispatcher})
	cx, canc := context.WithCancel(context.Background())
	e := make(chan error)
	go n.Start(cx, e)
	return ctx, canc, nil
}
