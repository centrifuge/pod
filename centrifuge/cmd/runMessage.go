package cmd

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/messageserver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runMessageCmd = &cobra.Command{
	Use:   "run-message",
	Short: "run a centrifuge message node",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: move the viper stuff into RunNode.
		messageserver.RunNode(viper.GetString("constellation.key"), viper.GetString("constellation.socket"), viper.GetString("constellation.configuration"), viper.GetString("nodePort"))
	},
}

func init() {
	rootCmd.AddCommand(runMessageCmd)
}
