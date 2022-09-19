//go:build testworld

package testworld

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/centrifuge/go-centrifuge/pallets"
	"net/http"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"

	"github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"

	keystoreTypes "github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyTypes "github.com/centrifuge/chain-custom-types/pkg/proxy"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/gavv/httpexpect"
	logging "github.com/ipfs/go-log"
)

//var log = logging.Logger("host")

//var hostConfig = []struct {
//	name         string
//	multiAccount bool
//}{
//	{"Alice", false},
//	{"Bob", true},
//	{"Charlie", true},
//	{"Kenny", false},
//	{"Eve", false},
//	// Mallory has a mock document.Serivce to facilitate some Byzantine test
//	{"Mallory", false},
//}
//
//const defaultP2PTimeout = "10s"

// hostTestSuite encapsulates test utilities on top of each host
type hostTestSuite struct {
	testAccount *testAccount
	httpExpect  *httpexpect.Expect
}

// hostManager is the hostManager of the hosts at Testworld (Robert)
type hostManager struct {
	log *logging.ZapEventLogger
	// bernard is the bootnode for all the hosts
	//bernard *host

	// maeve is the webhook receiver for all hosts
	maeve *webhookReceiver

	testAccountMap map[testAccountName]*testAccount

	// canc is the cancel signal for all hosts
	rootCtxCanc context.CancelFunc

	// currently needed to restart a node
	// parent context
	rootCtx context.Context

	config config.Configuration

	serviceContext map[string]any

	podOperator *signerAccount
	nodeAdmin   *signerAccount
}

func newHostManager(
	config config.Configuration,
	serviceContext map[string]any,
	testAccountMap map[testAccountName]*testAccount,
) (*hostManager, error) {
	podOperator, err := getSignerAccount(config.GetPodOperatorSecretSeed())

	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator signer account: %w", err)
	}

	nodeAdmin, err := getSignerAccount(config.GetPodAdminSecretSeed())

	if err != nil {
		return nil, fmt.Errorf("couldn't get node admin signer account: %w", err)
	}

	return &hostManager{
		log:            logging.Logger("testworld-host-manager"),
		config:         config,
		serviceContext: serviceContext,
		testAccountMap: testAccountMap,
		podOperator:    podOperator,
		nodeAdmin:      nodeAdmin,
	}, nil
}

type signerAccount struct {
	AccountID  *types.AccountID
	Address    string
	SecretSeed string
}

func getSignerAccount(secretSeed string) (*signerAccount, error) {
	kp, err := subkey.DeriveKeyPair(sr25519.Scheme{}, secretSeed)

	if err != nil {
		return nil, fmt.Errorf("couldn't derive signer account key pair: %w", err)
	}

	accountID, err := types.NewAccountID(kp.AccountID())

	if err != nil {
		return nil, fmt.Errorf("couldn't create signer account ID: %w", err)
	}

	return &signerAccount{
		AccountID:  accountID,
		Address:    kp.SS58Address(auth.CentrifugeNetworkID),
		SecretSeed: secretSeed,
	}, nil
}

//func (r *hostManager) reLive(t *testing.T, name string) {
//	r.startHost(name)
//	// wait for the host to be live, here its 11 seconds allowed but the host should come alive before that and this will return faster
//	ok, err := r.getTestAccount(name).isLive(11 * time.Second)
//	if ok {
//		return
//	}
//	t.Error(err)
//}
//
//func (r *hostManager) startHost(name string) {
//	go r.niceHosts[name].live(r.rootCtx)
//}

const (
	nodeStartupErrorTimeout = 10 * time.Second
)

