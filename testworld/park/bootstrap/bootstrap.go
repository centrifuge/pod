//go:build testworld

package testworld

import (
	"context"
	"fmt"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/http"
	httpv2 "github.com/centrifuge/go-centrifuge/http/v2"
	httpv3 "github.com/centrifuge/go-centrifuge/http/v3"
	identityv2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	p2pUtils "github.com/centrifuge/go-centrifuge/testingutils/p2p"
	proxyUtils "github.com/centrifuge/go-centrifuge/testingutils/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("testworld")
)

var (
	testKeyringPairs = map[testHostName]signature.KeyringPair{
		testHostAlice:   keyrings.AliceKeyRingPair,
		testHostBob:     keyrings.BobKeyRingPair,
		testHostCharlie: keyrings.CharlieKeyRingPair,
		testHostDave:    keyrings.DaveKeyRingPair,
		// Eve/Ferdie are the PodAdmin/PodOperator of each host.
	}
)

func createTestHosts(webhookURL string) (map[testHostName]*testHost, error) {
	var (
		bootstrapPeers []string
		lastP2PPort    int
		lastAPIPort    int
	)

	testHosts := make(map[testHostName]*testHost)

	for hostName, keyringPair := range testKeyringPairs {
		log.Infof("Bootstrapping services for - %s", hostName)

		hostServiceCtx, err := bootstrapTestHostServices(&bootstrapPeers, &lastP2PPort, &lastAPIPort)

		if err != nil {
			return nil, fmt.Errorf("couldn't bootstrap test host services: %w", err)
		}

		log.Infof("Bootstrapping test host account for - %s", hostName)

		acc, podAuthProxy, err := bootstrapTestHostAccount(hostServiceCtx, keyringPair, webhookURL)

		if err != nil {
			return nil, fmt.Errorf("couldn't bootstrap test host account: %w", err)
		}

		testHosts[hostName] = &testHost{
			krp:          keyringPair,
			acc:          acc,
			podAuthProxy: podAuthProxy,
			serviceCtx:   hostServiceCtx,
		}
	}

	return testHosts, nil
}

func bootstrapTestHostServices(
	bootstrapPeers *[]string,
	lastP2PPort *int,
	lastAPIPort *int,
) (map[string]any, error) {
	hostCfg, hostCfgFile, err := createTestHostConfig(*bootstrapPeers, *lastP2PPort, *lastAPIPort)

	if err != nil {
		return nil, fmt.Errorf("couldn't create test host config: %w", err)
	}

	hostServiceCtx := make(map[string]any)
	hostServiceCtx[config.BootstrappedConfigFile] = hostCfgFile

	bootstrapper := newBootstrapper(getTestworldBootstrappers())

	if err := bootstrapper.Bootstrap(hostServiceCtx); err != nil {
		return nil, fmt.Errorf("couldn't bootstrap services: %w", err)
	}

	localP2PAddress, err := p2pUtils.GetLocalP2PAddress(hostCfg)

	if err != nil {
		return nil, fmt.Errorf("couldn't get local P2P address: %w", err)
	}

	*bootstrapPeers = append(*bootstrapPeers, localP2PAddress)
	*lastP2PPort = hostCfg.GetP2PPort()
	*lastAPIPort = hostCfg.GetServerPort()

	return hostServiceCtx, nil
}

func createTestHostConfig(
	bootstrapPeers []string,
	lastP2PPort int,
	lastAPIPort int,
) (config.Configuration, string, error) {
	hostCfg, hostCfgFile, err := config.CreateTestConfig(func(cfgArgs map[string]any) {
		if bootstrapPeers != nil {
			cfgArgs["bootstraps"] = bootstrapPeers
		}

		if lastP2PPort != 0 {
			cfgArgs["p2pPort"] = lastP2PPort + 1
		}

		if lastAPIPort != 0 {
			cfgArgs["apiPort"] = lastAPIPort + 1
		}
	})

	if err != nil {
		return nil, "", fmt.Errorf("couldn't create host config: %w", err)
	}

	return hostCfg, hostCfgFile, nil
}

type testworldBootstrapper struct {
	bootstrappers []bootstrap.TestBootstrapper
}

func (b *testworldBootstrapper) Bootstrap(serviceCtx map[string]any) error {
	for _, bootstrapper := range b.bootstrappers {
		if err := bootstrapper.TestBootstrap(serviceCtx); err != nil {
			return err
		}
	}

	return nil
}

func newBootstrapper(bootstrappers []bootstrap.TestBootstrapper) *testworldBootstrapper {
	return &testworldBootstrapper{bootstrappers}
}

func getTestworldBootstrappers() []bootstrap.TestBootstrapper {
	return []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		&jobs.Bootstrapper{},
		&integration_test.Bootstrapper{},
		centchain.Bootstrapper{},
		&pallets.Bootstrapper{},
		&dispatcher.Bootstrapper{},
		&identityv2.Bootstrapper{},
		documents.Bootstrapper{},
		&entityrelationship.Bootstrapper{},
		generic.Bootstrapper{},
		pending.Bootstrapper{},
		&ipfs_pinning.TestBootstrapper{},
		&nftv3.Bootstrapper{},
		&p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		&entity.Bootstrapper{},
		httpv2.Bootstrapper{},
		&httpv3.Bootstrapper{},
		&http.Bootstrapper{},
	}
}

func bootstrapTestHostAccount(
	serviceCtx map[string]any,
	krp signature.KeyringPair,
	webhookURL string,
) (config.Account, *signerAccount, error) {
	accountID, err := types.NewAccountID(krp.PublicKey)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	acc, err := identityv2.CreateTestAccount(serviceCtx, accountID, webhookURL)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	podAuthProxy, err := createTestHostAccountProxies(serviceCtx, krp)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't create test account proxies: %w", err)
	}

	if err := identityv2.AddAccountKeysToStore(serviceCtx, acc); err != nil {
		return nil, nil, fmt.Errorf("couldn't add test account keys to store: %w", err)
	}

	return acc, podAuthProxy, nil
}

func createTestHostAccountProxies(serviceCtx map[string]any, krp signature.KeyringPair) (*signerAccount, error) {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)

	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator: %w", err)
	}

	podAuthProxy, err := generateProxyAccount()

	if err != nil {
		return nil, fmt.Errorf("couldn't generate proxy account: %w", err)
	}

	proxyPairs := []identityv2.ProxyPair{
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.PodOperation,
		},
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.KeystoreManagement,
		},
		{
			Delegate:  podAuthProxy.AccountID,
			ProxyType: proxyType.PodAuth,
		},
	}

	if err := identityv2.AddTestProxies(serviceCtx, krp, proxyPairs...); err != nil {
		return nil, fmt.Errorf("couldn't add test proxies: %w", err)
	}

	delegatorAccountID, err := types.NewAccountID(krp.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	err = proxyUtils.WaitForProxiesToBeAdded(
		context.Background(),
		serviceCtx,
		delegatorAccountID,
		podOperator.GetAccountID(),
		podAuthProxy.AccountID,
	)

	if err != nil {
		return nil, fmt.Errorf("proxies were not added: %w", err)
	}

	return podAuthProxy, nil
}
