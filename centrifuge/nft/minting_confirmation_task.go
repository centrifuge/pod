package nft

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

const (
	MintingConfirmationTaskName string = "MintingConfirmationTaskName"
	TokenIDParam                string = "TokenIDParam"
	BlockHeight                 string = "BlockHeight"
)

// PaymentObligationMintedFilterer filters the approved NFTs
type PaymentObligationMintedFilterer interface {

	// FilterPaymentObligationMinted filters PaymentObligationMinted events
	FilterPaymentObligationMinted(opts *bind.FilterOpts) (*EthereumPaymentObligationContractPaymentObligationMintedIterator, error)
}

// MintingConfirmationTask confirms the minting of a payment obligation NFT
type MintingConfirmationTask struct {
	TokenID                         string
	BlockHeight                     uint64
	EthContextInitializer           func() (ctx context.Context, cancelFunc context.CancelFunc)
	EthContext                      context.Context
	PaymentObligationMintedFilterer PaymentObligationMintedFilterer
	Config                          Config
}

func NewMintingConfirmationTask(
	nftApprovedFilterer PaymentObligationMintedFilterer,
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc),
) *MintingConfirmationTask {
	return &MintingConfirmationTask{
		PaymentObligationMintedFilterer: nftApprovedFilterer,
		EthContextInitializer:           ethContextInitializer,
	}
}

func (nftc *MintingConfirmationTask) Name() string {
	return MintingConfirmationTaskName
}

func (nftc *MintingConfirmationTask) Init() error {
	queue.Queue.Register(MintingConfirmationTaskName, nftc)
	return nil
}

// ParseKwargs - define a method to parse CentID
func (nftc *MintingConfirmationTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	tokenID, ok := kwargs[TokenIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + TokenIDParam)
	}
	nftc.TokenID, ok = tokenID.(string)
	if !ok {
		return fmt.Errorf("malformed kwarg [%s]", TokenIDParam)
	}

	nftc.BlockHeight, err = parseBlockHeight(kwargs)
	if err != nil {
		return err
	}
	return nil
}

func parseBlockHeight(valMap map[string]interface{}) (uint64, error) {
	if bhi, ok := valMap[BlockHeight]; ok {
		bhf, ok := bhi.(float64)
		if ok {
			return uint64(bhf), nil
		} else {
			return 0, errors.New("value can not be parsed")
		}
	}
	return 0, errors.New("value can not be parsed")
}

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (nftc *MintingConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the minting of token [%x]", nftc.TokenID)
	if nftc.EthContext == nil {
		nftc.EthContext, _ = nftc.EthContextInitializer()
	}

	fOpts := &bind.FilterOpts{
		Context: nftc.EthContext,
		Start:   nftc.BlockHeight,
	}

	for {
		iter, err := nftc.PaymentObligationMintedFilterer.FilterPaymentObligationMinted(
			fOpts,
		)
		if err != nil {
			return nil, centerrors.Wrap(err, "failed to start filtering token minted logs")
		}

		err = utils.LookForEvent(iter)
		if err == nil && iter.Event.TokenId.String() == nftc.TokenID {
			log.Infof("Received filtered event NFT minted for token [%s] \n", nftc.TokenID)
			return iter.Event, nil
		}

		if err != utils.EventNotFound {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to filter nft minted events")
}
