package cmd

import (
	"github.com/lucasvo/go-centrifuge/centrifuge/signingkeys"
	"github.com/spf13/cobra"
)

var privateKeyFile string
var publicKeyFile string

// generateSigningKeyCmd represents the generateSigningKey command
var generateSigningKeyCmd = &cobra.Command{
	Use:   "generateSigningKey",
	Short: "generate a signing key for use with centrifuge documents",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		_, _= signingkeys.GenerateKeypair(publicKeyFile, privateKeyFile)
	},
}

func init() {
	rootCmd.AddCommand(generateSigningKeyCmd)
	generateSigningKeyCmd.Flags().StringVarP(&privateKeyFile, "private", "p", "", "Private key path")
	generateSigningKeyCmd.Flags().StringVarP(&publicKeyFile, "public", "q", "", "Public key path")
}
