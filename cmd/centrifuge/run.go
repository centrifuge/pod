package main

import (
	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/spf13/cobra"
)

var runMigrations bool

func init() {
	// runCmd represents the run command
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "run a centrifuge node",
		Long:  ``,
		Run: func(cm *cobra.Command, args []string) {
			// cm requires a config file
			cfgFile := ensureConfigFile()

			// Check if migrations should run
			if runMigrations {
				err := doMigrate()
				if err != nil {
					log.Fatal(err)
				}
			}

			// the following call will block
			cmd.RunBootstrap(cfgFile)
		},
	}

	runCmd.Flags().BoolVarP(&runMigrations, "runmigrations", "m", true, "Run Migrations at startup (-m=false)")
	rootCmd.AddCommand(runCmd)
}
