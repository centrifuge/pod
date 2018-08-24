package cmd

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/spf13/cobra"
)



func init() {

	//specific param
	var privateKeyFileParam string
	var publicKeyFileParam string
	var createSigningKeyParam bool
	var createEncryptionKeyParam bool
	var curveTypeParam string


	// generateSigningKeyCmd represents the generateSigningKey command
	var generateSigningKeyCmd = &cobra.Command{
		Use:   "generatekeys",
		Short: "generate key for use with centrifuge documents",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if createSigningKeyParam {
				keytools.GenerateSigningKeyPair(publicKeyFileParam, privateKeyFileParam, curveTypeParam)
			}
			if createEncryptionKeyParam {
				panic("Not implemented")
			}
		},
	}

	generateSigningKeyCmd.Flags().BoolVarP(&createSigningKeyParam, "signing", "s", true, "Generate signing keys")
	generateSigningKeyCmd.Flags().BoolVarP(&createEncryptionKeyParam, "encryption", "e", false, "Generate encryption keys")
	generateSigningKeyCmd.Flags().StringVarP(&privateKeyFileParam, "private", "p", "", "Private key path")
	generateSigningKeyCmd.Flags().StringVarP(&publicKeyFileParam, "public", "q", "", "Public key path")
	generateSigningKeyCmd.Flags().StringVarP(&curveTypeParam, "type", "t", "", "type of the curve (default: 'ed25519') ")
	rootCmd.AddCommand(generateSigningKeyCmd)
}
