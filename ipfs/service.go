package ipfs

import (
	"bytes"
	"context"
	"io/ioutil"

	logging "github.com/ipfs/go-log"
	icore "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
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

func New(api icore.CoreAPI) Service {
	log := logging.Logger("ipfs-service")

	return &service{
		api: api,
		log: log,
	}
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
