package cmd

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/spf13/cobra"
)

var privateKeyFile string
var publicKeyFile string
var createSigningKey bool
var createEncryptionKey bool
var curveType string

// generateSigningKeyCmd represents the generateSigningKey command
var generateSigningKeyCmd = &cobra.Command{
	Use:   "generatekeys",
	Short: "generate key for use with centrifuge documents",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if createSigningKey {
			keytools.GenerateSigningKeyPair(publicKeyFile, privateKeyFile,curveType)
		}
		if createEncryptionKey {
			panic("Not implemented")
		}
	},
}

func init() {
	generateSigningKeyCmd.Flags().BoolVarP(&createSigningKey, "signing", "s", true, "Generate signing keys")
	generateSigningKeyCmd.Flags().BoolVarP(&createEncryptionKey, "encryption", "e", false, "Generate encryption keys")
	generateSigningKeyCmd.Flags().StringVarP(&privateKeyFile, "private", "p", "", "Private key path")
	generateSigningKeyCmd.Flags().StringVarP(&publicKeyFile, "public", "q", "", "Public key path")
	generateSigningKeyCmd.Flags().StringVarP(&curveType, "type", "t", "", "type of the curve (default: 'ed25519') ")
	rootCmd.AddCommand(generateSigningKeyCmd)
}
