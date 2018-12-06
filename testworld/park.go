// +build testworld

package testworld

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"testing"

	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/cmd"
	"github.com/centrifuge/go-centrifuge/config"
	ctx "github.com/centrifuge/go-centrifuge/context"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/gavv/httpexpect"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("host")

var hostConfig = []struct {
	name             string
	apiPort, p2pPort int64
}{
	{"Alice", 8084, 38204},
	{"Bob", 8085, 38205},
	{"Charlie", 8086, 38206},
	{"Kenny", 8087, 38207},
}

// hostManager is the hostManager of the hosts at Testworld (Robert)
type hostManager struct {

	// network settings
	ethNodeUrl, accountKeyPath, accountPassword, network string

	txPoolAccess bool

	// contractAddresses are the addresses of centrifuge contracts on Ethereum
	contractAddresses *config.SmartContractAddresses

	// bernard is the bootnode for all the hosts
	bernard *host

	// niceHosts are the happy and nice hosts at the Testworld such as Teddy
	niceHosts map[string]*host

	// TODO create evil hosts such as William (or Eve)

	// canc is the cancel signal for all hosts
	canc context.CancelFunc

	// TODO: context should be removed from hostManager
	// currently needed to restart a node
	// parent context
	cancCtx context.Context
}

func newHostManager(
	ethNodeUrl, accountKeyPath, accountPassword, network string,
	txPoolAccess bool,
	smartContractAddrs *config.SmartContractAddresses) *hostManager {
	return &hostManager{
		ethNodeUrl:        ethNodeUrl,
		accountKeyPath:    accountKeyPath,
		accountPassword:   accountPassword,
		network:           network,
		txPoolAccess:      txPoolAccess,
		contractAddresses: smartContractAddrs,
		niceHosts:         make(map[string]*host),
	}
}

func (r *hostManager) restartHost(name string) {
	r.startHost(name)
	time.Sleep(time.Second * 1)
}

func (r *hostManager) startHost(name string) {
	go r.niceHosts[name].start(r.cancCtx)
}

