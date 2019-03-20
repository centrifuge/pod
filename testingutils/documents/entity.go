package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
)

func CreateEntityData() entitypb.Entity {
	return entitypb.Entity{
		Identity: testingidentity.GenerateRandomDID().ToAddress().Bytes(),
		LegalName:"Company Test",
		Addresses:[]*entitypb.Address{{IsMain:true,State:"Germany"},{IsMain:false,State:"US"}},
	}
}

func CreateEntityPayload() *cliententitypb.EntityCreatePayload {
	return &cliententitypb.EntityCreatePayload{
		Data: &cliententitypb.EntityData{
			Identity: testingidentity.GenerateRandomDID().ToAddress().Bytes(),
			LegalName:"Company Test",
			Addresses:[]*entitypb.Address{{IsMain:true,State:"Germany"},{IsMain:false,State:"US"}},
		},
		Collaborators: []string{testingidentity.GenerateRandomDID().String()},
	}
}
