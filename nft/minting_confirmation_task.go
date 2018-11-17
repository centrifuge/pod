package nft

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/gocelery"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

const (
	mintingConfirmationTaskName string = "MintingConfirmationTaskName"
	tokenIDParam                string = "TokenIDParam"
)

// paymentObligationMintedFilterer filters the approved NFTs
type paymentObligationMintedFilterer interface {

	// FilterPaymentObligationMinted filters PaymentObligationMinted events
	FilterPaymentObligationMinted(opts *bind.FilterOpts) (*EthereumPaymentObligationContractPaymentObligationMintedIterator, error)
}

// mintingConfirmationTask confirms the minting of a payment obligation NFT
type mintingConfirmationTask struct {
	TokenID                         string
	BlockHeight                     uint64
	Timeout                         time.Duration
	EthContextInitializer           func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc)
	EthContext                      context.Context
	PaymentObligationMintedFilterer paymentObligationMintedFilterer
}

func newMintingConfirmationTask(
	timeout time.Duration,
	nftApprovedFilterer paymentObligationMintedFilterer,
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc),
) *mintingConfirmationTask {
	return &mintingConfirmationTask{
		Timeout:                         timeout,
		PaymentObligationMintedFilterer: nftApprovedFilterer,
		EthContextInitializer:           ethContextInitializer,
	}
}

// Name returns mintingConfirmationTaskName
func (nftc *mintingConfirmationTask) Name() string {
	return mintingConfirmationTaskName
}

// Init registers the task to the queue
func (nftc *mintingConfirmationTask) Init() error {
	queue.Queue.Register(mintingConfirmationTaskName, nftc)
	return nil
}

// Copy returns a new instance of mintingConfirmationTask
func (nftc *mintingConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &mintingConfirmationTask{
		nftc.TokenID,
		nftc.BlockHeight,
		nftc.Timeout,
		nftc.EthContextInitializer,
		nftc.EthContext,
		nftc.PaymentObligationMintedFilterer,
	}, nil
}

// ParseKwargs - define a method to parse CentID
func (nftc *mintingConfirmationTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	tokenID, ok := kwargs[tokenIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + tokenIDParam)
	}
	nftc.TokenID, ok = tokenID.(string)
	if !ok {
		return fmt.Errorf("malformed kwarg [%s]", tokenIDParam)
	}

	nftc.BlockHeight, err = queue.ParseBlockHeight(kwargs)
	if err != nil {
		return err
	}

	tdRaw, ok := kwargs[queue.TimeoutParam]
	if ok {
		td, err := queue.GetDuration(tdRaw)
		if err != nil {
			return fmt.Errorf("malformed kwarg [%s] because [%s]", queue.TimeoutParam, err.Error())
		}
		nftc.Timeout = td
	}

	return nil
}

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (nftc *mintingConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the minting of token [%x]", nftc.TokenID)
	if nftc.EthContext == nil {
		nftc.EthContext, _ = nftc.EthContextInitializer(nftc.Timeout)
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
		if err == nil {
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
