package cmd

import (
	"os"

	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var targetDataDir string
var ethNodeUrl string
var accountKeyPath string
var accountPassword string
var network string
var apiPort int64
var p2pPort int64
var bootstraps []string

func createIdentity() (identity.CentID, error) {
	centrifugeId := identity.NewRandomCentID()
	identityService := identity.EthereumIdentityService{}
	_, confirmations, err := identityService.CreateIdentity(centrifugeId)
	if err != nil {
		return [identity.CentIDByteLength]byte{}, err
	}
	_ = <-confirmations

	return centrifugeId, nil
}

func generateKeys() {
	p2pPub, p2pPvt := config.Config.GetSigningKeyPair()
	ethAuthPub, ethAuthPvt := config.Config.GetEthAuthKeyPair()
	keytools.GenerateSigningKeyPair(p2pPub, p2pPvt, "ed25519")
	keytools.GenerateSigningKeyPair(p2pPub, p2pPvt, "ed25519")
	keytools.GenerateSigningKeyPair(ethAuthPub, ethAuthPvt, "secp256k1")
}

func addKeys() error {
	err := identity.AddKeyFromConfig(identity.KeyPurposeP2p)
	if err != nil {
		panic(err)
	}
	err = identity.AddKeyFromConfig(identity.KeyPurposeSigning)
	if err != nil {
		panic(err)
	}
	err = identity.AddKeyFromConfig(identity.KeyPurposeEthMsgAuth)
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
				"ethNodeUrl":      ethNodeUrl,
				"bootstraps":      bootstraps,
				"apiPort":         apiPort,
				"p2pPort":         p2pPort,
			}

			v, err := config.CreateConfigFile(data)
			if err != nil {
				panic(err)
			}
			log.Infof("Config File Created: %s\n", v.ConfigFileUsed())

			config.Bootstrap(v.ConfigFileUsed())
			generateKeys()
			baseBootstrap()
			id, err := createIdentity()
			if err != nil {
				panic(err)
			}

			v.Set("identityId", id.String())
			err = v.WriteConfig()
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			config.Config.V.Set("identityId", id.String())

			log.Infof("Identity created [%s] [%x]", id.String(), id.ByteArray())

			err = addKeys()
			if err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	createConfigCmd.Flags().StringVarP(&targetDataDir, "targetdir", "t", home+"/datadir", "Target Data Dir")
	createConfigCmd.Flags().StringVarP(&ethNodeUrl, "ethnodeurl", "e", "ws://127.0.0.1:9546", "URL of Ethereum Client Node (WS only)")
	createConfigCmd.Flags().StringVarP(&accountKeyPath, "accountkeypath", "z", home+"/datadir/main.key", "Path of Ethereum Account Key JSON file")
	createConfigCmd.Flags().StringVarP(&accountPassword, "accountpwd", "k", "", "Ethereum Account Password")
	createConfigCmd.Flags().Int64VarP(&apiPort, "apiPort", "a", 8082, "Api Port")
	createConfigCmd.Flags().Int64VarP(&p2pPort, "p2pPort", "p", 38202, "Peer-to-Peer Port")
	createConfigCmd.Flags().StringVarP(&network, "network", "n", "centrifugeRussianhillEthRinkeby", "Default Network")
	createConfigCmd.Flags().StringSliceVarP(&bootstraps, "bootstraps", "b", nil, "Bootstrap P2P Nodes")
	rootCmd.AddCommand(createConfigCmd)
}
