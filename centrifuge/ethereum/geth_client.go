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
	mainAccountName = "main"
)

var gc *GethClient
var gcInit sync.Once

// getDefaultContextTimeout retrieves the default duration before an Ethereum call context should time out
func getDefaultContextTimeout() (time.Duration) {
	return viper.GetDuration("ethereum.contextWaitTimeout")
}

func DefaultWaitForTransactionMiningContext() (ctx context.Context) {
	toBeDone := time.Now().Add(getDefaultContextTimeout())
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

// GetConnection returns the connection to the configured `ethereum.gethSocket`.
// Note that this is a singleton and is the same connection for the whole application.
func GetConnection() (EthereumClient) {
	gcInit.Do(func() {
		client, err := ethclient.Dial(viper.GetString("ethereum.gethSocket"))
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		} else {
			gc = &GethClient{client}
		}
	})
	return gc
}

// GetGethTxOpts retrieves the geth transaction options for the given account name. The account name influences which configuration
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

	authedTransactionOpts, err := bind.NewTransactor(strings.NewReader(key), password)
	if err != nil {
		err = errors.Errorf("Failed to load key with error: %v", err);
		log.Println(err.Error())
		return nil, err
	} else {
		authedTransactionOpts.GasPrice = big.NewInt(viper.GetInt64("ethereum.gasPrice"))
		authedTransactionOpts.GasLimit = uint64(viper.GetInt64("ethereum.gasLimit"))
		return authedTransactionOpts, nil
	}
}

// TODO - make this lookup smarter so it can really take pending transactions into account
// right now, if there is high concurrency, the code will still error out
// this has to do with the transaction pool taking presedence before even sending transactions to geth
// see go-ethereum/core/tx_pool.go line 639 -> https://github.com/ethereum/go-ethereum/blob/master/core/tx_pool.go#L639
func IncreaseNoncePlus1(opts *bind.TransactOpts) (error) {
	//var ctx context.Context
	//if opts.Context != nil {
	//	ctx = opts.Context
	//} else {
	//	ctx = context.TODO()
	//}
	//client := GetConnection().GetClient()
	//nonce, err := client.PendingNonceAt(ctx, opts.From)
	//if err != nil {
	//	return fmt.Errorf("failed to retrieve pending account nonce: %v", err)
	//}
	//if &nonce == nil {
	//	nonce, err = client.NonceAt(ctx, opts.From, nil)
	//	if err != nil {
	//		return fmt.Errorf("failed to retrieve account nonce from latest block: %v", err)
	//	}
	//}
	//opts.Nonce = big.NewInt(int64(nonce) + 1)

	return nil
}

func GetGethCallOpts() (auth *bind.CallOpts) {
	// Assuring that pending transactions are taken into account by go-ethereum when asking for things like
	// specific transactions and client's nonce
	return &bind.CallOpts{Pending: true}
}
