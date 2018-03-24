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
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		if centrifugeId == "" {
			panic("Centrifuge ID not provided")
		}
		bCentId, err := tools.StringToByte32(centrifugeId)
		if err != nil {
			panic(err)
		}
		log.Printf("Centrifuge ID: String [%v] in bytes [%v]", centrifugeId, bCentId)
		publicKey, _ := keytools.GetSigningKeysFromConfig()

		log.Printf("P2PKey: %v\n", publicKey)
		var bPk [32]byte
		copy(bPk[:], publicKey)
		pid, err := keytools.PublicKeyToP2PKey(bPk)
		log.Printf("PID: %v\n", pid.Pretty())

		var m = make(map[int][][32]byte)
		var pk [32]byte
		copy(pk[:], publicKey)
		m[1] = append(m[1], pk)
		id := identity.Identity{centrifugeId, m}
		confirmations := make(chan *identity.Identity, 1)

		log.Printf("Creating Identity [%v]", id.CentrifugeId)
		identity.CreateIdentity(id, confirmations)
		registeredIdentity := <-confirmations
		log.Printf("Identity [%v] Created", registeredIdentity.CentrifugeId)

		log.Printf("Adding Key [%v] to Identity [%v]", id.Keys[1][0],id.CentrifugeId)
		identity.AddKeyToIdentity(id, 1, confirmations)
		addedToIdentity := <-confirmations
		log.Printf("Key [%v] Added to Identity [%v]", addedToIdentity.Keys[1][0], addedToIdentity.CentrifugeId)
	},
}

func init() {
	createIdentityCmd.Flags().StringVarP(&centrifugeId, "centrifugeid", "i", "", "Centrifuge ID")
	rootCmd.AddCommand(createIdentityCmd)
}
