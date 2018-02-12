package cmd

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/spf13/cobra"
)

var privateKeyFile string
var publicKeyFile string
var createSigningKey bool
var createEncryptionKey bool

// generateSigningKeyCmd represents the generateSigningKey command
var generateSigningKeyCmd = &cobra.Command{
	Use:   "gemeratekeys",
	Short: "generate key sfor use with centrifuge documents",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		if createSigningKey {
		_, _= keytools.GenerateSigningKeypair(publicKeyFile, privateKeyFile)
		}
		if createEncryptionKey {
			panic("Not implemented")
		}
	},
}

func init() {
	generateSigningKeyCmd.Flags().BoolVarP(&createSigningKey, "signing",  "s", true, "Generate signing keys")
	generateSigningKeyCmd.Flags().BoolVarP(&createEncryptionKey, "encryption", "e", false, "Generate encryption keys")
	generateSigningKeyCmd.Flags().StringVarP(&privateKeyFile, "private", "p", "", "Private key path")
	generateSigningKeyCmd.Flags().StringVarP(&publicKeyFile, "public", "q", "", "Public key path")
	rootCmd.AddCommand(generateSigningKeyCmd)
}
