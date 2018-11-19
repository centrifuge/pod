package nft

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/ethereum"

	"github.com/ethereum/go-ethereum/common"

	"github.com/centrifuge/gocelery"

	"github.com/centrifuge/go-centrifuge/centerrors"
		"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

const (
	mintingConfirmationTaskName string = "MintingConfirmationTaskName"
	tokenIDParam                string = "TokenIDParam"
	blockHeightParam            string = "BlockHeightParam"
	registryAddressParam        string = "RegistryAddressParam"
)

// paymentObligationMintedFilterer filters the approved NFTs
type paymentObligationMintedFilterer interface {

	// FilterPaymentObligationMinted filters PaymentObligationMinted events
	FilterPaymentObligationMinted(opts *bind.FilterOpts) (*EthereumPaymentObligationContractPaymentObligationMintedIterator, error)
}

// mintingConfirmationTask confirms the minting of a payment obligation NFT
type mintingConfirmationTask struct {
	//task parameter
	TokenID         string
	BlockHeight     uint64
	RegistryAddress string

	//state
	EthContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc)
}

func newMintingConfirmationTask(
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc),
) *mintingConfirmationTask {
	return &mintingConfirmationTask{

		EthContextInitializer: ethContextInitializer,
	}
}

// TaskTypeName returns mintingConfirmationTaskName
func (nftc *mintingConfirmationTask) TaskTypeName() string {
	return mintingConfirmationTaskName
}

// Copy returns a new instance of mintingConfirmationTask
func (nftc *mintingConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &mintingConfirmationTask{
		nftc.TokenID,
		nftc.BlockHeight,
		nftc.RegistryAddress,
		nftc.EthContextInitializer,
	}, nil
}

// ParseKwargs - define a method to parse CentID
func (nftc *mintingConfirmationTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	// parse TokenID
	tokenID, ok := kwargs[tokenIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + tokenIDParam)
	}
	nftc.TokenID, ok = tokenID.(string)
	if !ok {
		return fmt.Errorf("malformed kwarg [%s]", tokenIDParam)
	}

	// parse BlockHeight
	nftc.BlockHeight, err = parseBlockHeight(kwargs)
	if err != nil {
		return err
	}

	//parse RegistryAddress
	registryAddress, ok := kwargs[registryAddressParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + registryAddressParam)
	}

	nftc.RegistryAddress, ok = registryAddress.(string)
	if !ok {
		return fmt.Errorf("malformed kwarg [%s]", registryAddressParam)
	}

	return nil
}

func parseBlockHeight(valMap map[string]interface{}) (uint64, error) {
	if bhi, ok := valMap[blockHeightParam]; ok {
		bhf, ok := bhi.(float64)
		if ok {
			return uint64(bhf), nil
		}
	}
	return 0, errors.New("value can not be parsed")
}

// RunTask calls listens to events from geth related to MintingConfirmationTask#TokenID and records result.
func (nftc *mintingConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the minting of token [%x]", nftc.TokenID)

	ethContext, _ := nftc.EthContextInitializer()

	fOpts := &bind.FilterOpts{
		Context: ethContext,
		Start:   nftc.BlockHeight,
	}

	var filter paymentObligationMintedFilterer
	var err error

	filter, err = bindContract(common.HexToAddress(nftc.RegistryAddress), ethereum.GetClient())

	if err != nil {
		return nil, err
	}

	for {
		iter, err := filter.FilterPaymentObligationMinted(
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
