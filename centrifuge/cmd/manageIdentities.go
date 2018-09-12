package cmd

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/spf13/cobra"
)

var centrifugeIdString string

var createIdentityCmd = &cobra.Command{
	Use:   "createidentity",
	Short: "creates identity with signing key as p2p id against ethereum",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		defaultBootstrap()
		identityService := identity.EthereumIdentityService{}
		centrifugeId, err := identity.CentrifugeIdStringToSlice(centrifugeIdString)
		if err != nil {
			panic(err)
		}
		_, confirmations, err := identityService.CreateIdentity(centrifugeId)
		watchIdentity := <-confirmations
		if err != nil {
			panic(err)
		}
		log.Infof("Identity created [%s]", watchIdentity.Identity.GetCentrifugeID())
	},
}

//We should support multiple types of keys to add, at the moment only keyPurpose 1 - PeerID/Signature/Encryption
var addKeyCmd = &cobra.Command{
	Use:   "addkey",
	Short: "add a signing key as p2p id against ethereum",
	Long:  "add a signing key as p2p id against ethereum",
	Run: func(cmd *cobra.Command, args []string) {
		defaultBootstrap()
		identityService := identity.EthereumIdentityService{}

		publicKey, _ := ed25519.GetSigningKeyPairFromConfig()
		idKey := []byte{}
		copy(idKey[:], publicKey[:32])

		centrifugeId, err := identity.CentrifugeIdStringToSlice(centrifugeIdString)
		if err != nil {
			panic(err)
		}
		id, err := identityService.LookupIdentityForID(centrifugeId)

		if err != nil {
			panic(err)
		}

		confirmations, err := id.AddKeyToIdentity(identity.KeyPurposeP2p, idKey)
		if err != nil {
			panic(err)
		}
		watchAddedToIdentity := <-confirmations

		lastKey, errLocal := watchAddedToIdentity.Identity.GetLastKeyForPurpose(identity.KeyPurposeP2p)
		if errLocal != nil {
			err = errLocal
			return
		}
		log.Infof("Key [%v] Added to Identity [%s]", lastKey, watchAddedToIdentity.Identity)
		return
	},
}

func init() {
	createIdentityCmd.Flags().StringVarP(&centrifugeIdString, "centrifugeid", "i", "", "Centrifuge ID")
	addKeyCmd.Flags().StringVarP(&centrifugeIdString, "centrifugeid", "i", "", "Centrifuge ID")
	rootCmd.AddCommand(createIdentityCmd)
	rootCmd.AddCommand(addKeyCmd)
}