func (r *hostManager) init() error {
	r.rootCtx, r.rootCtxCanc = context.WithCancel(context.Background())

	// start listening to webhooks
	_, port, err := utils.GetFreeAddrPort()
	if err != nil {
		return fmt.Errorf("couldn't get free port: %w", err)
	}

	r.maeve = newWebhookReceiver(port, "/webhook")
	go r.maeve.start(r.rootCtx)

	nodeServers, err := node.GetServers(r.serviceContext)

	if err != nil {
		panic(fmt.Errorf("couldn't get node servers: %w", err))
	}

	node := node.New(nodeServers)

	errChan := make(chan error)

	nodeCtx := context.WithValue(r.rootCtx, bootstrap.NodeObjRegistry, r.serviceContext)

	go node.Start(nodeCtx, errChan)

	select {
	case err := <-errChan:
		panic(fmt.Errorf("couldn't start node: %w", err))
	case <-time.After(nodeStartupErrorTimeout):
		r.log.Debug("Node started successfully")
	}

	err = r.processTestAccounts()
	if err != nil {
		return fmt.Errorf("couldn't create accounts: %w", err)
	}

	//_, apiPort, err := utils.GetFreeAddrPort()
	//if err != nil {
	//	return err
	//}
	//
	//_, p2pPort, err := utils.GetFreeAddrPort()
	//if err != nil {
	//	return err
	//}

	//r.bernard = r.createHost(
	//	"Bernard",
	//	"",
	//	defaultP2PTimeout,
	//	apiPort,
	//	p2pPort,
	//	r.config.CreateHostConfigs,
	//	false,
	//	nil,
	//)
	//
	//if err = r.bernard.init(); err != nil {
	//	return err
	//}
	//
	//// start and wait for Bernard since other hosts depend on him
	//go r.bernard.live(r.cancCtx)
	//_, err = r.bernard.isLive(10 * time.Second)
	//if err != nil {
	//	return errors.New("bernard couldn't be made alive %v", err)
	//}
	//
	//bootnode, err := r.bernard.p2pURL()
	//if err != nil {
	//	return err
	//}

	//// start hosts
	//for _, h := range hostConfig {
	//	m := r.maeve.url()
	//
	//	_, apiPort, err := utils.GetFreeAddrPort()
	//	if err != nil {
	//		return fmt.Errorf("failed to get free port for api: %w", err)
	//	}
	//
	//	_, p2pPort, err := utils.GetFreeAddrPort()
	//	if err != nil {
	//		return fmt.Errorf("failed to get free port for p2p: %w", err)
	//	}
	//
	//	r.niceHosts[h.name] = r.createHost(
	//		h.name,
	//		m,
	//		defaultP2PTimeout,
	//		apiPort,
	//		p2pPort,
	//		r.config.CreateHostConfigs,
	//		h.multiAccount,
	//		[]string{bootnode},
	//	)
	//
	//	err = r.niceHosts[h.name].init()
	//	if err != nil {
	//		return err
	//	}
	//
	//	r.startHost(h.name)
	//}
	// make sure hosts are alive and print host centIDs
	//for name, host := range r.niceHosts {
	//	// Temporary until we have a proper healthcheck in place
	//	time.Sleep(2 * time.Second)
	//	_, err = host.isLive(10 * time.Second)
	//	if err != nil {
	//		return errors.New("%s couldn't be made alive %v", host.name, err)
	//	}
	//	fmt.Printf("Identity for %s is %s \n", name, host.identity.ToHexString())
	//	if r.config.CreateHostConfigs {
	//		err = host.processTestAccounts(r.maeve, r.getHostTestSuite(&testing.T{}, host.name).httpExpect)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//	err = host.loadAccounts(r.getHostTestSuite(&testing.T{}, host.name).httpExpect)
	//	if err != nil {
	//		return err
	//	}
	//
	//	dids := append(host.accounts, host.identity.ToHexString())
	//	for _, did := range dids {
	//		acc, err := host.configService.GetAccount(common.HexToAddress(did).Bytes())
	//		if err != nil {
	//			return err
	//		}
	//
	//		// hack to update the webhooks since the we pick a random port for webhook everytime
	//		a := acc.(*configstore.Account)
	//		a.WebhookURL = r.maeve.url()
	//		err = host.configService.UpdateAccount(a)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//}
	return nil
}

func (r *hostManager) getTestAccount(name testAccountName) *testAccount {
	if h, ok := r.testAccountMap[name]; ok {
		return h
	}

	return nil
}

func (r *hostManager) stop() {
	r.rootCtxCanc()
}

//func (r *hostManager) addNiceHost(name string, host *host) {
//	r.niceHosts[name] = host
//}

//func (r *hostManager) createTempHost(name, p2pTimeout string, apiPort, p2pPort int, createConfig, multiAccount bool, bootstraps []string) *host {
//	tempHost := r.createHost(name, r.maeve.url(), p2pTimeout, apiPort, p2pPort, createConfig, multiAccount, bootstraps)
//	r.tempHosts[name] = tempHost
//	return tempHost
//}

//func (r *hostManager) startTempHost(name string) error {
//	tempHost, ok := r.tempHosts[name]
//	if !ok {
//		return errors.New("host %s not found as temp host", name)
//	}
//	err := tempHost.init()
//	if err != nil {
//		return err
//	}
//	go tempHost.live(r.rootCtx)
//
//	return nil
//}

