package cmd

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/server"
	"github.com/spf13/cobra"
	"sync"

	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a centrifuge node",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cc.Bootstrap()
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
	rootCmd.AddCommand(runCmd)
}
