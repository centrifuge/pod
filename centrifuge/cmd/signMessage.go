package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/spf13/cobra"
)

func init() {

	//specific param
	var privateKeyFileParam string
	var curveTypeParam string
	var messageParam string


	var signMessageCmd = &cobra.Command{
		Use:   "sign",
		Short: "sign a message with private key",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			signature := keytools.SignMessage(privateKeyFileParam, messageParam, curveTypeParam)
			fmt.Println(hex.EncodeToString(signature))

		},
	}

	rootCmd.AddCommand(signMessageCmd)
	signMessageCmd.Flags().StringVarP(&messageParam, "message", "m", "", "message to sign (max 32 bytes)")
	signMessageCmd.Flags().StringVarP(&privateKeyFileParam, "private", "p", "", "private key path")
	signMessageCmd.Flags().StringVarP(&curveTypeParam, "type", "t", "", "type of the curve (default: 'secp256k1') ")
}