func (r *hostManager) init() error {
	r.cancCtx, r.canc = context.WithCancel(context.Background())
	r.bernard = r.createHost("Bernard", 8081, 38201, nil)
	err := r.bernard.init()
	if err != nil {
		return err
	}

	// start and wait for Bernard since other hosts depend on him
	go r.bernard.start(r.cancCtx)
	time.Sleep(10 * time.Second)

	bootnode, err := r.bernard.p2pURL()
	if err != nil {
		return err
	}

	// start hosts
	for _, h := range hostConfig {
		r.niceHosts[h.name] = r.createHost(h.name, h.apiPort, h.p2pPort, []string{bootnode})

		err := r.niceHosts[h.name].init()
		if err != nil {
			return err
		}
		r.startHost(h.name)

	}
	// print host centIDs
	for name, host := range r.niceHosts {
		i, err := host.id()
		if err != nil {
			return err
		}
		fmt.Printf("CentID for %s is %s \n", name, i)
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

func (r *hostManager) createHost(name string, apiPort, p2pPort int64, bootstraps []string) *host {
	return newHost(
		name,
		r.ethNodeUrl,
		r.accountKeyPath,
		r.accountPassword,
		r.network,
		apiPort, p2pPort, bootstraps,
		r.txPoolAccess,
		r.contractAddresses,
	)
}

type hostTestSuite struct {
	name   string
	host   *host
	id     identity.CentID
	expect *httpexpect.Expect
}

func (r *hostManager) getHostTestSuite(t *testing.T, name string) hostTestSuite {
	host := r.getHost(name)
	expect := host.createHttpExpectation(t)
	id, err := host.id()
	if err != nil {
		t.Error(err)
	}
	return hostTestSuite{name: name, host: host, id: id, expect: expect}

}

type host struct {
	name, dir, ethNodeUrl, accountKeyPath, accountPassword, network,
	identityFactoryAddr, identityRegistryAddr, anchorRepositoryAddr, paymentObligationAddr string
	apiPort, p2pPort   int64
	bootstrapNodes     []string
	bootstrappedCtx    map[string]interface{}
	txPoolAccess       bool
	smartContractAddrs *config.SmartContractAddresses
	config             config.Config
	identity           identity.Identity
	node               *node.Node
	canc               context.CancelFunc
}

func newHost(
	name, ethNodeUrl, accountKeyPath, accountPassword, network string,
	apiPort, p2pPort int64,
	bootstraps []string,
	txPoolAccess bool,
	smartContractAddrs *config.SmartContractAddresses,
) *host {
	return &host{
		name:               name,
		ethNodeUrl:         ethNodeUrl,
		accountKeyPath:     accountKeyPath,
		accountPassword:    accountPassword,
		network:            network,
		apiPort:            apiPort,
		p2pPort:            p2pPort,
		bootstrapNodes:     bootstraps,
		txPoolAccess:       txPoolAccess,
		smartContractAddrs: smartContractAddrs,
		dir:                "peerconfigs/" + name,
	}
}

func (h *host) init() error {
	err := cmd.CreateConfig(h.dir, h.ethNodeUrl, h.accountKeyPath, h.accountPassword, h.network, h.apiPort, h.p2pPort, h.bootstrapNodes, h.txPoolAccess, h.smartContractAddrs)
	if err != nil {
		return err
	}
	m := ctx.MainBootstrapper{}
	m.PopulateBaseBootstrappers()
	h.bootstrappedCtx = map[string]interface{}{
		config.BootstrappedConfigFile: h.dir + "/config.yaml",
	}
	err = m.Bootstrap(h.bootstrappedCtx)
	if err != nil {
		return err
	}
	h.config = h.bootstrappedCtx[config.BootstrappedConfig].(config.Config)
	idService := h.bootstrappedCtx[identity.BootstrappedIDService].(identity.Service)
	idBytes, err := h.config.GetIdentityID()
	if err != nil {
		return err
	}
	id, err := identity.ToCentID(idBytes)
	if err != nil {
		return err
	}
	h.identity, err = idService.LookupIdentityForID(id)
	if err != nil {
		return err
	}
	return nil
}

func (h *host) start(c context.Context) error {
	srvs, err := node.GetServers(h.bootstrappedCtx)
	if err != nil {
		return fmt.Errorf("failed to load servers: %v", err)
	}

	h.node = node.New(srvs)
	feedback := make(chan error)
	// may be we can pass a context that exists in c here
	cancCtx, canc := context.WithCancel(context.WithValue(c, bootstrap.NodeObjRegistry, h.bootstrappedCtx))

	// cancel func of individual node
	h.canc = canc

	go h.node.Start(cancCtx, feedback)
	controlC := make(chan os.Signal, 1)
	signal.Notify(controlC, os.Interrupt)
	select {
	case err := <-feedback:
		log.Error(h.name+" encountered error ", err)
		return err
	case sig := <-controlC:
		log.Info(h.name+" shutting down because of ", sig)
		canc()
		err := <-feedback
		return err
	}

}

func (h *host) createInvoice(e *httpexpect.Expect, status int, inv map[string]interface{}) (*httpexpect.Object, error) {
	return createInvoice(e, status, inv), nil
}

func (h *host) updateInvoice(e *httpexpect.Expect, status int, docIdentifier string, inv map[string]interface{}) (*httpexpect.Object, error) {
	return updateInvoice(e, status, docIdentifier, inv), nil
}

func (h *host) createHttpExpectation(t *testing.T) *httpexpect.Expect {
	return createInsecureClient(t, fmt.Sprintf("https://localhost:%d", h.config.GetServerPort()))
}

func (h *host) id() (identity.CentID, error) {
	return h.identity.CentID(), nil
}

func (h *host) p2pURL() (string, error) {
	lastB58Key, err := h.identity.CurrentP2PKey()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/ipfs/%s", h.p2pPort, lastB58Key), nil
}
