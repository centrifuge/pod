package ethereum

import (
	"context"
	"fmt"
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
	logging "github.com/ipfs/go-log"
)

const (
	transactionUnderpriced = "replacement transaction underpriced"
	nonceTooLow            = "nonce too low"
)

var log = logging.Logger("geth-client")
var gc Client
var gcMu sync.RWMutex

// GetDefaultContextTimeout retrieves the default duration before an Ethereum write call context should time out
func GetDefaultContextTimeout() time.Duration {
	return config.Config.GetEthereumContextWaitTimeout()
}

// defaultReadContext returns context with timeout for read operations
func defaultReadContext() (ctx context.Context, cancelFunc context.CancelFunc) {
	toBeDone := time.Now().Add(config.Config.GetEthereumContextReadWaitTimeout())
	return context.WithDeadline(context.Background(), toBeDone)
}

// DefaultWaitForTransactionMiningContext returns context with timeout for write operations
func DefaultWaitForTransactionMiningContext() (ctx context.Context, cancelFunc context.CancelFunc) {
	toBeDone := time.Now().Add(GetDefaultContextTimeout())
	return context.WithDeadline(context.Background(), toBeDone)
}

// Config defines functions to get ethereum details
type Config interface {
	GetEthereumGasPrice() *big.Int
	GetEthereumGasLimit() uint64
	GetEthereumNodeURL() string
	GetEthereumAccount(accountName string) (account *config.AccountConfig, err error)
	GetEthereumIntervalRetry() time.Duration
	GetEthereumMaxRetries() int
	GetTxPoolAccessEnabled() bool
}

// Client can be implemented by any chain client
type Client interface {
	GetClient() *ethclient.Client
	GetRPCClient() *rpc.Client
	GetHost() *url.URL
	GetTxOpts(accountName string) (*bind.TransactOpts, error)
	SubmitTransactionWithRetries(contractMethod interface{}, opts *bind.TransactOpts, params ...interface{}) (tx *types.Transaction, err error)
}

// GethClient implements Client for Ethereum
type GethClient struct {
	client    *ethclient.Client
	rpcClient *rpc.Client
	host      *url.URL
	accounts  map[string]*bind.TransactOpts
	accMu     sync.Mutex // accMu to protect accounts
	config    Config

	// nonceMu to ensure one transaction at a time
	nonceMu sync.Mutex
}

// NewGethClient returns an GethClient which implements Client
func NewGethClient(config Config) (Client, error) {
	log.Info("Opening connection to Ethereum:", config.GetEthereumNodeURL())
	u, err := url.Parse(config.GetEthereumNodeURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ethereum node URL: %v", err)
	}

	c, err := rpc.Dial(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %v", err)
	}

	return &GethClient{
		client:    ethclient.NewClient(c),
		rpcClient: c,
		host:      u,
		accounts:  make(map[string]*bind.TransactOpts),
		nonceMu:   sync.Mutex{},
		accMu:     sync.Mutex{},
		config:    config,
	}, nil
}

// SetClient sets the Client
// Note that this is a singleton and is the same connection for the whole application.
func SetClient(client Client) {
	gcMu.Lock()
	defer gcMu.Unlock()
	gc = client
}

// GetClient returns the current Client
func GetClient() Client {
	gcMu.RLock()
	defer gcMu.RUnlock()
	return gc
}

// GetTxOpts returns a cached options if available else creates and returns new options
func (gc *GethClient) GetTxOpts(accountName string) (*bind.TransactOpts, error) {
	gc.accMu.Lock()
	defer gc.accMu.Unlock()

	if _, ok := gc.accounts[accountName]; ok {
		opts := gc.accounts[accountName]

		// Important to nil the nonce on the cached txopts, otherwise with high concurrency will be outdated
		opts.Nonce = nil
		return opts, nil
	}

	txOpts, err := gc.getGethTxOpts(accountName)
	if err != nil {
		return nil, err
	}

	gc.accounts[accountName] = txOpts
	return txOpts, nil
}

// GetClient returns the ethereum client
func (gc *GethClient) GetClient() *ethclient.Client {
	return gc.client
}

// GetRPCClient returns the RPC client
func (gc *GethClient) GetRPCClient() *rpc.Client {
	return gc.rpcClient
}

// GetHost returns the node url
func (gc *GethClient) GetHost() *url.URL {
	return gc.host
}

// getGethTxOpts retrieves the geth transaction options for the given account name. The account name influences which configuration
// is used.
func (gc *GethClient) getGethTxOpts(accountName string) (*bind.TransactOpts, error) {
	account, err := gc.config.GetEthereumAccount(accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get ethereum account: %v", err)
	}

	opts, err := bind.NewTransactor(strings.NewReader(account.Key), account.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to create new transaction opts: %v", err)
	}

	opts.GasPrice = gc.config.GetEthereumGasPrice()
	opts.GasLimit = gc.config.GetEthereumGasLimit()
	opts.Context = context.Background()
	return opts, nil
}

