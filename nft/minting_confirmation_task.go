package nft

import (
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

const (
	mintingConfirmationTaskName string = "MintingConfirmationTaskName"
	tokenIDParam                string = "TokenIDParam"
	registryAddressParam        string = "RegistryAddressParam"
	tenantIDParam               string = "Tenant ID"
)

// paymentObligationMintedFilterer filters the approved NFTs
type paymentObligationMintedFilterer interface {

	// FilterPaymentObligationMinted filters PaymentObligationMinted events
	FilterPaymentObligationMinted(opts *bind.FilterOpts) (*EthereumPaymentObligationContractPaymentObligationMintedIterator, error)
}

// mintingConfirmationTask confirms the minting of a payment obligation NFT
type mintingConfirmationTask struct {
	transactions.BaseTask

	//task parameter
	tenantID        identity.CentID
	tokenID         string
	blockHeight     uint64
	registryAddress string
	timeout         time.Duration

	//state
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc)
}

func newMintingConfirmationTask(
	timeout time.Duration,
	ethContextInitializer func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc),
	txService transactions.Service,
) *mintingConfirmationTask {
	return &mintingConfirmationTask{
		timeout:               timeout,
		ethContextInitializer: ethContextInitializer,
		BaseTask:              transactions.BaseTask{TxService: txService},
	}
}

// TaskTypeName returns mintingConfirmationTaskName
func (nftc *mintingConfirmationTask) TaskTypeName() string {
	return mintingConfirmationTaskName
}

// Copy returns a new instance of mintingConfirmationTask
func (nftc *mintingConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &mintingConfirmationTask{
		timeout:               nftc.timeout,
		ethContextInitializer: nftc.ethContextInitializer,
		BaseTask:              transactions.BaseTask{TxService: nftc.TxService},
	}, nil
}

// ParseKwargs - define a method to parse CentID
func (nftc *mintingConfirmationTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	err = nftc.ParseTransactionID(kwargs)
	if err != nil {
		return err
	}

	tenantID, ok := kwargs[tenantIDParam].(string)
	if !ok {
		return errors.New("missing tenant ID")
	}

	nftc.tenantID, err = identity.CentIDFromString(tenantID)
	if err != nil {
		return err
	}

	// parse TokenID
	tokenID, ok := kwargs[tokenIDParam]
	if !ok {
		return errors.New("undefined kwarg " + tokenIDParam)
	}
	nftc.tokenID, ok = tokenID.(string)
	if !ok {
		return errors.New("malformed kwarg [%s]", tokenIDParam)
	}

	// parse blockHeight
	nftc.blockHeight, err = queue.ParseBlockHeight(kwargs)
	if err != nil {
		return err
	}

	//parse registryAddress
	registryAddress, ok := kwargs[registryAddressParam]
	if !ok {
		return errors.New("undefined kwarg " + registryAddressParam)
	}

	nftc.registryAddress, ok = registryAddress.(string)
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
		nftc.timeout = td
	}

	return nil
}

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (nftc *mintingConfirmationTask) RunTask() (resp interface{}, err error) {
	log.Infof("Waiting for confirmation for the minting of token [%x]", nftc.tokenID)

	ethContext, cancelF := nftc.ethContextInitializer(nftc.timeout)
	defer cancelF()
	fOpts := &bind.FilterOpts{
		Context: ethContext,
		Start:   nftc.blockHeight,
	}

	defer func() {
		if err != nil {
			log.Infof("failed to mint NFT: %v\n", err)
		} else {
			log.Infof("NFT minted successfully: %v\n", nftc.tokenID)
		}

		err = nftc.UpdateTransaction(nftc.tenantID, nftc.TaskTypeName(), err)
	}()

	var filter paymentObligationMintedFilterer
	filter, err = bindContract(common.HexToAddress(nftc.registryAddress), ethereum.GetClient())
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
			log.Infof("Received filtered event NFT minted for token [%s] \n", nftc.tokenID)
			return iter.Event, nil
		}

		if err != utils.ErrEventNotFound {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}
}
