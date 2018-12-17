package main

import (
	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a centrifuge node",
	Long:  ``,
	Run: func(cm *cobra.Command, args []string) {
		//cm requires a config file
		cfgFile := ensureConfigFile()
		// the following call will block
		cmd.RunBootstrap(cfgFile)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
