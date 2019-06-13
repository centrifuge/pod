// +build unit integration

package testingdocuments

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/identity"
	entitypb2 "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func CreateRelationship() *entitypb.EntityRelationship {
	did, _ := identity.StringsToDIDs("0xed03Fa80291fF5DDC284DE6b51E716B130b05e20", "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb")
	return &entitypb.EntityRelationship{
		OwnerIdentity:  did[0][:],
		TargetIdentity: did[1][:],
	}
}

func CreateRelationshipPayload() *entitypb2.RelationshipPayload {
	did2 := "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb"
	entityID := hexutil.Encode(utils.RandomSlice(32))

	return &entitypb2.RelationshipPayload{
		TargetIdentity: did2,
		DocumentId:     entityID,
	}
}
