package did

import (
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/ethereum/go-ethereum/common"
	id "github.com/centrifuge/go-centrifuge/identity"
)

// DID stores the identity address of the user
type DID common.Address

func (d DID) toAddress() common.Address {
	return common.Address(d)
}

// NewDID returns a DID based on a common.Address
func NewDID(address common.Address) DID {
	return DID(address)
}

// NewDIDFromString returns a DID based on a hex string
func NewDIDFromString(address string) DID {
	return DID(common.HexToAddress(address))
}


type Identity interface {

}

type contract interface {

}


type identity struct  {
	contract contract
	config id.Config
	client ethereum.Client
}

func NewIdentity(config id.Config, client ethereum.Client) identity {
	return identity{config:config,client:client}
}


