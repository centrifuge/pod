package ethereum

import (
	"log"
	"strings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
)

func GetConnection() (client *ethclient.Client) {
	//TODO this should be more flexible to support any connection type to get, not just IPC
	client, err := ethclient.Dial(viper.GetString("ethereum.gethIpc"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	return
}

func GetGethTxOpts() (auth *bind.TransactOpts) {
	auth, err := bind.NewTransactor(strings.NewReader(viper.GetString("ethereum.key")), viper.GetString("ethereum.password"))
	if err != nil {
		log.Fatalf("Failed to load key")
	}
	return
}

func GetGethCallOpts() (auth *bind.CallOpts) {
	return &bind.CallOpts{Pending: true}
}