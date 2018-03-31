package cmd

import (
	"github.com/spf13/cobra"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
)

var centrifugeId string

var createIdentityCmd = &cobra.Command{
	Use:   "createidentity",
	Short: "creates identity with signing key as p2p id against ethereum",
	Long:  "creates identity with signing key as p2p id against ethereum",
	Run: func(cmd *cobra.Command, args []string) {
		err := identity.ManageEthereumIdentity(centrifugeId, identity.ACTION_CREATE)
		if err != nil {
			panic(err)
		}
	},
}

var addKeyCmd = &cobra.Command{
	Use:   "addkey",
	Short: "add a signing key as p2p id against ethereum",
	Long:  "add a signing key as p2p id against ethereum",
	Run: func(cmd *cobra.Command, args []string) {
		err := identity.ManageEthereumIdentity(centrifugeId, identity.ACTION_ADDKEY)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	createIdentityCmd.Flags().StringVarP(&centrifugeId, "centrifugeid", "i", "", "Centrifuge ID")
	addKeyCmd.Flags().StringVarP(&centrifugeId, "centrifugeid", "i", "", "Centrifuge ID")
	rootCmd.AddCommand(createIdentityCmd)
	rootCmd.AddCommand(addKeyCmd)
}
