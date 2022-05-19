package ipfs

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrBlockRetrieval = errors.Error("couldn't retrieve block")
	ErrBlockRead      = errors.Error("couldn't read block")
	ErrBlockPut       = errors.Error("couldn't put block")
	ErrBlockStat      = errors.Error("couldn't stat block")
)
