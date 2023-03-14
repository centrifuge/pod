package main

import (
	"github.com/centrifuge/pod/crypto"
	"github.com/spf13/cobra"
)

func init() {
	// specific param
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
				err := crypto.GenerateSigningKeyPair(publicKeyFileParam, privateKeyFileParam, crypto.CurveType(curveTypeParam))
				if err != nil {
					log.Fatal(err)
				}
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
	generateSigningKeyCmd.Flags().StringVarP(&curveTypeParam, "type", "t", "", "type of the curve (supported: 'ed25519')")
	rootCmd.AddCommand(generateSigningKeyCmd)
}
