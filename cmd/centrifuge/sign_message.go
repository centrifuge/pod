package main

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"
)

func init() {
	var privateKeyFileParam string
	var curveTypeParam string
	var messageParam string

	var signMessageCmd = &cobra.Command{
		Use:   "sign",
		Short: "sign a message with private key",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			privateKey, err := utils.ReadKeyFromPemFile(privateKeyFileParam, utils.PrivateKey)

			if err != nil {
				log.Fatal(err)
			}
			signature, err := crypto.SignMessage(privateKey, []byte(messageParam), curveTypeParam)
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
}
