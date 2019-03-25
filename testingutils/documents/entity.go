package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
)

func CreateEntityData() entitypb.Entity {
	return entitypb.Entity{
		Identity:  testingidentity.GenerateRandomDID().ToAddress().Bytes(),
		LegalName: "Company Test",
		Contacts:  []*entitypb.Contact{{Name: "Satoshi Nakamoto"}},
		Addresses: []*entitypb.Address{{IsMain: true,
			AddressLine1: "Sample Street 1",
			Zip:          "12345",
			State:        "Germany",
		}, {IsMain: false, State: "US"}},
	}
}

func CreateEntityPayload() *cliententitypb.EntityCreatePayload {
	return &cliententitypb.EntityCreatePayload{
		Data: &cliententitypb.EntityData{
			Identity:  testingidentity.GenerateRandomDID().ToAddress().String(),
			LegalName: "Company Test",
			Contacts:  []*entitypb.Contact{{Name: "Satoshi Nakamoto"}},
			Addresses: []*entitypb.Address{{IsMain: true,
				AddressLine1: "Sample Street 1",
				Zip:          "12345",
				State:        "Germany",
			}, {IsMain: false, State: "US"}},
		},
		Collaborators: []string{testingidentity.GenerateRandomDID().String()},
	}
}
