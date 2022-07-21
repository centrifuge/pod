package main

import (
	"os"

	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	apiHost, targetDataDir, network string
	apiPort, p2pPort                int64
	bootstraps                      []string
	centChainURL                    string
	authenticationEnabled           bool
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
			)

			if err != nil {
				log.Info(targetDataDir, network, apiPort, p2pPort, bootstraps)
				log.Fatalf("error: %v", err)
			}
		},
	}

	createConfigCmd.Flags().StringVarP(&targetDataDir, "targetdir", "t", home+"/datadir", "Target Data Dir")
	createConfigCmd.Flags().StringVarP(&apiHost, "nodeHost", "s", "127.0.0.1", "API server host")
	createConfigCmd.Flags().Int64VarP(&apiPort, "apiPort", "a", 8082, "Api Port")
	createConfigCmd.Flags().Int64VarP(&p2pPort, "p2pPort", "p", 38202, "Peer-to-Peer Port")
	createConfigCmd.Flags().StringVarP(&network, "network", "n", "flint", "Default Network")
	createConfigCmd.Flags().StringSliceVarP(&bootstraps, "bootstraps", "b", nil, "Bootstrap P2P Nodes")
	createConfigCmd.Flags().StringVar(&centChainURL, "centchainurl", "ws://127.0.0.1:9946", "Centrifuge Chain URL")
	createConfigCmd.Flags().BoolVarP(&authenticationEnabled, "authenticationenabled", "a", true, "Enable authentication on the node")
	rootCmd.AddCommand(createConfigCmd)
}
