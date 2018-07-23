package cmd

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/spf13/cobra"
)

var centrifugeIdString string

var createIdentityCmd = &cobra.Command{
	Use:   "createidentity",
	Short: "creates identity with signing key as p2p id against ethereum",
	Long:  "creates identity with signing key as p2p id against ethereum",
	Run: func(cmd *cobra.Command, args []string) {
		publicKey, _ := keytools.GetSigningKeyPairFromConfig()
		idKey, err := tools.SliceToByte32(publicKey)
		if err != nil {
			panic(err)
		}
		centrifugeId, err := identity.CentrifugeIdStringToSlice(centrifugeIdString)
		if err != nil {
			panic(err)
		}
		err = identity.CreateEthereumIdentityFromApi(centrifugeId, idKey)
		if err != nil {
			panic(err)
		}
	},
}

//We should support multiple types of keys to add, at the moment only keyType 1 - PeerID/Signature/Encryption
var addKeyCmd = &cobra.Command{
	Use:   "addkey",
	Short: "add a signing key as p2p id against ethereum",
	Long:  "add a signing key as p2p id against ethereum",
	Run: func(cmd *cobra.Command, args []string) {
		publicKey, _ := keytools.GetSigningKeyPairFromConfig()
		idKey, err := tools.SliceToByte32(publicKey)
		if err != nil {
			panic(err)
		}
		centrifugeId, err := identity.CentrifugeIdStringToSlice(centrifugeIdString)
		if err != nil {
			panic(err)
		}
		err = identity.AddKeyToIdentityFromApi(centrifugeId, identity.KEY_TYPE_PEERID, idKey)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	createIdentityCmd.Flags().StringVarP(&centrifugeIdString, "centrifugeid", "i", "", "Centrifuge ID")
	addKeyCmd.Flags().StringVarP(&centrifugeIdString, "centrifugeid", "i", "", "Centrifuge ID")
	rootCmd.AddCommand(createIdentityCmd)
	rootCmd.AddCommand(addKeyCmd)
}
