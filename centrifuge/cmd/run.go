package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/server"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"sync"

)


// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a centrifuge node",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			server.ServeNode()
		}()
		go func() {
			defer wg.Done()
			p2p.RunP2P()
		}()
		wg.Wait()
	},
}

var Destination string


func init() {

	// Set defaults for Server
	viper.SetDefault("nodeHostname", "localhost")
	viper.SetDefault("nodePort", 8022)
	viper.SetDefault("p2p.port", 53202)

	runCmd.Flags().String("destination", "", "Destination to send a message to")
	viper.BindPFlag("p2p.destination", rootCmd.PersistentFlags().Lookup("destination"))

	rootCmd.AddCommand(runCmd)
}
