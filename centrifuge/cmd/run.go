package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/server"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"sync"

	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
)


// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a centrifuge node",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		(&cc.CentNode{}).BootstrapDependencies()
		//node.BootstrapDependencies()
		// Below WaitGroup not the best solution to the concurrency issue. If one of the two methods return because of
		// an error it will have to be killed manually and restarted.
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
	rootCmd.AddCommand(runCmd)
}