//func (r *hostManager) createHost(name, webhookURL string, p2pTimeout string, apiPort, p2pPort int, createConfig, multiAccount bool, bootstraps []string) *host {
//	return &host{
//		name:           name,
//		network:        r.config.Network,
//		apiHost:        "0.0.0.0",
//		apiPort:        apiPort,
//		p2pPort:        p2pPort,
//		p2pTimeout:     p2pTimeout,
//		bootstrapNodes: bootstraps,
//		dir:            fmt.Sprintf("hostconfigs/%s/%s", r.config.Network, name),
//		createConfig:   createConfig,
//		multiAccount:   multiAccount,
//		centChainURL:   r.config.CentChainURL,
//	}
//}

func (r *hostManager) processTestAccounts() error {
	ctx := context.Background()

	proxyAPI, ok := r.serviceContext[pallets.BootstrappedProxyAPI].(proxy.API)

	if !ok {
		return errors.New("proxy API not initialised")
	}

	keystoreAPI, ok := r.serviceContext[pallets.BootstrappedKeystoreAPI].(keystore.API)

	if !ok {
		return errors.New("keystore API not initialised")
	}

	cfgService, ok := r.serviceContext[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return errors.New("config service not initialised")
	}

	for testAccountName, testAccount := range r.testAccountMap {
		r.log.Infof("Creating account for - %s", testAccountName)

		accountID, err := testAccount.AccountID()

		if err != nil {
			return fmt.Errorf("couldn't get account ID: %w", err)
		}

		adminToken, err := r.getNodeAdminToken()

		if err != nil {
			return fmt.Errorf("couldn't create mock JW3T")
		}

		createAccountReq := coreapi.GenerateAccountPayload{
			Account: coreapi.Account{
				Identity:         accountID,
				WebhookURL:       r.maeve.url(),
				PrecommitEnabled: false,
			},
		}

		mr, err := json.Marshal(createAccountReq)

		if err != nil {
			return fmt.Errorf("couldn't create request payload")
		}

		var payload map[string]any

		if err := json.Unmarshal(mr, &payload); err != nil {
			return fmt.Errorf("couldn't unmarshal payload: %w", err)
		}

		expectCfg := httpexpect.Config{
			BaseURL:  fmt.Sprintf("http://%s", r.config.GetServerAddress()),
			Client:   createInsecureClient(),
			Reporter: &panicReporter{log: logging.Logger("testworld-init")},
		}

		expect := httpexpect.WithConfig(expectCfg)

		httpAcc, err := generateAccount(expect, adminToken, http.StatusCreated, payload)

		if err != nil {
			return fmt.Errorf("couldn't create account: %w", err)
		}

		proxy, err := generateProxyAccount()

		if err != nil {
			return fmt.Errorf("couldn't generate proxy account: %w", err)
		}

		testAccount.proxy = proxy

		if err := proxyAPI.AddProxy(ctx, proxy.AccountID, proxyTypes.PodAuth, 0, testAccount.keyRing); err != nil {
			return fmt.Errorf("couldn't pod auth proxy: %w", err)
		}

		if err := proxyAPI.AddProxy(ctx, httpAcc.PodOperatorAccountID, proxyTypes.PodOperation, 0, testAccount.keyRing); err != nil {
			return fmt.Errorf("couldn't add pod operator as pod operation proxy: %w", err)
		}

		if err := proxyAPI.AddProxy(ctx, httpAcc.PodOperatorAccountID, proxyTypes.KeystoreManagement, 0, testAccount.keyRing); err != nil {
			return fmt.Errorf("couldn't add pod operator as keystore management proxy: %w", err)
		}

		keys := []*keystoreTypes.AddKey{
			{
				Key:     types.NewHash(httpAcc.DocumentSigningPublicKey.Bytes()),
				Purpose: keystoreTypes.KeyPurposeP2PDocumentSigning,
				KeyType: keystoreTypes.KeyTypeECDSA,
			},
			{
				Key:     types.NewHash(httpAcc.P2PPublicSigningKey.Bytes()),
				Purpose: keystoreTypes.KeyPurposeP2PDocumentSigning,
				KeyType: keystoreTypes.KeyTypeECDSA,
			},
		}

		account, err := cfgService.GetAccount(accountID.ToBytes())

		if err != nil {
			return fmt.Errorf("couldn't retrieve account from storage: %w", err)
		}

		ctx := contextutil.WithAccount(ctx, account)

		if _, err := keystoreAPI.AddKeys(ctx, keys); err != nil {
			return fmt.Errorf("could add public keys to keystore: %w", err)
		}

		r.log.Infof("created account for host %s", testAccountName)
	}

	return nil
}