/**
SubmitTransactionWithRetries submits transaction to the ethereum chain
Blocking Function that sends transaction using reflection wrapped in a retrial block. It is based on the transactionUnderpriced error,
meaning that a transaction is being attempted to run twice, and the logic is to override the existing one. As we have constant
gas prices that means that a concurrent transaction race condition event has happened.
- contractMethod: Contract Method that implements GenericEthereumAsset (usually autogenerated binding from abi)
- params: Arbitrary number of parameters that are passed to the function fname call
Note: contractMethod must always return "*types.Transaction, error"
*/
func (gc *GethClient) SubmitTransactionWithRetries(contractMethod interface{}, opts *bind.TransactOpts, params ...interface{}) (*types.Transaction, error) {
	var current int
	f := reflect.ValueOf(contractMethod)
	maxTries := gc.config.GetEthereumMaxRetries()

	gc.nonceMu.Lock()
	defer gc.nonceMu.Unlock()

	for {
		if current >= maxTries {
			return nil, fmt.Errorf("max concurrent transaction tries reached")
		}

		current++
		err := gc.incrementNonce(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to increment nonce: %v", err)
		}

		log.Infof("Incrementing Nonce to [%v]\n", opts.Nonce.String())
		var in []reflect.Value
		in = append(in, reflect.ValueOf(opts))
		for _, p := range params {
			in = append(in, reflect.ValueOf(p))
		}

		result := f.Call(in)
		var tx *types.Transaction
		if !result[0].IsNil() {
			tx = result[0].Interface().(*types.Transaction)
		}

		if !result[1].IsNil() {
			err = result[1].Interface().(error)
		}

		if err == nil {
			return tx, nil
		}

		if (err.Error() == transactionUnderpriced) || (err.Error() == nonceTooLow) {
			log.Warningf("Concurrent transaction identified, trying again [%d/%d]\n", current, maxTries)
			time.Sleep(gc.config.GetEthereumIntervalRetry())
			continue
		}

		return nil, err
	}

}

// incrementNonce updates the opts.Nonce to next valid nonce
// We pick the current nonce by getting latest transactions included in the blocks including pending blocks
// then we check the txpool to see if there any new transactions from the address that are not included in any block
// If there are no pending transactions in txpool, we use the current nonce
// else we increment the greater of current nonce and txpool calculated nonce
func (gc *GethClient) incrementNonce(opts *bind.TransactOpts) error {
	// get current nonce
	ctx, cancel := context.WithTimeout(context.Background(), GetDefaultContextTimeout())
	defer cancel()

	n, err := gc.client.PendingNonceAt(ctx, opts.From)
	if err != nil {
		return fmt.Errorf("failed to get chain nonce for %s: %v", opts.From.String(), err)
	}

	// set the nonce
	opts.Nonce = new(big.Int).Add(new(big.Int).SetUint64(n), big.NewInt(1))

	// check if the txpool access is enabled
	if !gc.config.GetTxPoolAccessEnabled() {
		log.Warningf("Ethereum Client doesn't support txpool API, may cause concurrency issues.")
		return nil
	}

	// check for any transactions in txpool
	var res map[string]map[string]map[string][]string
	err = gc.rpcClient.CallContext(ctx, &res, "txpool_inspect")
	if err != nil {
		return fmt.Errorf("failed to get txpool data: %v", err)
	}

	// no pending transaction from this account in tx pool
	if len(res["pending"][opts.From.Hex()]) < 1 {
		return nil
	}

	var keys []int
	for k, _ := range res["pending"][opts.From.Hex()] {
		ki, err := strconv.Atoi(k)
		if err != nil {
			return fmt.Errorf("failed to convert nonce: %v", err)
		}

		keys = append(keys, ki)
	}

	// there are some pending transactions in txpool, check their nonce
	// pick the largest one and increment it
	sort.Ints(keys)
	lastPoolNonce := keys[len(keys)-1]
	if uint64(lastPoolNonce) >= n {
		opts.Nonce = new(big.Int).Add(big.NewInt(int64(lastPoolNonce)), big.NewInt(1))
	}

	return nil
}

// GetGethCallOpts returns the Call options with default
func GetGethCallOpts() (*bind.CallOpts, context.CancelFunc) {
	// Assuring that pending transactions are taken into account by go-ethereum when asking for things like
	// specific transactions and client's nonce
	// with timeout context, in case eth node is not in sync
	ctx, cancel := defaultReadContext()
	return &bind.CallOpts{Pending: true, Context: ctx}, cancel
}
