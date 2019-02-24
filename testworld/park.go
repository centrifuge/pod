// +build testworld

package testworld

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers"

	"testing"

	"time"

	"net/http"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/gavv/httpexpect"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("host")

var hostConfig = []struct {
	name             string
	apiPort, p2pPort int64
	multiAccount     bool
}{
	{"Alice", 8084, 38204, false},
	{"Bob", 8085, 38205, true},
	{"Charlie", 8086, 38206, true},
	{"Kenny", 8087, 38207, false},
}

const defaultP2PTimeout = "10s"

// hostTestSuite encapsulates test utilities on top of each host
type hostTestSuite struct {
	name       string
	host       *host
	id         identity.DID
	httpExpect *httpexpect.Expect
}

// hostManager is the hostManager of the hosts at Testworld (Robert)
type hostManager struct {

	// network settings
	ethNodeUrl, accountKeyPath, accountPassword, network, twConfigName string

	txPoolAccess bool

	// contractAddresses are the addresses of centrifuge contracts on Ethereum
	contractAddresses *config.SmartContractAddresses

	// contractBytecode are the bytecode of centrifuge contracts on Ethereum
	contractBytecode *config.SmartContractBytecode

	// bernard is the bootnode for all the hosts
	bernard *host

	// niceHosts are the happy and nice hosts at the Testworld such as Teddy
	niceHosts map[string]*host

	// tempHosts are hosts created at runtime, they should be part of niceHosts/naughtyHosts as well
	tempHosts map[string]*host

	// TODO create evil hosts such as William (or Eve)

	// canc is the cancel signal for all hosts
	canc context.CancelFunc

	// TODO: context should be removed from hostManager
	// currently needed to restart a node
	// parent context
	cancCtx context.Context
}

func newHostManager(
	ethNodeUrl, accountKeyPath, accountPassword, network, twConfigName string,
	txPoolAccess bool,
	smartContractAddrs *config.SmartContractAddresses,
	smartContractBytecode *config.SmartContractBytecode) *hostManager {
	return &hostManager{
		ethNodeUrl:        ethNodeUrl,
		accountKeyPath:    accountKeyPath,
		accountPassword:   accountPassword,
		twConfigName:      twConfigName,
		network:           network,
		txPoolAccess:      txPoolAccess,
		contractAddresses: smartContractAddrs,
		contractBytecode:  smartContractBytecode,
		niceHosts:         make(map[string]*host),
		tempHosts:         make(map[string]*host),
	}
}

func (r *hostManager) reLive(t *testing.T, name string) {
	r.startHost(name)
	// wait for the host to be live, here its 11 seconds allowed but the host should come alive before that and this will return faster
	ok, err := r.getHost(name).isLive(11 * time.Second)
	if ok {
		return
	} else {
		t.Error(err)
	}
}

func (r *hostManager) startHost(name string) {
	go r.niceHosts[name].live(r.cancCtx)
}

func (r *hostManager) init(createConfig bool) error {
	r.cancCtx, r.canc = context.WithCancel(context.Background())
	r.bernard = r.createHost("Bernard", r.twConfigName, defaultP2PTimeout, 8081, 38201, createConfig, false, nil)
	err := r.bernard.init()
	if err != nil {
		return err
	}

	// start and wait for Bernard since other hosts depend on him
	go r.bernard.live(r.cancCtx)
	_, err = r.bernard.isLive(10 * time.Second)
	if err != nil {
		return errors.New("bernard couldn't be made alive %v", err)
	}

	bootnode, err := r.bernard.p2pURL()
	if err != nil {
		return err
	}

	// start hosts
	for _, h := range hostConfig {
		r.niceHosts[h.name] = r.createHost(h.name, r.twConfigName, defaultP2PTimeout, h.apiPort, h.p2pPort, createConfig, h.multiAccount, []string{bootnode})

		err := r.niceHosts[h.name].init()
		if err != nil {
			return err
		}
		r.startHost(h.name)

	}
	// make sure hosts are alive and print host centIDs
	for name, host := range r.niceHosts {
		_, err = host.isLive(10 * time.Second)
		if err != nil {
			return errors.New("%s couldn't be made alive %v", host.name, err)
		}
		i, err := host.id()
		if err != nil {
			return err
		}
		fmt.Printf("CentID for %s is %s \n", name, i)
		if createConfig {
			_ = host.createAccounts(r.getHostTestSuite(&testing.T{}, host.name).httpExpect)
		}
		_ = host.loadAccounts(r.getHostTestSuite(&testing.T{}, host.name).httpExpect)
	}
	return nil
}

