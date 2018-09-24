package cmd

import "github.com/spf13/cobra"

func init() {
	var createConfigCmd = &cobra.Command{
		Use:   "create-config",
		Short: "sign a message with private key",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	rootCmd.AddCommand(createConfigCmd)
}
