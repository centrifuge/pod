package cmd

import (
	"context"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/identity/did"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"

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
	p2pPub, p2pPvt := config.GetP2PKeyPair()
	signPub, signPvt := config.GetSigningKeyPair()
	ethAuthPub, ethAuthPvt := config.GetEthAuthKeyPair()
	crypto.GenerateSigningKeyPair(p2pPub, p2pPvt, "ed25519")
	crypto.GenerateSigningKeyPair(signPub, signPvt, "ed25519")
	crypto.GenerateSigningKeyPair(ethAuthPub, ethAuthPvt, "secp256k1")
}

func addKeys(config config.Configuration, idService identity.Service) error {
	err := idService.AddKeyFromConfig(config, identity.KeyPurposeP2P)
	if err != nil {
		return err
	}
	err = idService.AddKeyFromConfig(config, identity.KeyPurposeSigning)
	if err != nil {
		return err
	}
	err = idService.AddKeyFromConfig(config, identity.KeyPurposeEthMsgAuth)
	if err != nil {
		return err
	}
	return nil
}

func getKeyPairsFromConfig(config config.Configuration) (map[int]did.Key, error) {

	keys := map[int]did.Key{}
	var pk []byte

	// ed25519 keys
	// KeyPurposeP2P
	pk, _, err := ed25519.GetSigningKeyPair(config.GetP2PKeyPair())
	if err != nil {
		return nil, err
	}
	pk32, err := utils.SliceToByte32(pk)
	if err != nil {
		return nil, err
	}
	keys[identity.KeyPurposeP2P] = did.NewKey(pk32, big.NewInt(identity.KeyPurposeP2P),big.NewInt(did.KeyTypeECDSA))

	// KeyPurposeSigning
	pk, _, err = ed25519.GetSigningKeyPair(config.GetSigningKeyPair())
	if err != nil {
		return nil, err
	}
	keys[identity.KeyPurposeSigning] = did.NewKey(pk32, big.NewInt(identity.KeyPurposeSigning),big.NewInt(did.KeyTypeECDSA))

	// secp256k1 keys
	// KeyPurposeEthMsgAuth
	pk, _, err = secp256k1.GetEthAuthKey(config.GetEthAuthKeyPair())
	if err != nil {
		return nil, err
	}
	pubKey, err := hexutil.Decode(secp256k1.GetAddress(pk))
	if err != nil {
		return nil, err
	}
	pk32, err = utils.SliceToByte32(pubKey)
	if err != nil {
		return nil, err
	}

	keys[identity.KeyPurposeEthMsgAuth] = did.NewKey(pk32, big.NewInt(identity.KeyPurposeEthMsgAuth),big.NewInt(did.KeyTypeECDSA))

	return keys, nil
}


func addKeysFromConfig(config config.Configuration, tctx context.Context,idSrv did.Service) error {
	keys, err := getKeyPairsFromConfig(config)
	if err != nil {
		return err
	}
	err = idSrv.AddKey(tctx, keys[identity.KeyPurposeP2P])
	if err != nil {
		return err
	}

	err = idSrv.AddKey(tctx, keys[identity.KeyPurposeSigning])
	if err != nil {
		return err
	}

	err = idSrv.AddKey(tctx, keys[identity.KeyPurposeEthMsgAuth])
	if err != nil {
		return err
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

	// create keys locally
	generateKeys(cfg)

	tc, err := configstore.TempAccount(cfg.GetEthereumDefaultAccountName(), cfg)
	if err != nil {
		return err
	}

	tctx, err := contextutil.New(context.Background(), tc)
	if err != nil {
		return err
	}

	identityFactory := ctx[did.BootstrappedDIDFactory].(did.Factory)
	idService := ctx[did.BootstrappedDIDService].(did.Service)

	did, err := identityFactory.CreateIdentity(tctx)
	if err != nil {
		return err
	}

	v.Set("identityId", did.ToAddress().String())
	err = v.WriteConfig()
	if err != nil {
		return err
	}
	cfg.Set("identityId", did.ToAddress().String())
	log.Infof("Identity created [%s]", did.ToAddress().String())


	err = addKeysFromConfig(cfg,tctx,idService)
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

// CreateConfig creates a config file using provide parameters and the default config
// Deprecated
func CreateConfigDeprecated(
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

	tctx, err := contextutil.New(context.Background(), tc)
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
