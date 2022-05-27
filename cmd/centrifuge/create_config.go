package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	apiHost, targetDataDir, ethNodeURL, accountKeyPath, network  string
	apiPort, p2pPort                                             int64
	bootstraps                                                   []string
	centChainURL, centChainID, centChainSecret, centChainAddress string
	identityFactoryAddr                                          string
	ipfsPinningServiceURL, ipfsPinningServiceJWT                 string
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
			var contractAddrs *config.SmartContractAddresses
			var ethPassword string
			if network == "testing" {
				contractAddrs = &config.SmartContractAddresses{IdentityFactoryAddr: identityFactoryAddr}
			} else {
				_, err := fmt.Fprintln(os.Stderr, "Enter your Ethereum Account Password:")
				if err != nil {
					log.Fatal(err)
				}

				pwd, err := terminal.ReadPassword(syscall.Stdin)
				if err != nil {
					// lets take empty password
					log.Error(err)
				}

				ethPassword = string(pwd)
			}

			cfgVals := &config.ConfigVals{
				TargetDataDir:         targetDataDir,
				EthNodeURL:            ethNodeURL,
				AccountKeyPath:        accountKeyPath,
				AccountPassword:       ethPassword,
				Network:               network,
				ApiHost:               apiHost,
				ApiPort:               apiPort,
				P2pPort:               p2pPort,
				Bootstraps:            bootstraps,
				PreCommitEnabled:      false,
				P2pConnectionTimeout:  "",
				SmartContractAddrs:    contractAddrs,
				WebhookURL:            "",
				CentChainURL:          centChainURL,
				CentChainID:           centChainID,
				CentChainSecret:       centChainSecret,
				CentChainAddr:         centChainAddress,
				IpfsPinningServiceURL: ipfsPinningServiceURL,
				IpfsPinningServiceJWT: ipfsPinningServiceJWT,
			}

			if err = cmd.CreateConfig(cfgVals); err != nil {
				log.Info(targetDataDir,
					accountKeyPath,
					network,
					ethNodeURL,
					apiPort,
					p2pPort,
					bootstraps)
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
	createConfigCmd.Flags().StringVarP(&network, "network", "n", "flint", "Default Network")
	createConfigCmd.Flags().StringSliceVarP(&bootstraps, "bootstraps", "b", nil, "Bootstrap P2P Nodes")
	createConfigCmd.Flags().StringVar(&centChainURL, "centchainurl", "ws://127.0.0.1:9946", "Centrifuge Chain URL")
	createConfigCmd.Flags().StringVar(&centChainID, "centchainid", "", "Centrifuge Chain Account ID")
	createConfigCmd.Flags().StringVar(&centChainSecret, "centchainsecret", "", "Centrifuge Chain Secret URI")
	createConfigCmd.Flags().StringVar(&centChainAddress, "centchainaddr", "", "Centrifuge Chain ss58addr")
	createConfigCmd.Flags().StringVar(&identityFactoryAddr, "identityFactory", "", "Ethereum Identity factory address for testing network")
	createConfigCmd.Flags().StringVar(&ipfsPinningServiceURL, "ipfsPinningServiceURL", "https://api.pinata.cloud", "URL of the IPFS pinning service")
	createConfigCmd.Flags().StringVar(&ipfsPinningServiceJWT, "ipfsPinningServiceJWT", "", "JWT token used to authenticate with IPFS pinning service")
	rootCmd.AddCommand(createConfigCmd)
}
