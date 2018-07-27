package cmd

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/spf13/cobra"
)

var centrifugeIdString string

var createIdentityCmd = &cobra.Command{
	Use:   "createidentity",
	Short: "creates identity with signing key as p2p id against ethereum",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		identityService := identity.EthereumIdentityService{}
		centrifugeId, err := identity.CentrifugeIdStringToSlice(centrifugeIdString)
		if err != nil {
			panic(err)
		}
		confirmations := make(chan *identity.WatchIdentity, 1)
		_, err = identityService.CreateIdentity(centrifugeId, confirmations)
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
		identityService := identity.EthereumIdentityService{}

		publicKey, _ := keytools.GetSigningKeyPairFromConfig()
		idKey := []byte{}
		copy(idKey[:], publicKey[:32])

		centrifugeId, err := identity.CentrifugeIdStringToSlice(centrifugeIdString)
		if err != nil {
			panic(err)
		}
		id, err := identityService.LookupIdentityForId(centrifugeId)

		if err != nil {
			panic(err)
		}

		confirmations := make(chan *identity.WatchIdentity, 1)
		err = id.AddKeyToIdentity(identity.KEY_TYPE_PEERID, idKey, confirmations)
		if err != nil {
			panic(err)
		}
		watchAddedToIdentity := <-confirmations

		lastKey, errLocal := watchAddedToIdentity.Identity.GetLastKeyForType(identity.KEY_TYPE_PEERID)
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
