//go:build testworld
// +build testworld

package testworld

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"testing"
	"time"

	v2 "github.com/centrifuge/go-centrifuge/identity/v2"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/p2p"
	mockdoc "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gavv/httpexpect"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("host")

var hostConfig = []struct {
	name         string
	multiAccount bool
}{
	{"Alice", false},
	{"Bob", true},
	{"Charlie", true},
	{"Kenny", false},
	{"Eve", false},
	// Mallory has a mock document.Serivce to facilitate some Byzantine test
	{"Mallory", false},
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
	// bernard is the bootnode for all the hosts
	bernard *host

	// maeve is the webhook receiver for all hosts
	maeve *webhookReceiver

	// niceHosts are the happy and nice hosts at the Testworld such as Teddy
	niceHosts map[string]*host

	// tempHosts are hosts created at runtime, they should be part of niceHosts/naughtyHosts as well
	tempHosts map[string]*host

	// canc is the cancel signal for all hosts
	canc context.CancelFunc

	// currently needed to restart a node
	// parent context
	cancCtx context.Context

	// Dapp Smart contract Addresses
	dappAddresses map[string]string

	config networkConfig
}

func newHostManager(config networkConfig) *hostManager {
	return &hostManager{
		config:        config,
		niceHosts:     make(map[string]*host),
		tempHosts:     make(map[string]*host),
		dappAddresses: config.DappAddresses,
	}
}

func (r *hostManager) reLive(t *testing.T, name string) {
	r.startHost(name)
	// wait for the host to be live, here its 11 seconds allowed but the host should come alive before that and this will return faster
	ok, err := r.getHost(name).isLive(11 * time.Second)
	if ok {
		return
	}
	t.Error(err)
}

func (r *hostManager) startHost(name string) {
	go r.niceHosts[name].live(r.cancCtx)
}

