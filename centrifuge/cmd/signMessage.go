package cmd

import (
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"github.com/spf13/cobra"
)

func init() {

	//specific param
	var privateKeyFileParam string
	var curveTypeParam string
	var messageParam string
	var ethereumSignFlag bool

	var signMessageCmd = &cobra.Command{
		Use:   "sign",
		Short: "sign a message with private key",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			signature := keytools.SignMessage(privateKeyFileParam, messageParam, curveTypeParam, ethereumSignFlag)
			fmt.Println(utils.ByteArrayToHex(signature))

		},
	}

	rootCmd.AddCommand(signMessageCmd)
	signMessageCmd.Flags().StringVarP(&messageParam, "message", "m", "", "message to sign")
	signMessageCmd.Flags().StringVarP(&privateKeyFileParam, "private", "p", "", "private key path")
	signMessageCmd.Flags().StringVarP(&curveTypeParam, "type", "t", "", "type of the curve (supported:'secp256k1')")
	signMessageCmd.Flags().BoolVarP(&ethereumSignFlag, "ethereum", "e", false, "sign message according to Ethereum specification")
}
