package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/identity"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
)

func CreateEntityData() entitypb.Entity {
	did, _ := identity.NewDIDFromString("0xed03Fa80291fF5DDC284DE6b51E716B130b05e20")
	return entitypb.Entity{
		Identity:  did[:],
		LegalName: "Company Test",
		Contacts:  []*entitypb.Contact{{Name: "Satoshi Nakamoto"}},
		/*Addresses: []*entitypb.Address{{IsMain: true,  TODO: precise proofs should support addresses
			AddressLine1: "Sample Street 1",
			Zip:          "12345",
			State:        "Germany",
		}, {IsMain: false, State: "US"}},*/
	}
}

func CreateEntityPayload() *cliententitypb.EntityCreatePayload {
	return &cliententitypb.EntityCreatePayload{
		Data: &cliententitypb.EntityData{
			Identity:  "0xed03Fa80291fF5DDC284DE6b51E716B130b05e20",
			LegalName: "Company Test",
			Contacts:  []*entitypb.Contact{{Name: "Satoshi Nakamoto"}},
			/*Addresses: []*entitypb.Address{{IsMain: true, TODO: precise proofs should support addresses
				AddressLine1: "Sample Street 1",
				Zip:          "12345",
				State:        "Germany",
			}, {IsMain: false, State: "US"}}, */
		},
		Collaborators: []string{testingidentity.GenerateRandomDID().String()},
	}
}
