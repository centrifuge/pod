package cmd

import (
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print centrifuge version",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("go-centrifuge version", version.GetVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
