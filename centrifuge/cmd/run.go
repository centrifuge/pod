package cmd

import (
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a centrifuge node",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		//cmd requires a config file
		readConfigFile()
		// the following call will block
		defaultBootstrap()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
