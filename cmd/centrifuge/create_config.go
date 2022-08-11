package main

import (
	"os"

	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	apiHost, targetDataDir, network                                       string
	apiPort, p2pPort                                                      int
	bootstraps                                                            []string
	centChainURL                                                          string
	authenticationEnabled                                                 bool
	ipfsPinningServiceName, ipfsPinningServiceURL, ipfsPinningServiceAuth string
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
			err = cmd.CreateConfig(
				targetDataDir,
				network,
				apiHost,
				apiPort,
				p2pPort,
				bootstraps,
				"",
				centChainURL,
				authenticationEnabled,
				ipfsPinningServiceName,
				ipfsPinningServiceURL,
				ipfsPinningServiceAuth,
			)

			if err != nil {
				log.Info(targetDataDir, network, apiPort, p2pPort, bootstraps)
				log.Fatalf("error: %v", err)
			}
		},
	}

	createConfigCmd.Flags().StringVarP(&targetDataDir, "targetdir", "t", home+"/datadir", "Target Data Dir")
	createConfigCmd.Flags().StringVarP(&apiHost, "nodeHost", "s", "127.0.0.1", "API server host")
	createConfigCmd.Flags().IntVarP(&apiPort, "apiPort", "a", 8082, "Api Port")
	createConfigCmd.Flags().IntVarP(&p2pPort, "p2pPort", "p", 38202, "Peer-to-Peer Port")
	createConfigCmd.Flags().StringVarP(&network, "network", "n", "catalyst", "Default Network")
	createConfigCmd.Flags().StringSliceVarP(&bootstraps, "bootstraps", "b", nil, "Bootstrap P2P Nodes")
	createConfigCmd.Flags().StringVar(&centChainURL, "centchainurl", "ws://127.0.0.1:9946", "Centrifuge Chain URL")
	createConfigCmd.Flags().BoolVarP(&authenticationEnabled, "authenticationenabled", "", true, "Enable authentication on the node")
	createConfigCmd.Flags().StringVarP(&ipfsPinningServiceName, "ipfsPinningServiceName", "", "pinata", "Name of the IPFS pinning service")
	createConfigCmd.Flags().StringVarP(&ipfsPinningServiceURL, "ipfsPinningServiceURL", "", "https://api.pinata.cloud", "URL of the IPFS pinning service")
	createConfigCmd.Flags().StringVarP(&ipfsPinningServiceAuth, "ipfsPinningServiceAuth", "", "", "JWT token used to authenticate with IPFS pinning service")
	rootCmd.AddCommand(createConfigCmd)
}
