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
	AnchorRepositoryConfirmationTaskName string = "AnchorRepositoryConfirmationTaskName"
	AnchorIDParam                        string = "AnchorIDParam"
	CentrifugeIDParam                    string = "CentrifugeIDParam"
	BlockHeight                          string = "BlockHeight"
	AddressParam                         string = "AddressParam"
)

type anchorCommittedWatcher interface {
	FilterAnchorCommitted(
		opts *bind.FilterOpts,
		from []common.Address,
		anchorId []*big.Int,
		centrifugeId []*big.Int) (*EthereumAnchorRepositoryContractAnchorCommittedIterator, error)
}

// anchorConfirmationTask is a queued task to watch ID registration events on Ethereum using EthereumAnchoryRepositoryContract.
// To see how it gets registered see bootstrapper.go and to see how it gets used see setUpRegistrationEventListener method
type anchorConfirmationTask struct {
	// task parameters
	From         common.Address
	AnchorID     AnchorID
	CentrifugeID identity.CentID
	BlockHeight  uint64

	// state
	EthContextInitializer   func() (ctx context.Context, cancelFunc context.CancelFunc)
	EthContext              context.Context
	AnchorCommittedFilterer anchorCommittedWatcher
}

// Name returns AnchorRepositoryConfirmationTaskName
func (act *anchorConfirmationTask) Name() string {
	return AnchorRepositoryConfirmationTaskName
}

// Init registers the task to the queue
func (act *anchorConfirmationTask) Init() error {
	queue.Queue.Register(act.Name(), act)
	return nil
}

// Copy returns a new instance of anchorConfirmationTask
func (act *anchorConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &anchorConfirmationTask{
		act.From,
		act.AnchorID,
		act.CentrifugeID,
		act.BlockHeight,
		act.EthContextInitializer,
		act.EthContext,
		act.AnchorCommittedFilterer,
	}, nil
}

// ParseKwargs parses args to anchorConfirmationTask
func (act *anchorConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
	anchorID, ok := kwargs[AnchorIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + AnchorIDParam)
	}

	anchorIDBytes, err := getBytesAnchorID(anchorID)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", AnchorIDParam, err.Error())
	}

	act.AnchorID = anchorIDBytes

	//parse the centrifuge id
	centID, ok := kwargs[CentrifugeIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + CentrifugeIDParam)
	}

	centIDBytes, err := getBytesCentrifugeID(centID)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", CentrifugeIDParam, err.Error())
	}

	act.CentrifugeID = centIDBytes

	// parse the address
	address, ok := kwargs[AddressParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + AddressParam)
	}

	addressStr, ok := address.(string)
	if !ok {
		return fmt.Errorf("param is not hex string " + AddressParam)
	}

	addressTyped, err := getAddressFromHexString(addressStr)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", AddressParam, err.Error())
	}

	if bhi, ok := kwargs[BlockHeight]; ok {
		bhf, ok := bhi.(float64)
		if ok {
			act.BlockHeight = uint64(bhf)
		}
	}

	act.From = addressTyped
	return nil
}

// RunTask calls listens to events from geth related to anchorConfirmationTask#AnchorID and records result.
func (act *anchorConfirmationTask) RunTask() (interface{}, error) {
	log.Infof("Waiting for confirmation for the anchorID [%x]", act.AnchorID)
	if act.EthContext == nil {
		act.EthContext, _ = act.EthContextInitializer()
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
			log.Infof("Received filtered event Anchor Confirmation for AnchorID [%x] and CentrifugeId [%s]\n", act.AnchorID.BigInt(), act.CentrifugeID.String())
			return iter.Event, nil
		}

		if err != utils.EventNotFound {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to filter anchor events")
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
