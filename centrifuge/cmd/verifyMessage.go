package cmd

import (
	"github.com/spf13/cobra"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"fmt"
	"encoding/hex"
)

var verifyMsgCmd  = &cobra.Command{
	Use:   "verify",
	Short: "verify a signature",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		signatureBytes, err := hex.DecodeString(signature)

		if(err != nil){
			log.Fatal(err)
		}
		correctSignature := keytools.VerifyMessage(publicKeyFile,message,signatureBytes,curveType)
		fmt.Println(correctSignature)
	},
}

func init() {
	rootCmd.AddCommand(verifyMsgCmd)
	verifyMsgCmd.Flags().StringVarP(&message, "message", "m", "", "message to verify (max 32 bytes)")
	verifyMsgCmd.Flags().StringVarP(&publicKeyFile, "public", "q", "", "public key path")
	verifyMsgCmd.Flags().StringVarP(&curveType, "type", "t", "", "type of the curve (default: 'secp256k1') ")
	verifyMsgCmd.Flags().StringVarP(&signature, "signature", "s", "", "signature")
}


