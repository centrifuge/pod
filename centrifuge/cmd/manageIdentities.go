package cmd

import (
	"github.com/spf13/cobra"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"log"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
)

var centrifugeId string

var createIdentityCmd = &cobra.Command{
	Use:   "createidentity",
	Short: "creates identity with signing key as p2p id against ethereum",
	Long:  "creates identity with signing key as p2p id against ethereum",
	Run: func(cmd *cobra.Command, args []string) {
		if centrifugeId == "" {
			panic("Centrifuge ID not provided")
		}
		bCentId, err := tools.StringToByte32(centrifugeId)
		if err != nil {
			panic(err)
		}
		log.Printf("Centrifuge ID: String [%v] in bytes [%v]", centrifugeId, bCentId)

		exists, err := identity.CheckIdentityExists(centrifugeId)
		if err != nil  || exists {
			panic(err)
		}
		publicKey, _ := keytools.GetSigningKeyPairFromConfig()

		log.Printf("P2PKey: %v\n", publicKey)
		var bPk [32]byte
		copy(bPk[:], publicKey)
		pid, err := keytools.PublicKeyToP2PKey(bPk)
		log.Printf("PID: %v\n", pid.Pretty())

		var m = make(map[int][]identity.IdentityKey)
		var pk [32]byte
		copy(pk[:], publicKey)
		m[1] = append(m[1], identity.IdentityKey{pk})
		id := identity.Identity{centrifugeId, m}
		confirmations := make(chan *identity.Identity, 1)

		log.Printf("Creating Identity [%v]", id.CentrifugeId)
		identity.CreateIdentity(id, confirmations)
		registeredIdentity := <-confirmations
		log.Printf("Identity [%v] Created", registeredIdentity.CentrifugeId)

		log.Printf("Adding Key [%v] to Identity [%v]", id.Keys[1][0],id.CentrifugeId)
		identity.AddKeyToIdentity(id, 1, confirmations)
		addedToIdentity := <-confirmations
		log.Printf("Key [%v] Added to Identity [%v]", addedToIdentity.Keys[1][0].String(), addedToIdentity.CentrifugeId)
		log.Printf("Identity ToString method: [%s]", addedToIdentity.String())
	},
}

var addKeyCmd = &cobra.Command{
	Use:   "addkey",
	Short: "add a signing key as p2p id against ethereum",
	Long:  "add a signing key as p2p id against ethereum",
	Run: func(cmd *cobra.Command, args []string) {
		if centrifugeId == "" {
			panic("Centrifuge ID not provided")
		}
		bCentId, err := tools.StringToByte32(centrifugeId)
		if err != nil {
			panic(err)
		}
		log.Printf("Centrifuge ID: String [%v] in bytes [%v]", centrifugeId, bCentId)

		exists, err := identity.CheckIdentityExists(centrifugeId)
		if err != nil  || !exists {
			panic(err)
		}

		currentId, err := identity.ResolveP2PIdentityForId(centrifugeId, 1)

		publicKey, _ := keytools.GetSigningKeyPairFromConfig()

		log.Printf("P2PKey: %v\n", publicKey)
		var bPk [32]byte
		copy(bPk[:], publicKey)
		pid, err := keytools.PublicKeyToP2PKey(bPk)
		log.Printf("PID: %v\n", pid.Pretty())

		if currentId.Keys[1][len(currentId.Keys[1])-1].Key == bPk {
			log.Printf("Key trying to be added already exists as latest. Skipping Update.")
			return
		}

		var m = make(map[int][]identity.IdentityKey)
		var pk [32]byte
		copy(pk[:], publicKey)
		m[1] = append(m[1], identity.IdentityKey{pk})
		id := identity.Identity{centrifugeId, m}
		confirmations := make(chan *identity.Identity, 1)
		log.Printf("Adding Key [%v] to Identity [%v]", id.Keys[1][0],id.CentrifugeId)
		identity.AddKeyToIdentity(id, 1, confirmations)
		addedToIdentity := <-confirmations
		log.Printf("Key [%v] Added to Identity [%v]", addedToIdentity.Keys[1][0].String(), addedToIdentity.CentrifugeId)
		log.Printf("Identity ToString method: [%s]", addedToIdentity.String())
	},
}

func init() {
	createIdentityCmd.Flags().StringVarP(&centrifugeId, "centrifugeid", "i", "", "Centrifuge ID")
	addKeyCmd.Flags().StringVarP(&centrifugeId, "centrifugeid", "i", "", "Centrifuge ID")
	rootCmd.AddCommand(createIdentityCmd)
	rootCmd.AddCommand(addKeyCmd)
}