func (r *hostManager) init() error {
	r.cancCtx, r.canc = context.WithCancel(context.Background())

	// start listening to webhooks
	_, port, err := utils.GetFreeAddrPort()
	if err != nil {
		return err
	}

	r.maeve = newWebhookReceiver(port, "/webhook")
	go r.maeve.start(r.cancCtx)

	_, apiPort, err := utils.GetFreeAddrPort()
	if err != nil {
		return err
	}

	_, p2pPort, err := utils.GetFreeAddrPort()
	if err != nil {
		return err
	}

	r.bernard = r.createHost("Bernard", "", defaultP2PTimeout, int64(apiPort), int64(p2pPort),
		r.config.CreateHostConfigs,
		false, nil)

	if err = r.bernard.init(); err != nil {
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
		m := r.maeve.url()
		_, apiPort, err := utils.GetFreeAddrPort()
		if err != nil {
			return fmt.Errorf("failed to get free port for api: %w", err)
		}

		_, p2pPort, err := utils.GetFreeAddrPort()
		if err != nil {
			return fmt.Errorf("failed to get free port for p2p: %w", err)
		}
		r.niceHosts[h.name] = r.createHost(h.name, m, defaultP2PTimeout, int64(apiPort), int64(p2pPort),
			r.config.CreateHostConfigs,
			h.multiAccount, []string{bootnode})
		err = r.niceHosts[h.name].init()
		if err != nil {
			return err
		}
		r.startHost(h.name)
	}
	// make sure hosts are alive and print host centIDs
	for name, host := range r.niceHosts {
		// Temporary until we have a proper healthcheck in place
		time.Sleep(2 * time.Second)
		_, err = host.isLive(10 * time.Second)
		if err != nil {
			return errors.New("%s couldn't be made alive %v", host.name, err)
		}
		i, err := host.id()
		if err != nil {
			return err
		}
		fmt.Printf("DID for %s is %s \n", name, i)
		if r.config.CreateHostConfigs {
			err = host.createAccounts(r.maeve, r.getHostTestSuite(&testing.T{}, host.name).httpExpect)
			if err != nil {
				return err
			}
		}
		err = host.loadAccounts(r.getHostTestSuite(&testing.T{}, host.name).httpExpect)
		if err != nil {
			return err
		}

		dids := append(host.accounts, host.identity.ToHexString())
		for _, did := range dids {
			acc, err := host.configService.GetAccount(common.HexToAddress(did).Bytes())
			if err != nil {
				return err
			}

			// hack to update the webhooks since the we pick a random port for webhook everytime
			a := acc.(*configstore.Account)
			a.WebhookURL = r.maeve.url()
			_, err = host.configService.UpdateAccount(a)
			if err != nil {
				return err
			}
		}
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

func (r *hostManager) createTempHost(name, p2pTimeout string, apiPort, p2pPort int64, createConfig, multiAccount bool, bootstraps []string) *host {
	tempHost := r.createHost(name, r.maeve.url(), p2pTimeout, apiPort, p2pPort, createConfig, multiAccount, bootstraps)
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

func (r *hostManager) createHost(name, webhookURL string, p2pTimeout string, apiPort, p2pPort int64, createConfig, multiAccount bool, bootstraps []string) *host {
	return &host{
		name:             name,
		ethNodeUrl:       r.config.EthNodeURL,
		webhookURL:       webhookURL,
		accountKeyPath:   r.config.EthAccountKeyPath,
		accountPassword:  r.config.EthAccountPassword,
		network:          r.config.Network,
		apiHost:          "0.0.0.0",
		apiPort:          apiPort,
		p2pPort:          p2pPort,
		p2pTimeout:       p2pTimeout,
		bootstrapNodes:   bootstraps,
		dir:              fmt.Sprintf("hostconfigs/%s/%s", r.config.Network, name),
		createConfig:     createConfig,
		multiAccount:     multiAccount,
		centChainAddress: r.config.CentChainS58Address,
		centChainID:      r.config.CentChainAccountID,
		centChainSecret:  r.config.CentChainSecret,
		centChainURL:     r.config.CentChainURL,
		dappAddresses:    r.dappAddresses,
	}
}

func (r *hostManager) getHostTestSuite(t *testing.T, name string) hostTestSuite {
	host := r.getHost(name)
	expect := host.createHTTPExpectation(t)
	id, err := host.id()
	if err != nil {
		t.Error(err)
	}
	return hostTestSuite{name: name, host: host, id: id, httpExpect: expect}
}

type host struct {
	name, dir, ethNodeUrl, webhookURL, accountKeyPath, accountPassword, network, apiHost,
	identityFactoryAddr, identityRegistryAddr, anchorRepositoryAddr, invoiceUnpaidAddr, p2pTimeout string
	apiPort, p2pPort                                             int64
	bootstrapNodes                                               []string
	bootstrappedCtx                                              map[string]interface{}
	config                                                       config.Configuration
	identity                                                     identity.DID
	idService                                                    v2.Service
	node                                                         *node.Node
	canc                                                         context.CancelFunc
	createConfig                                                 bool
	multiAccount                                                 bool
	accounts                                                     []string
	p2pClient                                                    documents.Client
	configService                                                config.Service
	anchorSrv                                                    anchors.Service
	entityService                                                entity.Service
	centChainURL, centChainID, centChainAddress, centChainSecret string
	dappAddresses                                                map[string]string
	nftAPI                                                       nft.API
}

func (h *host) init() error {
	if h.createConfig {
		err := cmd.CreateConfig(
			h.dir, h.ethNodeUrl, h.accountKeyPath, h.accountPassword,
			h.network, h.apiHost, h.apiPort, h.p2pPort, h.bootstrapNodes, false, h.p2pTimeout,
			h.webhookURL,
			h.centChainURL)
		if err != nil {
			return err
		}

		values := map[string]interface{}{
			"ethereum.accounts.main.key":      os.Getenv("CENT_ETHEREUM_ACCOUNTS_MAIN_KEY"),
			"ethereum.accounts.main.password": os.Getenv("CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD"),
		}
		err = updateConfig(h.dir, values)
		if err != nil {
			return err
		}
	} else {
		values := map[string]interface{}{"notifications.endpoint": h.webhookURL}
		err := updateConfig(h.dir, values)
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

	if h.name == "Mallory" {
		malloryDocMockSrv := new(mockdoc.MockService)
		h.bootstrappedCtx["BootstrappedDocumentService"] = malloryDocMockSrv
		p2pBoot := p2p.Bootstrapper{}
		err := p2pBoot.Bootstrap(h.bootstrappedCtx)
		if err != nil {
			return err
		}
	}

	h.config = h.bootstrappedCtx[bootstrap.BootstrappedConfig].(config.Configuration)
	idBytes, err := h.config.GetIdentityID()
	if err != nil {
		return err
	}
	h.identity, err = identity.NewDIDFromBytes(idBytes)
	if err != nil {
		return err
	}
	h.idFactory = h.bootstrappedCtx[identity.BootstrappedDIDFactory].(identity.Factory)
	h.idService = h.bootstrappedCtx[identity.BootstrappedDIDService].(identity.Service)
	h.p2pClient = h.bootstrappedCtx[bootstrap.BootstrappedPeer].(documents.Client)
	h.configService = h.bootstrappedCtx[config.BootstrappedConfigStorage].(config.Service)
	h.tokenRegistry = h.bootstrappedCtx[bootstrap.BootstrappedNFTService].(documents.TokenRegistry)
	h.anchorSrv = h.bootstrappedCtx[anchors.BootstrappedAnchorService].(anchors.Service)
	h.entityService = h.bootstrappedCtx[entity.BootstrappedEntityService].(entity.Service)
	centAPI := h.bootstrappedCtx[centchain.BootstrappedCentChainClient].(centchain.API)
	h.nftAPI = nft.NewAPI(centAPI)
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
		log.Errorf("%s encountered error %v", h.name, err)
		return err
	case sig := <-controlC:
		log.Errorf("%s shutting down because of %s", h.name, sig.String())
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
			res, err := c.Get(fmt.Sprintf("http://localhost:%d/ping", h.config.GetServerPort()))
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

func (h *host) transferNFT(e *httpexpect.Expect, auth string, status int, params map[string]interface{}) (*httpexpect.Object, error) {
	return transferNFT(e, auth, status, params), nil
}

func (h *host) ownerOfNFT(e *httpexpect.Expect, auth string, status int, params map[string]interface{}) (*httpexpect.Value, error) {
	return ownerOfNFT(e, auth, status, params), nil
}

func (h *host) createAccounts(maeve *webhookReceiver, e *httpexpect.Expect) error {
	if !h.multiAccount {
		return nil
	}
	// create 3 accounts
	cacc := map[string]map[string]string{
		"centrifuge_chain_account": {
			"id":            h.centChainID,
			"secret":        h.centChainSecret,
			"ss_58_address": h.centChainAddress,
		},
	}

	for i := 0; i < 3; i++ {
		log.Infof("creating account %d for host %s", i, h.name)
		did, err := generateAccount(maeve, e, h.identity.String(), http.StatusCreated, cacc)
		if err != nil {
			return err
		}
		log.Infof("created account %d for host %s: %s", i, h.name, did)
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

func (h *host) createHTTPExpectation(t *testing.T) *httpexpect.Expect {
	return createInsecureClientWithExpect(t, fmt.Sprintf("http://localhost:%d", h.config.GetServerPort()))
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