func (r *hostManager) getHost(name string) *host {
	if h, ok := r.niceHosts[name]; ok {
		return h
	}
	return nil
}

func (r *hostManager) stop() {
	r.canc()
}

func (r *hostManager) addNiceHost(name string, host *host) {
	r.niceHosts[name] = host
}

func (r *hostManager) createTempHost(name, twConfigName, p2pTimeout string, apiPort, p2pPort int64, createConfig, multiAccount bool, bootstraps []string) *host {
	tempHost := r.createHost(name, twConfigName, p2pTimeout, apiPort, p2pPort, createConfig, multiAccount, bootstraps)
	r.tempHosts[name] = tempHost
	return tempHost
}

func (r *hostManager) startTempHost(name string) error {
	tempHost, ok := r.tempHosts[name]
	if !ok {
		return errors.New("host %s not found as temp host", name)
	}
	err := tempHost.init()
	if err != nil {
		return err
	}
	go tempHost.live(r.cancCtx)

	return nil
}

func (r *hostManager) createHost(name, twConfigName, p2pTimeout string, apiPort, p2pPort int64, createConfig, multiAccount bool, bootstraps []string) *host {
	return newHost(
		name,
		r.ethNodeUrl,
		r.accountKeyPath,
		r.accountPassword,
		r.network,
		twConfigName,
		p2pTimeout,
		apiPort, p2pPort, bootstraps,
		r.txPoolAccess,
		createConfig,
		multiAccount,
		r.contractAddresses,
		r.contractBytecode,
	)
}

func (r *hostManager) getHostTestSuite(t *testing.T, name string) hostTestSuite {
	host := r.getHost(name)
	expect := host.createHttpExpectation(t)
	id, err := host.id()
	if err != nil {
		t.Error(err)
	}
	return hostTestSuite{name: name, host: host, id: id, httpExpect: expect}

}

type host struct {
	name, dir, ethNodeUrl, accountKeyPath, accountPassword, network,
	identityFactoryAddr, identityRegistryAddr, anchorRepositoryAddr, paymentObligationAddr, p2pTimeout string
	apiPort, p2pPort      int64
	bootstrapNodes        []string
	bootstrappedCtx       map[string]interface{}
	txPoolAccess          bool
	smartContractAddrs    *config.SmartContractAddresses
	smartContractBytecode *config.SmartContractBytecode
	config                config.Configuration
	identity              identity.DID
	idService             identity.ServiceDID
	node                  *node.Node
	canc                  context.CancelFunc
	createConfig          bool
	multiAccount          bool
	accounts              []string
}

func newHost(
	name, ethNodeUrl, accountKeyPath, accountPassword, network, twConfigName, p2pTimeout string,
	apiPort, p2pPort int64,
	bootstraps []string,
	txPoolAccess, createConfig, multiAccount bool,
	smartContractAddrs *config.SmartContractAddresses,
	smartContractBytecode *config.SmartContractBytecode,
) *host {
	return &host{
		name:                  name,
		ethNodeUrl:            ethNodeUrl,
		accountKeyPath:        accountKeyPath,
		accountPassword:       accountPassword,
		network:               network,
		apiPort:               apiPort,
		p2pPort:               p2pPort,
		p2pTimeout:            p2pTimeout,
		bootstrapNodes:        bootstraps,
		txPoolAccess:          txPoolAccess,
		smartContractAddrs:    smartContractAddrs,
		smartContractBytecode: smartContractBytecode,
		dir:                   fmt.Sprintf("hostconfigs/%s/%s", twConfigName, name),
		createConfig:          createConfig,
		multiAccount:          multiAccount,
	}
}

