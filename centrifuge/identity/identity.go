package identity

import (
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"log"
	"fmt"
)

type Identity struct {
	CentrifugeId string
	Keys map[int][][32]byte
}

func (id *Identity) String() string {
	return fmt.Sprintf("CentrifugeId [%v], Key [%v]", id.CentrifugeId, id.Keys)
}

func (id *Identity) LastB58Key(keyType int) (ret string, err error) {
	if len(id.Keys[keyType]) == 0 {
		return
	}
	switch keyType {
	case 0:
		log.Printf("Error not authorized type")
	case 1:
		p2pId, err1 := keytools.PublicKeyToP2PKey(id.Keys[keyType][len(id.Keys[keyType])-1])
		if err1 != nil {
			err = err1
			return
		}
		ret = p2pId.Pretty()
	default:
		log.Printf("keyType Not found")
	}
	return
}

func ResolveIdentityForKey(centrifugeId string, keyType int) (id Identity, err error) {
	if (viper.GetBool("identity.ethereum.enabled")) {
		id, err = doResolveIdentityForKey(centrifugeId, keyType)
	}
	return
}

func CreateIdentity(identity Identity, confirmations chan<- *Identity) (err error) {
	if (viper.GetBool("identity.ethereum.enabled")) {
		err = doCreateIdentity(identity, confirmations)
	}
	return
}

func AddKeyToIdentity(identity Identity, keyType int, confirmations chan<- *Identity) (err error) {
	if (viper.GetBool("identity.ethereum.enabled")) {
		err = doAddKeyToIdentity(identity, keyType, confirmations)
	}
	return
}