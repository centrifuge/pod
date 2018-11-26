package anchors

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

const (
	anchorRepositoryConfirmationTaskName string = "anchorRepositoryConfirmationTaskName"
	anchorIDParam                        string = "anchorIDParam"
	centIDParam                          string = "centIDParam"
	blockHeight                          string = "blockHeight"
	addressParam                         string = "addressParam"
)

type anchorCommittedWatcher interface {
	FilterAnchorCommitted(
		opts *bind.FilterOpts,
		from []common.Address,
		anchorID []*big.Int,
		centID []*big.Int) (*EthereumAnchorRepositoryContractAnchorCommittedIterator, error)
}

// anchorConfirmationTask is a queued task to watch ID registration events on Ethereum using EthereumAnchoryRepositoryContract.
// To see how it gets registered see bootstrapper.go and to see how it gets used see setUpRegistrationEventListener method
type anchorConfirmationTask struct {
	// task parameters
	From         common.Address
	AnchorID     AnchorID
	CentrifugeID identity.CentID
	BlockHeight  uint64
	Timeout      time.Duration

	// state
	EthContextInitializer   func(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc)
	EthContext              context.Context
	AnchorCommittedFilterer anchorCommittedWatcher
}

// TaskTypeName returns anchorRepositoryConfirmationTaskName
func (act *anchorConfirmationTask) TaskTypeName() string {
	return anchorRepositoryConfirmationTaskName
}

// Copy returns a new instance of anchorConfirmationTask
func (act *anchorConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &anchorConfirmationTask{
		act.From,
		act.AnchorID,
		act.CentrifugeID,
		act.BlockHeight,
		act.Timeout,
		act.EthContextInitializer,
		act.EthContext,
		act.AnchorCommittedFilterer,
	}, nil
}

// ParseKwargs parses args to anchorConfirmationTask
func (act *anchorConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
	anchorID, ok := kwargs[anchorIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + anchorIDParam)
	}

	anchorIDBytes, err := getBytesAnchorID(anchorID)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", anchorIDParam, err.Error())
	}

	act.AnchorID = anchorIDBytes

	//parse the centrifuge id
	centID, ok := kwargs[centIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + centIDParam)
	}

	centIDBytes, err := getBytesCentrifugeID(centID)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", centIDParam, err.Error())
	}

	act.CentrifugeID = centIDBytes

	// parse the address
	address, ok := kwargs[addressParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + addressParam)
	}

	addressStr, ok := address.(string)
	if !ok {
		return fmt.Errorf("param is not hex string " + addressParam)
	}

	addressTyped, err := getAddressFromHexString(addressStr)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", addressParam, err.Error())
	}
	act.From = addressTyped

	if bhi, ok := kwargs[blockHeight]; ok {
		bhf, ok := bhi.(float64)
		if ok {
			act.BlockHeight = uint64(bhf)
		}
	}

	// Override default timeout param
	tdRaw, ok := kwargs[queue.TimeoutParam]
	if ok {
		td, err := queue.GetDuration(tdRaw)
		if err != nil {
			return fmt.Errorf("malformed kwarg [%s] because [%s]", queue.TimeoutParam, err.Error())
		}
		act.Timeout = td
	}

	return nil
}

// RunTask calls listens to events from geth related to anchorConfirmationTask#AnchorID and records result.
func (act *anchorConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the anchorID [%x]", act.AnchorID)
	if act.EthContext == nil {
		act.EthContext, _ = act.EthContextInitializer(act.Timeout)
	}

	fOpts := &bind.FilterOpts{
		Context: act.EthContext,
		Start:   act.BlockHeight,
	}

	for {
		iter, err := act.AnchorCommittedFilterer.FilterAnchorCommitted(
			fOpts,
			[]common.Address{act.From},
			[]*big.Int{act.AnchorID.BigInt()},
			[]*big.Int{act.CentrifugeID.BigInt()},
		)
		if err != nil {
			return nil, centerrors.Wrap(err, "failed to start filtering anchor event logs")
		}

		err = utils.LookForEvent(iter)
		if err == nil {
			log.Infof("Received filtered event Anchor Confirmation for AnchorID [%x] and CentrifugeID [%s]\n", act.AnchorID.BigInt(), act.CentrifugeID.String())
			return iter.Event, nil
		}

		if err != utils.ErrEventNotFound {
			return nil, err
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func getBytesAnchorID(key interface{}) (AnchorID, error) {
	var fixed [AnchorIDLength]byte
	b, ok := key.([]interface{})
	if !ok {
		return fixed, errors.New("Could not parse interface to []byte")
	}
	// convert and copy b byte values
	for i, v := range b {
		fv := v.(float64)
		fixed[i] = byte(fv)
	}
	return fixed, nil
}

func getBytesCentrifugeID(key interface{}) (identity.CentID, error) {
	var fixed [identity.CentIDLength]byte
	b, ok := key.([]interface{})
	if !ok {
		return fixed, errors.New("Could not parse interface to []byte")
	}
	// convert and copy b byte values
	for i, v := range b {
		fv := v.(float64)
		fixed[i] = byte(fv)
	}
	return fixed, nil
}

func getAddressFromHexString(hex string) (common.Address, error) {
	return common.BytesToAddress(common.FromHex(hex)), nil
}
