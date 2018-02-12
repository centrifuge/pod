package testing

import (
	"log"
	"strings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
)


// Actual first implementation of the EthereumClient
type MockGethClient struct {
	Client *ethclient.Client
}

func (gethClient MockGethClient)GetClient()(*ethclient.Client){
	return gethClient.Client
}


func GetConnection() (ethereum.EthereumClient) {

	//TODO this should be more flexible to support any connection type to get, not just IPC
	client, err := ethclient.Dial(viper.GetString("ethereum.gethIpc"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	return GethClient{client}
}