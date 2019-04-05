package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/identity"
)

func CreateEntityRelationshipData() entitypb.EntityRelationship {
	did, _ := identity.NewDIDFromString("0xed03Fa80291fF5DDC284DE6b51E716B130b05e20")
	did2, _ := identity.NewDIDFromString("0x5F9132e0F92952abCb154A9b34563891ffe1AAcb")
	return entitypb.EntityRelationship{
		OwnerIdentity:  did[:],
		TargetIdentity: did2[:],
	}
}
