package main

import (
	"os"

	"github.com/centrifuge/go-centrifuge/cmd"
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
		Run: func(c *cobra.Command, args []string) {
			err := cmd.CreateConfigDeprecated(targetDataDir,
				ethNodeURL,
				accountKeyPath,
				accountPassword,
				network,
				apiPort,
				p2pPort,
				bootstraps,
				txPoolAccess,
				"",
				nil)
			if err != nil {
				log.Info(targetDataDir,
					accountKeyPath,
					accountPassword,
					network,
					ethNodeURL,
					apiPort,
					p2pPort,
					bootstraps,
					txPoolAccess)
				log.Fatalf("error: %v", err)
			}
		},
	}

	createConfigCmd.Flags().StringVarP(&targetDataDir, "targetdir", "t", home+"/datadir", "Target Data Dir")
	createConfigCmd.Flags().StringVarP(&ethNodeURL, "ethnodeurl", "e", "http://127.0.0.1:9545", "URL of Ethereum Client Node")
	createConfigCmd.Flags().StringVarP(&accountKeyPath, "accountkeypath", "z", home+"/datadir/main.key", "Path of Ethereum Account Key JSON file")
	createConfigCmd.Flags().StringVarP(&accountPassword, "accountpwd", "k", "", "Ethereum Account Password")
	createConfigCmd.Flags().Int64VarP(&apiPort, "apiPort", "a", 8082, "Api Port")
	createConfigCmd.Flags().Int64VarP(&p2pPort, "p2pPort", "p", 38202, "Peer-to-Peer Port")
	createConfigCmd.Flags().StringVarP(&network, "network", "n", "russianhill", "Default Network")
	createConfigCmd.Flags().StringSliceVarP(&bootstraps, "bootstraps", "b", nil, "Bootstrap P2P Nodes")
	createConfigCmd.Flags().BoolVarP(&txPoolAccess, "txpoolaccess", "x", true, "Transaction Pool access")
	rootCmd.AddCommand(createConfigCmd)
}
