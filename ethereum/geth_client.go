package ethereum

import (
	"context"
	"math/big"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-errors/errors"
	logging "github.com/ipfs/go-log"
)

const TransactionUnderpriced = "replacement transaction underpriced"
const NonceTooLow = "nonce too low"

var log = logging.Logger("geth-client")
var gc EthereumClient
var gcInit sync.Once

// GetDefaultContextTimeout retrieves the default duration before an Ethereum write call context should time out
func GetDefaultContextTimeout() time.Duration {
	return config.Config.GetEthereumContextWaitTimeout()
}

// GetDefaultReadContextTimeout retrieves the default duration before an Ethereum read call context should time out
func GetDefaultReadContextTimeout() time.Duration {
	return config.Config.GetEthereumContextReadWaitTimeout()
}

// DefaultWaitForReadContext returns context with timeout for read operations
func DefaultWaitForReadContext() (ctx context.Context, cancelFunc context.CancelFunc) {
	toBeDone := time.Now().Add(GetDefaultReadContextTimeout())
	return context.WithDeadline(context.Background(), toBeDone)
}

// DefaultWaitForTransactionMiningContext returns context with timeout for write operations
func DefaultWaitForTransactionMiningContext() (ctx context.Context, cancelFunc context.CancelFunc) {
	toBeDone := time.Now().Add(GetDefaultContextTimeout())
	return context.WithDeadline(context.TODO(), toBeDone)
}

type Config interface {
	GetEthereumGasPrice() *big.Int
	GetEthereumGasLimit() uint64
	GetEthereumNodeURL() string
	GetEthereumAccount(accountName string) (account *config.AccountConfig, err error)
	GetEthereumIntervalRetry() time.Duration
	GetEthereumMaxRetries() int
	GetTxPoolAccessEnabled() bool
}

// Abstract the "ethereum client" out so we can more easily support other clients
// besides Geth (e.g. quorum)
// Also make it easier to mock tests
type EthereumClient interface {
	GetClient() *ethclient.Client
	GetRpcClient() *rpc.Client
	GetHost() *url.URL
	GetTxOpts(accountName string) (*bind.TransactOpts, error)
	SubmitTransactionWithRetries(contractMethod interface{}, opts *bind.TransactOpts, params ...interface{}) (tx *types.Transaction, err error)
}

type GethClient struct {
	Client     *ethclient.Client
	RpcClient  *rpc.Client
	Host       *url.URL
	Accounts   map[string]*bind.TransactOpts
	nonceMutex sync.Mutex
	config     Config
}

func NewGethClient(config Config) *GethClient {
	return &GethClient{
		nil, nil, nil, make(map[string]*bind.TransactOpts), sync.Mutex{}, config,
	}
}

func (gethClient *GethClient) GetTxOpts(accountName string) (*bind.TransactOpts, error) {
	if _, ok := gethClient.Accounts[accountName]; !ok {
		txOpts, err := gethClient.getGethTxOpts(accountName)
		if err != nil {
			return nil, err
		}
		gethClient.Accounts[accountName] = txOpts
	}
	gethClient.Accounts[accountName].Nonce = nil // Important to nil the nonce on the cached txopts, otherwise with high concurrency will be outdated
	return gethClient.Accounts[accountName], nil
}

func (gethClient *GethClient) GetClient() *ethclient.Client {
	return gethClient.Client
}

func (gethClient *GethClient) GetRpcClient() *rpc.Client {
	return gethClient.RpcClient
}

func (gethClient *GethClient) GetHost() *url.URL {
	return gethClient.Host
}

func NewClientConnection(config Config) (*GethClient, error) {
	log.Info("Opening connection to Ethereum:", config.GetEthereumNodeURL())
	u, err := url.Parse(config.GetEthereumNodeURL())
	if err != nil {
		log.Errorf("Failed to connect to parse ethereum.gethSocket URL: %v", err)
		return &GethClient{}, err
	}
	c, err := rpc.Dial(u.String())
	if err != nil {
		log.Errorf("Failed to connect to the Ethereum client [%s]: %v", u.String(), err)
		return &GethClient{}, err
	}
	client := ethclient.NewClient(c)
	if err != nil {
		log.Errorf("Failed to connect to the Ethereum client [%s]: %v", u.String(), err)
		return &GethClient{}, err
	}

	gethClient := NewGethClient(config)
	gethClient.Client = client
	gethClient.RpcClient = c
	gethClient.Host = u

	return gethClient, nil
}

// Note that this is a singleton and is the same connection for the whole application.
func SetConnection(conn EthereumClient) {
	gcInit.Do(func() {
		gc = conn
	})
	return
}

// GetConnection returns the connection to the configured `ethereum.gethSocket`.
func GetConnection() EthereumClient {
	return gc
}

