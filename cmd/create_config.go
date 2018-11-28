package cmd

import (
	"os"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/keytools"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var targetDataDir string
var ethNodeURL string
var accountKeyPath string
var accountPassword string
var network string
var apiPort int64
var p2pPort int64
var bootstraps []string
var txPoolAccess bool

func createIdentity(idService identity.Service) (identity.CentID, error) {
	centID := identity.RandomCentID()
	_, confirmations, err := idService.CreateIdentity(centID)
	if err != nil {
		return [identity.CentIDLength]byte{}, err
	}
	_ = <-confirmations

	return centID, nil
}

func generateKeys(config config.Config) {
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

func init() {
	home, err := homedir.Dir()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	var createConfigCmd = &cobra.Command{
		Use:   "createconfig",
		Short: "Configures Node",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {

			data := map[string]interface{}{
				"targetDataDir":   targetDataDir,
				"accountKeyPath":  accountKeyPath,
				"accountPassword": accountPassword,
				"network":         network,
				"ethNodeURL":      ethNodeURL,
				"bootstraps":      bootstraps,
				"apiPort":         apiPort,
				"p2pPort":         p2pPort,
				"txpoolaccess":    txPoolAccess,
			}

			v, err := config.CreateConfigFile(data)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			log.Infof("Config File Created: %s\n", v.ConfigFileUsed())

			ctx, canc, _ := commandBootstrap(v.ConfigFileUsed())
			cfg := ctx[config.BootstrappedConfig].(*config.Configuration)
			generateKeys(cfg)

			idService := ctx[identity.BootstrappedIDService].(identity.Service)
			id, err := createIdentity(idService)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			v.Set("identityId", id.String())
			err = v.WriteConfig()
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			cfg.Set("identityId", id.String())

			log.Infof("Identity created [%s] [%x]", id.String(), id)

			err = addKeys(idService)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			canc()
		},
	}

	createConfigCmd.Flags().StringVarP(&targetDataDir, "targetdir", "t", home+"/datadir", "Target Data Dir")
	createConfigCmd.Flags().StringVarP(&ethNodeURL, "ethnodeurl", "e", "ws://127.0.0.1:9546", "URL of Ethereum Client Node")
	createConfigCmd.Flags().StringVarP(&accountKeyPath, "accountkeypath", "z", home+"/datadir/main.key", "Path of Ethereum Account Key JSON file")
	createConfigCmd.Flags().StringVarP(&accountPassword, "accountpwd", "k", "", "Ethereum Account Password")
	createConfigCmd.Flags().Int64VarP(&apiPort, "apiPort", "a", 8082, "Api Port")
	createConfigCmd.Flags().Int64VarP(&p2pPort, "p2pPort", "p", 38202, "Peer-to-Peer Port")
	createConfigCmd.Flags().StringVarP(&network, "network", "n", "russianhill", "Default Network")
	createConfigCmd.Flags().StringSliceVarP(&bootstraps, "bootstraps", "b", nil, "Bootstrap P2P Nodes")
	createConfigCmd.Flags().BoolVarP(&txPoolAccess, "txpoolaccess", "x", true, "Transaction Pool access")
	rootCmd.AddCommand(createConfigCmd)
}
