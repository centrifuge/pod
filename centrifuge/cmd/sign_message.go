package cmd

import (
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
			privateKey, err := utils.ReadKeyFromPemFile(privateKeyFileParam, utils.PrivateKey)

			if err != nil {
				log.Fatal(err)
			}
			signature, err := keytools.SignMessage(privateKey, []byte(messageParam), curveTypeParam, ethereumSignFlag)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(hexutil.Encode(signature))

		},
	}

	rootCmd.AddCommand(signMessageCmd)
	signMessageCmd.Flags().StringVarP(&messageParam, "message", "m", "", "message to sign")
	signMessageCmd.Flags().StringVarP(&privateKeyFileParam, "private", "p", "", "private key path")
	signMessageCmd.Flags().StringVarP(&curveTypeParam, "type", "t", "", "type of the curve (supported:'secp256k1')")
	signMessageCmd.Flags().BoolVarP(&ethereumSignFlag, "ethereum", "e", false, "sign message according to Ethereum specification")
}
