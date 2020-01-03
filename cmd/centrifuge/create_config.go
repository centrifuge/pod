package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	apiHost, targetDataDir, ethNodeURL, accountKeyPath, network string
	apiPort, p2pPort                                            int64
	bootstraps                                                  []string
	txPoolAccess                                                bool
	centChainID, centChainSecret, centChainAddress              string
)

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
			_, err := fmt.Fprintln(os.Stderr, "Enter your Ethereum Account Password:")
			if err != nil {
				log.Fatal(err)
			}

			pwd, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				// lets take empty password
				log.Error(err)
			}

			err = cmd.CreateConfig(
				targetDataDir,
				ethNodeURL,
				accountKeyPath,
				string(pwd),
				network,
				apiHost,
				apiPort,
				p2pPort,
				bootstraps,
				txPoolAccess,
				false,
				"",
				nil,
				"",
				centChainID, centChainSecret, centChainAddress)
			if err != nil {
				log.Info(targetDataDir,
					accountKeyPath,
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
	createConfigCmd.Flags().StringVarP(&apiHost, "nodeHost", "s", "127.0.0.1", "API server host")
	createConfigCmd.Flags().StringVarP(&accountKeyPath, "accountkeypath", "z", home+"/datadir/main.key", "Path of Ethereum Account Key JSON file")
	createConfigCmd.Flags().Int64VarP(&apiPort, "apiPort", "a", 8082, "Api Port")
	createConfigCmd.Flags().Int64VarP(&p2pPort, "p2pPort", "p", 38202, "Peer-to-Peer Port")
	createConfigCmd.Flags().StringVarP(&network, "network", "n", "russianhill", "Default Network")
	createConfigCmd.Flags().StringSliceVarP(&bootstraps, "bootstraps", "b", nil, "Bootstrap P2P Nodes")
	createConfigCmd.Flags().BoolVarP(&txPoolAccess, "txpoolaccess", "x", true, "Transaction Pool access (-x=false)")
	createConfigCmd.Flags().StringVar(&centChainID, "centchainid", "", "Centrifuge Chain Account ID")
	createConfigCmd.Flags().StringVar(&centChainSecret, "centchainsecret", "", "Centrifuge Chain Secret URI")
	createConfigCmd.Flags().StringVar(&centChainAddress, "centchainaddr", "", "Centrifuge Chain ss58addr")
	rootCmd.AddCommand(createConfigCmd)
}