// getGethTxOpts retrieves the geth transaction options for the given account name. The account name influences which configuration
// is used.
func (gethClient *GethClient) getGethTxOpts(accountName string) (*bind.TransactOpts, error) {
	account, err := gethClient.config.GetEthereumAccount(accountName)
	if err != nil {
		err = errors.Errorf("could not find configured ethereum key for account [%v]. please check your configuration.\n", accountName)
		log.Error(err.Error())
		return nil, err
	}

	authedTransactionOpts, err := bind.NewTransactor(strings.NewReader(account.Key), account.Password)
	if err != nil {
		err = errors.Errorf("Failed to load key with error: %v", err)
		log.Error(err.Error())
		return nil, err
	} else {
		authedTransactionOpts.GasPrice = gethClient.config.GetEthereumGasPrice()
		authedTransactionOpts.GasLimit = gethClient.config.GetEthereumGasLimit()
		authedTransactionOpts.Context = context.Background()
		return authedTransactionOpts, nil
	}
}

/**
Blocking Function that sends transaction using reflection wrapped in a retrial block. It is based on the TransactionUnderpriced error,
meaning that a transaction is being attempted to run twice, and the logic is to override the existing one. As we have constant
gas prices that means that a concurrent transaction race condition event has happened.
- contractMethod: Contract Method that implements GenericEthereumAsset (usually autogenerated binding from abi)
- params: Arbitrary number of parameters that are passed to the function fname call
*/
func (gethClient *GethClient) SubmitTransactionWithRetries(contractMethod interface{}, opts *bind.TransactOpts, params ...interface{}) (tx *types.Transaction, err error) {
	done := false
	maxTries := gethClient.config.GetEthereumMaxRetries()
	current := 0
	var f reflect.Value
	var in []reflect.Value
	var result []reflect.Value
	f = reflect.ValueOf(contractMethod)

	gethClient.nonceMutex.Lock()
	defer gethClient.nonceMutex.Unlock()

	for !done {
		if current >= maxTries {
			log.Error("Max Concurrent transaction tries reached")
			break
		}
		current += 1

		err = gethClient.incrementNonce(opts)
		if err != nil {
			return
		}
		in = make([]reflect.Value, len(params)+1)
		val := reflect.ValueOf(opts)
		in[0] = val
		for k, param := range params {
			in[k+1] = reflect.ValueOf(param)
		}
		result = f.Call(in)
		tx = result[0].Interface().(*types.Transaction)
		err = nil
		if result[1].Interface() != nil {
			err = result[1].Interface().(error)
		}

		if err != nil {
			if (err.Error() == TransactionUnderpriced) || (err.Error() == NonceTooLow) {
				log.Warningf("Concurrent transaction identified, trying again [%d/%d]\n", current, maxTries)
				time.Sleep(gethClient.config.GetEthereumIntervalRetry())
			} else {
				done = true
			}
		} else {
			done = true
		}
	}

	return
}

func (gethClient *GethClient) incrementNonce(opts *bind.TransactOpts) (err error) {
	if !gethClient.config.GetTxPoolAccessEnabled() {
		log.Warningf("Ethereum Client doesn't support txpool API, may cause concurrency issues.")
		return
	}

	var res map[string]map[string]map[string][]string
	// Important to not create a DeadLock if network latency
	txctx, _ := context.WithTimeout(context.Background(), GetDefaultContextTimeout())
	gc.GetRpcClient().CallContext(txctx, &res, "txpool_inspect")

	if len(res["pending"][opts.From.Hex()]) > 0 {
		chainNonce, err := gc.GetClient().PendingNonceAt(txctx, opts.From)
		if err != nil {
			log.Errorf("Error found when getting the Account Chain Nonce [%v]\n", err)
			return err
		}
		CalculateIncrement(chainNonce, res, opts)
	}

	return
}

func CalculateIncrement(chainNonce uint64, res map[string]map[string]map[string][]string, opts *bind.TransactOpts) {
	keys := make([]int, 0, len(res["pending"][opts.From.Hex()]))
	for k, _ := range res["pending"][opts.From.Hex()] {
		ki, _ := strconv.Atoi(k)
		keys = append(keys, ki)
	}
	sort.Ints(keys)
	lastPoolNonce := keys[len(keys)-1]
	if uint64(lastPoolNonce) >= chainNonce {
		opts.Nonce = new(big.Int).Add(big.NewInt(int64(lastPoolNonce)), big.NewInt(1))
		log.Infof("Incrementing Nonce to [%v]\n", opts.Nonce.String())
	}
}

func GetGethCallOpts() (*bind.CallOpts, context.CancelFunc) {
	// Assuring that pending transactions are taken into account by go-ethereum when asking for things like
	// specific transactions and client's nonce
	// with timeout context, in case eth node is not in sync
	ctx, cancelFunc := DefaultWaitForReadContext()
	return &bind.CallOpts{Pending: true, Context: ctx}, cancelFunc
}
