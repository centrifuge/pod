package ipfs

import (
	"bytes"
	"context"
	"io/ioutil"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/ipfs/interface-go-ipfs-core/options"

	logging "github.com/ipfs/go-log"
	"github.com/ipfs/interface-go-ipfs-core/path"

	icore "github.com/ipfs/interface-go-ipfs-core"
)

type Service interface {
	GetBlock(ctx context.Context, blockPath string) ([]byte, error)
	StatBlock(ctx context.Context, blockPath string) (icore.BlockStat, error)
	PutBlock(ctx context.Context, data []byte, opts ...options.BlockPutOption) (icore.BlockStat, error)
}

type service struct {
	api icore.CoreAPI
	log *logging.ZapEventLogger
}

func New(
	ctx context.Context,
	cfg config.Configuration,
) (Service, error) {
	log := logging.Logger("ipfs-node")

	if err := setupIPFSPlugins(cfg.GetIPFSPluginsPath()); err != nil {
		return nil, err
	}

	repoPath, err := createTempRepo()

	if err != nil {
		return nil, err
	}

	api, err := createAPI(ctx, repoPath)

	if err != nil {
		return nil, err
	}

	go func() {
		if err := bootstrapAPI(ctx, cfg, api); err != nil {
			log.Errorf("Couldn't bootstrap IPFS API - %s", err)
		} else {
			log.Info("IPFS bootstrap complete")
		}
	}()

	return &service{
		api: api,
		log: log,
	}, nil
}

func (s *service) GetBlock(ctx context.Context, blockPath string) ([]byte, error) {
	r, err := s.api.Block().Get(ctx, path.New(blockPath))

	if err != nil {
		s.log.Errorf("Couldn't retrieve block - %s", err)

		return nil, ErrBlockRetrieval
	}

	b, err := ioutil.ReadAll(r)

	if err != nil {
		s.log.Errorf("Couldn't read block - %s", err)

		return nil, ErrBlockRead
	}

	return b, nil
}

func (s *service) StatBlock(ctx context.Context, blockPath string) (icore.BlockStat, error) {
	blockStat, err := s.api.Block().Stat(ctx, path.New(blockPath))

	if err != nil {
		s.log.Errorf("Couldn't stat block - %s", err)

		return nil, ErrBlockStat
	}

	return blockStat, nil
}

func (s *service) PutBlock(ctx context.Context, data []byte, opts ...options.BlockPutOption) (icore.BlockStat, error) {
	blockStat, err := s.api.Block().Put(ctx, bytes.NewReader(data), opts...)

	if err != nil {
		s.log.Errorf("Couldn't put block - %s", err)

		return nil, ErrBlockPut
	}

	return blockStat, nil
}
