package repository

import (
	logging "github.com/ipfs/go-log"
	"math/big"
)

var log = logging.Logger("anchorRepository")

type AnchorRepository interface {
	PreCommit(anchorId *big.Int, signingRoot [32]byte, centrifugeId *big.Int, signature []byte, expirationBlock *big.Int) (<-chan *WatchCommit, error)
	Commit(anchorId *big.Int, documentRoot [32]byte, centrifugeId *big.Int, documentProofs [][32]byte, signature []byte) (<-chan *WatchPreCommit, error)
}


func PreCommit(anchorId *big.Int, signingRoot [32]byte, centrifugeId *big.Int, signature []byte, expirationBlock *big.Int) (<-chan *WatchPreCommit, error) {
	anchorRepository,_ := getConfiguredRepository()

	confirmations, err := anchorRepository.PreCommit(anchorId, signingRoot,centrifugeId,signature,expirationBlock)
	if err != nil {
		log.Errorf("Failed to pre-commit the anchor [id:%x, hash:%x ]: %v", anchorId, signingRoot, err)
	}
	return confirmations, err
}

func Commit(anchorId *big.Int, documentRoot [32]byte, centrifugeId *big.Int, documentProofs [][32]byte, signature []byte) (<-chan *WatchCommit, error) {
	anchorRepository,_ := getConfiguredRepository()

	confirmations, err := anchorRepository.Commit(anchorId, documentRoot, centrifugeId, documentProofs, signature)
	if err != nil {
		log.Errorf("Failed to commit the anchor [id:%x, hash:%x ]: %v", anchorId, documentRoot, err)
	}
	return confirmations, err
}



// getConfiguredRepository will later pull a configured repository (if not only using Ethereum as the anchor repository)
// For now hard-coded to the Ethereum setup
func getConfiguredRepository() (AnchorRepository, error) {
	return &EthereumAnchorRepository{}, nil
}
