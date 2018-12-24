package nft

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/centerrors"
	ccommon "github.com/centrifuge/go-centrifuge/common"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
)

const (
	mintingConfirmationTaskName string = "MintingConfirmationTaskName"
	tokenIDParam                string = "TokenIDParam"
	registryAddressParam        string = "RegistryAddressParam"
	txIDParam                   string = "transactionIDparam"
)

// paymentObligationMintedFilterer filters the approved NFTs
type paymentObligationMintedFilterer interface {

	// FilterPaymentObligationMinted filters PaymentObligationMinted events
	FilterPaymentObligationMinted(opts *bind.FilterOpts) (*EthereumPaymentObligationContractPaymentObligationMintedIterator, error)
}

// mintingConfirmationTask confirms the minting of a payment obligation NFT
type mintingConfirmationTask struct {
	//task parameter
	TransactionID   uuid.UUID
	TokenID         string
	BlockHeight     uint64
	RegistryAddress string
	Timeout         time.Duration

	//state
	EthContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc)
	TxRepository          transactions.Repository
}

func newMintingConfirmationTask(
	timeout time.Duration,
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc),
	txRepo transactions.Repository,
) *mintingConfirmationTask {
	return &mintingConfirmationTask{
		Timeout:               timeout,
		EthContextInitializer: ethContextInitializer,
		TxRepository:          txRepo,
	}
}

// TaskTypeName returns mintingConfirmationTaskName
func (nftc *mintingConfirmationTask) TaskTypeName() string {
	return mintingConfirmationTaskName
}

// Copy returns a new instance of mintingConfirmationTask
func (nftc *mintingConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &mintingConfirmationTask{
		nftc.TransactionID,
		nftc.TokenID,
		nftc.BlockHeight,
		nftc.RegistryAddress,
		nftc.Timeout,
		nftc.EthContextInitializer,
		nftc.TxRepository,
	}, nil
}

// ParseKwargs - define a method to parse CentID
func (nftc *mintingConfirmationTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	txID, ok := kwargs[txIDParam].(uuid.UUID)
	if !ok {
		return errors.New("malformed transactionID: %v", kwargs[txIDParam])
	}
	nftc.TransactionID = txID

	// parse TokenID
	tokenID, ok := kwargs[tokenIDParam]
	if !ok {
		return errors.New("undefined kwarg " + tokenIDParam)
	}
	nftc.TokenID, ok = tokenID.(string)
	if !ok {
		return errors.New("malformed kwarg [%s]", tokenIDParam)
	}

	// parse BlockHeight
	nftc.BlockHeight, err = queue.ParseBlockHeight(kwargs)
	if err != nil {
		return err
	}

	//parse RegistryAddress
	registryAddress, ok := kwargs[registryAddressParam]
	if !ok {
		return errors.New("undefined kwarg " + registryAddressParam)
	}

	nftc.RegistryAddress, ok = registryAddress.(string)
	if !ok {
		return errors.New("malformed kwarg [%s]", registryAddressParam)
	}

	// override TimeoutParam if provided
	tdRaw, ok := kwargs[queue.TimeoutParam]
	if ok {
		td, err := queue.GetDuration(tdRaw)
		if err != nil {
			return errors.New("malformed kwarg [%s] because [%s]", queue.TimeoutParam, err.Error())
		}
		nftc.Timeout = td
	}

	return nil
}

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (nftc *mintingConfirmationTask) RunTask() (resp interface{}, err error) {
	log.Infof("Waiting for confirmation for the minting of token [%x]", nftc.TokenID)

	ethContext, cancelF := nftc.EthContextInitializer(nftc.Timeout)
	defer cancelF()
	fOpts := &bind.FilterOpts{
		Context: ethContext,
		Start:   nftc.BlockHeight,
	}

	defer func() {
		tx, erri := nftc.TxRepository.Get(ccommon.DummyIdentity, nftc.TransactionID)
		if erri != nil {
			log.Infof("failed to fetch transaction: %v", erri)
			return
		}

		var msg string
		tx.Status = transactions.Success
		if err != nil {
			msg = fmt.Sprintf("failed to mint NFT: %v", err)
			tx.Status = transactions.Failed
		}

		txLog := transactions.NewLog(mintingConfirmationTaskName, msg)
		tx.Logs = append(tx.Logs, txLog)
		if erri = nftc.TxRepository.Save(tx); erri != nil {
			log.Infof("failed to save transaction: %v", erri)
		}
	}()

	var filter paymentObligationMintedFilterer
	filter, err = bindContract(common.HexToAddress(nftc.RegistryAddress), ethereum.GetClient())
	if err != nil {
		return nil, err
	}

	for {
		iter, err := filter.FilterPaymentObligationMinted(fOpts)
		if err != nil {
			return nil, centerrors.Wrap(err, "failed to start filtering token minted logs")
		}

		err = utils.LookForEvent(iter)
		if err == nil {
			log.Infof("Received filtered event NFT minted for token [%s] \n", nftc.TokenID)
			return iter.Event, nil
		}

		if err != utils.ErrEventNotFound {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}
}
