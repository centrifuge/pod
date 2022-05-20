package ipfs

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/centrifuge/go-centrifuge/config"

	"golang.org/x/sync/errgroup"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"

	iconfig "github.com/ipfs/go-ipfs-config"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	icore "github.com/ipfs/interface-go-ipfs-core"
)

func setupIPFSPlugins(path string) error {
	plugins, err := loader.NewPluginLoader(path)

	if err != nil {
		return fmt.Errorf("couldn't load IPFS plugins: %w", err)
	}

	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("couldn't initialize IPFS plugins: %w", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("couldn't inject plugins: %w", err)
	}

	return nil
}

const (
	repoPattern = "ipfs-node"
)

func createTempRepo() (string, error) {
	repoPath, err := ioutil.TempDir("", repoPattern)

	if err != nil {
		return "", fmt.Errorf("couldn't create temporary dir for IPFS: %w", err)
	}

	// Create a config with default options and a 2048 bit key
	cfg, err := iconfig.Init(ioutil.Discard, 2048)
	if err != nil {
		return "", fmt.Errorf("couldn't initialize the IPFS config: %w", err)
	}

	// TODO(cdamian): Check these
	//// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#ipfs-filestore
	//cfg.Experimental.FilestoreEnabled = true
	//// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#ipfs-urlstore
	//cfg.Experimental.UrlstoreEnabled = true
	//// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#ipfs-p2p
	//cfg.Experimental.Libp2pStreamMounting = true
	//// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#p2p-http-proxy
	//cfg.Experimental.P2pHttpProxy = true
	//// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#strategic-providing
	//cfg.Experimental.StrategicProviding = true

	if err := fsrepo.Init(repoPath, cfg); err != nil {
		return "", fmt.Errorf("couldnt initialize the IPFS repo: %w", err)
	}

	return repoPath, nil
}

func createIPFSAPI(cfg config.Configuration) (icore.CoreAPI, error) {
	if err := setupIPFSPlugins(cfg.GetIPFSPluginsPath()); err != nil {
		return nil, err
	}

	repoPath, err := createTempRepo()

	if err != nil {
		return nil, err
	}

	repo, err := fsrepo.Open(repoPath)

	if err != nil {
		return nil, fmt.Errorf("couldn't open IPFS repo: %w", err)
	}

	nodeOptions := &core.BuildCfg{
		Online: true,
		//Routing: libp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo:    repo,
	}

	ctx := context.Background()

	node, err := core.NewNode(ctx, nodeOptions)

	if err != nil {
		return nil, fmt.Errorf("couldn't create new IPFS node: %w", err)
	}

	// TODO(cdamian): Check opts
	api, err := coreapi.NewCoreAPI(node)

	if err != nil {
		return nil, fmt.Errorf("couldn't create IPFS API: %w", err)
	}

	if err := bootstrapAPI(ctx, cfg, api); err != nil {
		return nil, err
	}

	return api, nil
}

func bootstrapAPI(ctx context.Context, cfg config.Configuration, api icore.CoreAPI) error {
	eg, egCtx := errgroup.WithContext(ctx)

	for _, addrStr := range cfg.GetIFPSBootstrapPeers() {
		eg.Go(func() error {
			addr, err := ma.NewMultiaddr(addrStr)

			if err != nil {
				return fmt.Errorf("couldn't create multi address: %w", err)
			}

			pii, err := peer.AddrInfoFromP2pAddr(addr)

			if err != nil {
				return fmt.Errorf("couldn't retrieve address info from p2p address: %w", err)
			}

			addrInfo := peer.AddrInfo{ID: pii.ID}

			if err := api.Swarm().Connect(egCtx, addrInfo); err != nil {
				return fmt.Errorf("couldn't connect to peer: %w", err)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("couldn't bootstrap IPFS node: %w", err)
	}

	return nil
}
