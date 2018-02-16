package ethereum

import (
	"log"
	"strings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/spf13/viper"
	"github.com/go-errors/errors"
	"math/big"
)

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

	//TODO this should be more flexible to support any connection type to get, not just IPC
	//client, err := ethclient.Dial(viper.GetString("ethereum.gethIpc"))
	client, err := ethclient.Dial("http://localhost:9545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	return GethClient{client}
}

func GetGethTxOpts() (*bind.TransactOpts, error) {

	//account := "838f7dcA284eb69A9c489fE09c31cFf37DeFDEcA"
	key := `{"address":"0x838f7dca284eb69a9c489fe09c31cff37defdeca","crypto":{"cipher":"aes-128-ctr","ciphertext":"b16312912c00712f02b43ed3cdd3b3172195329415527f7ee218656888aa5d92","cipherparams":{"iv":"19494c514fae0e4d83d9a7e464e89e29"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"e9b7cf9b55eab4a54f6f6f5af98ca6add2ca49147d37f99a5fa26a89e9003517"},"mac":"04805d48727a24cc3ee2ac2198f7fd5be269e52ff105c125cd10b614ce0d856d"},"id":"cd3800bc-c85d-457b-925b-09d809d6b06e","version":3}`
	password := `ZhXfpAc#vHu4JTELA`

	//auth, err := bind.NewTransactor(strings.NewReader(viper.GetString("ethereum.key")), viper.GetString("ethereum.password"))
	auth, err := bind.NewTransactor(strings.NewReader(key), password)
	if err != nil {
		err = errors.Errorf("Failed to load key with error: %v", err);
		log.Fatal(err)

	}
	//nonce, err := auth.
	auth.GasPrice = big.NewInt(40000)
	auth.GasLimit = 4712388
	return auth, err
}

func GetGethCallOpts() (auth *bind.CallOpts) {
	return &bind.CallOpts{Pending: true}
}