type panicReporter struct {
	log *logging.ZapEventLogger
}

func (p *panicReporter) Errorf(message string, args ...interface{}) {
	p.log.Errorf(message, args...)

	panic("encountered error")
}

func generateProxyAccount() (*signerAccount, error) {
	kp, err := sr25519.Scheme{}.Generate()

	if err != nil {
		return nil, fmt.Errorf("couldn't generate key pair: %w", err)
	}

	return getSignerAccount(hexutil.Encode(kp.Seed()))
}

func (r *hostManager) getNodeAdminToken() (string, error) {
	return auth.CreateJW3Token(
		r.nodeAdmin.AccountID,
		r.nodeAdmin.AccountID,
		r.nodeAdmin.SecretSeed,
		auth.PodAdminProxyType,
	)
}

func (r *hostManager) getHostTestSuite(t *testing.T, name testAccountName) hostTestSuite {
	testAccount := r.getTestAccount(name)
	expect := r.createHTTPExpectation(t)
	return hostTestSuite{testAccount: testAccount, httpExpect: expect}
}

func (r *hostManager) createHTTPExpectation(t *testing.T) *httpexpect.Expect {
	return createInsecureClientWithExpect(t, fmt.Sprintf("http://localhost:%d", r.config.GetServerPort()))
}

//func (h *host) mintNFT(e *httpexpect.Expect, auth string, status int, inv map[string]interface{}) (*httpexpect.Object, error) {
//	return mintNFT(e, auth, status, inv), nil
//}
//
//func (h *host) transferNFT(e *httpexpect.Expect, auth string, status int, params map[string]interface{}) (*httpexpect.Object, error) {
//	return transferNFT(e, auth, status, params), nil
//}
//
//func (h *host) ownerOfNFT(e *httpexpect.Expect, auth string, status int, params map[string]interface{}) (*httpexpect.Value, error) {
//	return ownerOfNFT(e, auth, status, params), nil
//}

