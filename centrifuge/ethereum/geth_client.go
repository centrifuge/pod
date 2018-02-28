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
	"sync"
)

const (
	// Timeouts
	defaultWaitForTransactionMined = 30 * time.Second
	mainAccountName                = "main"
)

var gc *GethClient
var gcInit sync.Once

func DefaultWaitForTransactionMiningContext() (ctx context.Context) {
	toBeDone := time.Now().Add(defaultWaitForTransactionMined)
	ctx, _ = context.WithDeadline(context.TODO(), toBeDone)
	return
}

// Abstract the "ethereum client" out so we can more easily support other clients
// besides Geth (e.g. quorum)
// Also make it easier to mock tests
type EthereumClient interface {
	GetClient() (*ethclient.Client)
}

type GethClient struct {
	Client *ethclient.Client
}

func (gethClient GethClient) GetClient() (*ethclient.Client) {
	return gethClient.Client
}

func GetConnection() (EthereumClient) {
	gcInit.Do(func() {
		client, err := ethclient.Dial(viper.GetString("ethereum.gethIpc"))
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		} else {
			gc = &GethClient{client}
		}
	})
	return gc
}

// Retrieves the geth transaction options for the given account name. The account name influences which configuration
// is used. If no account name is provided the account as defined by `mainAccountName` constant is used
// It is not supported to call with more than one account name.
func GetGethTxOpts(optionalAccountName ...string) (*bind.TransactOpts, error) {
	var accountName string
	accsLen := len(optionalAccountName)
	if accsLen > 1 {
		err := errors.Errorf("error in use of method. can deal with maximum of one account name for ethereum transaction options. please check your code.")
		log.Fatalf(err.Error())
		return nil, err
	} else {
		switch accsLen {
		case 1:
			accountName = optionalAccountName[0]
		default:
			accountName = mainAccountName
		}
	}

	key := viper.GetString("ethereum.accounts." + accountName + ".key")

	// TODO: this could be done more elegantly if support for additional ways to configure keys should be added later on
	// e.g. if key files would be supported instead of inline keys
	if key == "" {
		err := errors.Errorf("could not find configured ethereum key for account [%v]. please check your configuration.\n", accountName)
		log.Printf(err.Error())
		return nil, err
	}

	password := viper.GetString("ethereum.accounts." + accountName + ".password")

	auth, err := bind.NewTransactor(strings.NewReader(key), password)
	if err != nil {
		err = errors.Errorf("Failed to load key with error: %v", err);
		log.Println(err.Error())
		return nil, err
	} else {
		auth.GasPrice = big.NewInt(viper.GetInt64("ethereum.gasPrice"))
		auth.GasLimit = uint64(viper.GetInt64("ethereum.gasLimit"))
		return auth, nil
	}
}

func GetGethCallOpts() (auth *bind.CallOpts) {
	// Assuring that pending transactions are taken into account by go-ethereum when asking for things like
	// specific transactions and client's nonce
	return &bind.CallOpts{Pending: true}
}