func (h *host) init() error {
	if h.createConfig {
		err := cmd.CreateConfig(h.dir, h.ethNodeUrl, h.accountKeyPath, h.accountPassword, h.network, h.apiPort, h.p2pPort, h.bootstrapNodes, h.txPoolAccess, h.p2pTimeout, h.smartContractAddrs, h.smartContractBytecode)
		if err != nil {
			return err
		}
	}

	m := bootstrappers.MainBootstrapper{}
	m.PopulateBaseBootstrappers()
	h.bootstrappedCtx = map[string]interface{}{
		config.BootstrappedConfigFile: h.dir + "/config.yaml",
	}
	err := m.Bootstrap(h.bootstrappedCtx)
	if err != nil {
		return err
	}
	h.config = h.bootstrappedCtx[bootstrap.BootstrappedConfig].(config.Configuration)
	idBytes, err := h.config.GetIdentityID()
	if err != nil {
		return err
	}
	h.identity = identity.NewDIDFromBytes(idBytes)
	h.idService = h.bootstrappedCtx[identity.BootstrappedDIDService].(identity.ServiceDID)
	if err != nil {
		return err
	}
	return nil
}

func (h *host) live(c context.Context) error {
	srvs, err := node.GetServers(h.bootstrappedCtx)
	if err != nil {
		return errors.New("failed to load servers: %v", err)
	}

	h.node = node.New(srvs)
	feedback := make(chan error)
	// may be we can pass a context that exists in c here
	cancCtx, canc := context.WithCancel(context.WithValue(c, bootstrap.NodeObjRegistry, h.bootstrappedCtx))

	// cancel func of individual host
	h.canc = canc

	go h.node.Start(cancCtx, feedback)
	controlC := make(chan os.Signal, 1)
	signal.Notify(controlC, os.Interrupt)
	select {
	case err := <-feedback:
		log.Info(h.name+" encountered error ", err)
		return err
	case sig := <-controlC:
		log.Info(h.name+" shutting down because of ", sig)
		canc()
		err := <-feedback
		return err
	}

}

func (h *host) kill() {
	h.canc()
}

// isLive waits for host to come alive until the given soft timeout has passed, or the hard timeout of 10s is passed
func (h *host) isLive(softTimeOut time.Duration) (bool, error) {
	sig := make(chan error)
	c := createInsecureClient()
	go func(sig chan<- error) {
		var fErr error
		// wait upto 10 seconds(hard timeout) for the host to be live
		for i := 0; i < 10; i++ {
			res, err := c.Get(fmt.Sprintf("https://localhost:%d/ping", h.config.GetServerPort()))
			fErr = err
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			if res.StatusCode == http.StatusOK {
				sig <- nil
				return
			}
		}
		sig <- fErr
	}(sig)
	t := time.After(softTimeOut)
	select {
	case <-t:
		return false, errors.New("host failed to live even after %f seconds", softTimeOut.Seconds())
	case err := <-sig:
		if err != nil {
			return false, err
		}
		return true, nil
	}
}

func (h *host) mintNFT(e *httpexpect.Expect, auth string, status int, inv map[string]interface{}) (*httpexpect.Object, error) {
	return mintNFT(e, auth, status, inv), nil
}

func (h *host) createAccounts(e *httpexpect.Expect) error {
	if !h.multiAccount {
		return nil
	}
	// create 3 accounts
	for i := 0; i < 3; i++ {
		res := generateAccount(e, h.identity.String(), http.StatusOK)
		res.Value("identity_id").String().NotEmpty()
	}
	return nil
}

func (h *host) loadAccounts(e *httpexpect.Expect) error {
	res := getAllAccounts(e, h.identity.String(), http.StatusOK)
	accounts := res.Value("data").Array()
	accIDs := getAccounts(accounts)
	keys := make([]string, 0, len(accIDs))
	for k := range accIDs {
		keys = append(keys, k)
	}
	h.accounts = keys
	return nil
}

func (h *host) createHttpExpectation(t *testing.T) *httpexpect.Expect {
	return createInsecureClientWithExpect(t, fmt.Sprintf("https://localhost:%d", h.config.GetServerPort()))
}

func (h *host) id() (identity.DID, error) {
	return h.identity, nil
}

func (h *host) p2pURL() (string, error) {
	lastB58Key, err := h.idService.CurrentP2PKey(h.identity)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/ipfs/%s", h.p2pPort, lastB58Key), nil
}