//type host struct {
//	name, dir, network, apiHost,
//	anchorRepositoryAddr, p2pTimeout string
//	apiPort, p2pPort      int
//	bootstrapNodes        []string
//	bootstrappedCtx       map[string]interface{}
//	config                config.Configuration
//	identity              *types.AccountID
//	idService             v2.Service
//	node                  *node.Node
//	canc                  context.CancelFunc
//	createConfig          bool
//	multiAccount          bool
//	accounts              []string
//	p2pClient             documents.Client
//	configService         config.Service
//	anchorSrv             anchors.Service
//	entityService         entity.Service
//	centChainURL          string
//	authenticationEnabled bool
//	webhookURL            string
//}
//
//func (h *host) init() error {
//	if h.createConfig {
//		//err := cmd.CreateConfig(
//		//	h.dir,
//		//	h.network,
//		//	h.apiHost,
//		//	h.apiPort,
//		//	h.p2pPort,
//		//	h.bootstrapNodes,
//		//	h.p2pTimeout,
//		//	h.centChainURL,
//		//	h.authenticationEnabled,
//		//)
//		//
//		//if err != nil {
//		//	return err
//		//}
//	} else {
//		values := map[string]interface{}{"notifications.endpoint": h.webhookURL}
//		err := updateConfig(h.dir, values)
//		if err != nil {
//			return err
//		}
//	}
//
//	m := bootstrappers.MainBootstrapper{}
//	m.PopulateBaseBootstrappers()
//	h.bootstrappedCtx = map[string]interface{}{
//		config.BootstrappedConfigFile: h.dir + "/config.yaml",
//	}
//	err := m.Bootstrap(h.bootstrappedCtx)
//	if err != nil {
//		return err
//	}
//
//	//if h.name == "Mallory" {
//	//	// TODO(cdamian) change this
//	//	malloryDocMockSrv := new(documents.DocumentMock)
//	//	h.bootstrappedCtx["BootstrappedDocumentService"] = malloryDocMockSrv
//	//	p2pBoot := p2p.Bootstrapper{}
//	//	err := p2pBoot.Bootstrap(h.bootstrappedCtx)
//	//	if err != nil {
//	//		return err
//	//	}
//	//}
//
//	h.config = h.bootstrappedCtx[bootstrap.BootstrappedConfig].(config.Configuration)
//	h.p2pClient = h.bootstrappedCtx[bootstrap.BootstrappedPeer].(documents.Client)
//	h.configService = h.bootstrappedCtx[config.BootstrappedConfigStorage].(config.Service)
//	h.anchorSrv = h.bootstrappedCtx[anchors.BootstrappedAnchorService].(anchors.Service)
//	h.entityService = h.bootstrappedCtx[entity.BootstrappedEntityService].(entity.Service)
//
//	return nil
//}
//
//func (h *host) live(c context.Context) error {
//	srvs, err := node.GetServers(h.bootstrappedCtx)
//	if err != nil {
//		return errors.New("failed to load servers: %v", err)
//	}
//
//	h.node = node.New(srvs)
//	feedback := make(chan error)
//	// may be we can pass a context that exists in c here
//	cancCtx, canc := context.WithCancel(context.WithValue(c, bootstrap.NodeObjRegistry, h.bootstrappedCtx))
//
//	// cancel func of individual host
//	h.canc = canc
//
//	go h.node.Start(cancCtx, feedback)
//	controlC := make(chan os.Signal, 1)
//	signal.Notify(controlC, os.Interrupt)
//	select {
//	case err := <-feedback:
//		//log.Errorf("%s encountered error %v", h.name, err)
//		return err
//	case <-controlC:
//		//log.Errorf("%s shutting down because of %s", h.name, sig.String())
//		canc()
//		err := <-feedback
//		return err
//	}
//}
//
//func (h *host) kill() {
//	h.canc()
//}
//
//// isLive waits for host to come alive until the given soft timeout has passed, or the hard timeout of 10s is passed
//func (h *host) isLive(softTimeOut time.Duration) (bool, error) {
//	sig := make(chan error)
//	c := createInsecureClient()
//	go func(sig chan<- error) {
//		var fErr error
//		// wait upto 10 seconds(hard timeout) for the host to be live
//		for i := 0; i < 10; i++ {
//			res, err := c.Get(fmt.Sprintf("http://localhost:%d/ping", h.config.GetServerPort()))
//			fErr = err
//			if err != nil {
//				time.Sleep(time.Second)
//				continue
//			}
//			if res.StatusCode == http.StatusOK {
//				sig <- nil
//				return
//			}
//		}
//		sig <- fErr
//	}(sig)
//	t := time.After(softTimeOut)
//	select {
//	case <-t:
//		return false, errors.New("host failed to live even after %f seconds", softTimeOut.Seconds())
//	case err := <-sig:
//		if err != nil {
//			return false, err
//		}
//		return true, nil
//	}
//}
//
//func (h *host) processTestAccounts(maeve *webhookReceiver, e *httpexpect.Expect) error {
//	if !h.multiAccount {
//		return nil
//	}
//	//// create 3 accounts
//	//cacc := map[string]map[string]string{
//	//	"centrifuge_chain_account": {
//	//		"id":            h.centChainID,
//	//		"secret":        h.centChainSecret,
//	//		"ss_58_address": h.centChainAddress,
//	//	},
//	//}
//
//	// TODO(cdamian): Why the magic 3?
//	for i := 0; i < 3; i++ {
//		log.Infof("creating account %d for host %s", i, h.name)
//
//		identity, err := generateAccount(maeve, e, h.identity.String(), http.StatusCreated, cacc)
//		if err != nil {
//			return err
//		}
//
//		log.Infof("created account %d for host %s: %s", i, h.name, identity.ToHexString())
//	}
//	return nil
//}
//
//func (h *host) loadAccounts(e *httpexpect.Expect) error {
//	res := getAllAccounts(e, h.identity.String(), http.StatusOK)
//	accounts := res.Value("data").Array()
//	accIDs := getAccounts(accounts)
//	keys := make([]string, 0, len(accIDs))
//	for k := range accIDs {
//		keys = append(keys, k)
//	}
//	h.accounts = keys
//	return nil
//}
//
//func (h *host) p2pURL() (string, error) {
//	ctx := context.Background()
//	lastB58Key, err := h.idService.GetLastKeyByPurpose(ctx, h.identity, types.KeyPurposeP2PDiscovery)
//	if err != nil {
//		return "", err
//	}
//	peerID := peer.ID(lastB58Key[:])
//	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/ipfs/%s", h.p2pPort, peerID.Pretty()), nil
//}
