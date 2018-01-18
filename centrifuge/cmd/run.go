package cmd

import (
	"github.com/lucasvo/go-centrifuge/centrifuge/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a centrifuge node",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: move the viper stuff into RunNode.
		server.RunNode(viper.GetString("constellation.key"), viper.GetString("constellation.socket"), viper.GetString("constellation.configuration"), viper.GetString("nodePort"))
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
