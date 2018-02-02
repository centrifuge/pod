package witness

import (
	"log"
	"strings"

	"github.com/CentrifugeInc/centrifuge-ethereum-contracts/centrifuge/witness"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
)

func getConnection() (client *ethclient.Client) {
	client, err := ethclient.Dial(viper.GetString("witness.ethereum.geth"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	return
}

func GetGethTxOpts() (auth *bind.TransactOpts) {
	auth, err := bind.NewTransactor(strings.NewReader(viper.GetString("witness.ethereum.key")), viper.GetString("witness.ethereum.password"))
	if err != nil {
		log.Fatalf("Failed to load key")
	}
	return
}

func GetGethCallOpts() (auth *bind.CallOpts) {
	return &bind.CallOpts{Pending: true}
}

func GetWitnessContract() (witnessContract *witness.EthereumWitness) {
	// Instantiate the contract and display its name
	client := getConnection()
	witnessContract, err := witness.NewEthereumWitness(common.HexToAddress(viper.GetString("witness.ethereum.contractAddress")), client)
	if err != nil {
		log.Fatalf("Failed to instantiate the witness contract contract: %v", err)
	}
	return
}
