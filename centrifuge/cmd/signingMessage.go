package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var signingMessage = &cobra.Command{
	Use:   "sign",
	Short: "sign a message with private key",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not supported yet")
	},
}

func init() {
	rootCmd.AddCommand(signingMessage)
}
