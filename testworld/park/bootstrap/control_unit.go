//go:build testworld

package bootstrap

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/bootstrap"
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
	p2pUtils "github.com/centrifuge/go-centrifuge/testingutils/p2p"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
)

func bootstrapHostControlUnit(
	bootstrapPeers *[]string,
	lastP2PPort *int,
	lastAPIPort *int,
) (*host.ControlUnit, error) {
	hostCfg, hostCfgFile, err := createHostConfig(*bootstrapPeers, *lastP2PPort, *lastAPIPort)

	if err != nil {
		return nil, fmt.Errorf("couldn't create test host config: %w", err)
	}

	hostServiceCtx := make(map[string]any)
	hostServiceCtx[config.BootstrappedConfigFile] = hostCfgFile

	testHostControlUnit := host.NewControlUnit(hostCfg, hostServiceCtx, getTestworldBootstrappers())

	if err := testHostControlUnit.Start(); err != nil {
		return nil, fmt.Errorf("couldn't bootstrap services: %w", err)
	}

	localP2PAddress, err := p2pUtils.GetLocalP2PAddress(hostCfg)

	if err != nil {
		return nil, fmt.Errorf("couldn't get local P2P address: %w", err)
	}

	*bootstrapPeers = append(*bootstrapPeers, localP2PAddress)
	*lastP2PPort = hostCfg.GetP2PPort()
	*lastAPIPort = hostCfg.GetServerPort()

	return testHostControlUnit, nil
}

func createHostConfig(
	bootstrapPeers []string,
	lastP2PPort int,
	lastAPIPort int,
) (config.Configuration, string, error) {
	hostCfg, hostCfgFile, err := config.CreateTestConfig(func(cfgArgs map[string]any) {
		if bootstrapPeers != nil {
			cfgArgs["bootstraps"] = bootstrapPeers
		}

		if lastP2PPort != 0 {
			cfgArgs["p2pPort"] = mustGetFreePort()
		}

		if lastAPIPort != 0 {
			cfgArgs["apiPort"] = mustGetFreePort()
		}
	})

	if err != nil {
		return nil, "", fmt.Errorf("couldn't create host config: %w", err)
	}

	return hostCfg, hostCfgFile, nil
}

func getTestworldBootstrappers() []bootstrap.TestBootstrapper {
	return []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		&jobs.Bootstrapper{},
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

func mustGetFreePort() int {
	_, port, err := utils.GetFreeAddrPort()

	if err != nil {
		panic("couldn't get free port")
	}

	return port
}
