package cmd

import (
	"sync"

	"github.com/centrifuge/go-centrifuge/centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/server"
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

		defaultBootstrap()
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

func init() {
	rootCmd.AddCommand(runCmd)
}
