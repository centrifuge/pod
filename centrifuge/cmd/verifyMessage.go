package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/spf13/cobra"
)



func init() {

	//specific param
	var curveTypeParam string
	var messageParam string
	var signatureParam string
	var publicKeyFileParam string

	var verifyMsgCmd = &cobra.Command{
		Use:   "verify",
		Short: "verify a signature",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			signatureBytes, err := hex.DecodeString(signatureParam)

			if err != nil {
				log.Fatal(err)
			}
			correctSignature := keytools.VerifyMessage(publicKeyFileParam, messageParam, signatureBytes, curveTypeParam)
			fmt.Println(correctSignature)
		},
	}


	rootCmd.AddCommand(verifyMsgCmd)
	verifyMsgCmd.Flags().StringVarP(&messageParam, "message", "m", "", "message to verify (max 32 bytes)")
	verifyMsgCmd.Flags().StringVarP(&publicKeyFileParam, "public", "q", "", "public key path")
	verifyMsgCmd.Flags().StringVarP(&curveTypeParam, "type", "t", "", "type of the curve (default: 'secp256k1') ")
	verifyMsgCmd.Flags().StringVarP(&signatureParam, "signature", "s", "", "signature")
}
