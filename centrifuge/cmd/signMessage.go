package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"encoding/hex"
)

var signMessageCmd = &cobra.Command{
	Use:   "sign",
	Short: "sign a message with private key",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		signature := keytools.SignMessage(privateKeyFile,message, curveType)
		fmt.Println(hex.EncodeToString(signature))

	},
}

func init() {
	rootCmd.AddCommand(signMessageCmd)
	signMessageCmd.Flags().StringVarP(&message, "message", "m", "", "message to sign (max 32 bytes)")
	signMessageCmd.Flags().StringVarP(&privateKeyFile, "private", "p", "", "private key path")
	signMessageCmd.Flags().StringVarP(&curveType, "type", "t", "", "type of the curve (default: 'secp256k1') ")
}

