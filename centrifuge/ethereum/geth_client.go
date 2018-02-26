package ethereum

import (
	"log"
	"strings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
	"github.com/go-errors/errors"
	"math/big"
	"time"
	"context"
)

const (
	// Timeouts
	defaultWaitForTransactionMined = 30 * time.Second
)


func DefaultWaitForTransactionMiningContext() (ctx context.Context){
	toBeDone := time.Now().Add(defaultWaitForTransactionMined)
	ctx,_ = context.WithDeadline(context.TODO(), toBeDone)
	return
}

// Abstract the "ethereum client" out so we can more easily support other clients
// besides Geth (e.g. quorum)
// Also make it easier to mock tests
type EthereumClient interface {
	GetClient() (*ethclient.Client)
}

// Actual first implementation of the EthereumClient
type GethClient struct {
	Client *ethclient.Client
}

func (gethClient GethClient) GetClient() (*ethclient.Client) {
	return gethClient.Client
}

func GetConnection() (EthereumClient) {
	//TODO this should come from a more centralized configuration service to avoid strewing config gets around
	client, err := ethclient.Dial(viper.GetString("ethereum.gethIpc"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	return GethClient{client}
}

func GetGethTxOpts() (*bind.TransactOpts, error) {
	key := viper.GetString("ethereum.accounts.main.key")
	password := viper.GetString("ethereum.accounts.main.password")

	auth, err := bind.NewTransactor(strings.NewReader(key), password)
	if err != nil {
		err = errors.Errorf("Failed to load key with error: %v", err);
		log.Fatal(err)

	}
	auth.GasPrice = big.NewInt(viper.GetInt64("ethereum.gasPrice"))
	auth.GasLimit = uint64(viper.GetInt64("ethereum.gasLimit"))
	return auth, err
}

func GetGethCallOpts() (auth *bind.CallOpts) {
	// Assuring that pending transactions are taken into account by go-ethereum when asking for things like
	// specific transactions and client's nonce
	return &bind.CallOpts{Pending: true}
}
