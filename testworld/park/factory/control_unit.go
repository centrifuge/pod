//go:build testworld

package factory

import (
	"fmt"

	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/config/configstore"
	"github.com/centrifuge/pod/dispatcher"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/documents/entity"
	"github.com/centrifuge/pod/documents/entityrelationship"
	"github.com/centrifuge/pod/documents/generic"
	"github.com/centrifuge/pod/http"
	httpv2 "github.com/centrifuge/pod/http/v2"
	httpv3 "github.com/centrifuge/pod/http/v3"
	identityv2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/ipfs"
	"github.com/centrifuge/pod/jobs"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	"github.com/centrifuge/pod/p2p"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pending"
	"github.com/centrifuge/pod/storage/leveldb"
	p2pUtils "github.com/centrifuge/pod/testingutils/p2p"
	"github.com/centrifuge/pod/testworld/park/host"
)

func CreateHostControlUnit(bootstrapPeers []string) (*host.ControlUnit, error) {
	hostCfg, hostCfgFile, err := createHostConfig(bootstrapPeers)

	if err != nil {
		return nil, fmt.Errorf("couldn't create test host config: %w", err)
	}

	if err := config.GenerateAndWriteP2PKeys(hostCfg); err != nil {
		return nil, fmt.Errorf("couldn't generate and write P2P keys: %w", err)
	}

	hostServiceCtx := make(map[string]any)
	hostServiceCtx[config.BootstrappedConfigFile] = hostCfgFile

	p2pAddress, err := p2pUtils.GetLocalP2PAddress(hostCfg)

	if err != nil {
		return nil, fmt.Errorf("couldn't get local P2P address: %w", err)
	}

	return host.NewControlUnit(
		hostCfg,
		hostServiceCtx,
		getTestworldBootstrappers(),
		p2pAddress,
	), nil
}

func createHostConfig(bootstrapPeers []string) (config.Configuration, string, error) {
	hostCfg, hostCfgFile, err := config.CreateTestConfig(func(cfgArgs map[string]any) {
		if bootstrapPeers != nil {
			cfgArgs["bootstraps"] = bootstrapPeers
		}
	})

	if err != nil {
		return nil, "", fmt.Errorf("couldn't create host config: %w", err)
	}

	return hostCfg, hostCfgFile, nil
}

func getTestworldBootstrappers() []bootstrap.Bootstrapper {
	return []bootstrap.Bootstrapper{
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
		&ipfs.TestBootstrapper{},
		&nftv3.Bootstrapper{},
		&p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		&entity.Bootstrapper{},
		httpv2.Bootstrapper{},
		&httpv3.Bootstrapper{},
		&http.Bootstrapper{},
	}
}
