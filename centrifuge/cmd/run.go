package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/server"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a centrifuge node",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		server.ServeNode()
	},
}

func init() {
	// Set defaults for Server
	viper.SetDefault("nodeHostname", "localhost")
	viper.SetDefault("nodePort", 8022)

	rootCmd.AddCommand(runCmd)
}
