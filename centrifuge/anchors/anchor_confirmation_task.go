package anchors

import (
	"context"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"
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

type AnchorCommittedWatcher interface {
	FilterAnchorCommitted(
		opts *bind.FilterOpts,
		from []common.Address,
		anchorId []*big.Int,
		centrifugeId []*big.Int) (*EthereumAnchorRepositoryContractAnchorCommittedIterator, error)
}

// AnchoringConfirmationTask is a queued task to watch ID registration events on Ethereum using EthereumAnchoryRepositoryContract.
// To see how it gets registered see bootstrapper.go and to see how it gets used see setUpRegistrationEventListener method
type AnchoringConfirmationTask struct {
	// task parameters
	From         common.Address
	AnchorID     AnchorID
	CentrifugeID identity.CentID
	BlockHeight  uint64

	// state
	EthContextInitializer   func() (ctx context.Context, cancelFunc context.CancelFunc)
	EthContext              context.Context
	AnchorCommittedFilterer AnchorCommittedWatcher
}

func NewAnchoringConfirmationTask(
	anchorCommittedWatcher AnchorCommittedWatcher,
	ethContextInitializer func() (ctx context.Context, cancelFunc context.CancelFunc),
) *AnchoringConfirmationTask {
	return &AnchoringConfirmationTask{
		AnchorCommittedFilterer: anchorCommittedWatcher,
		EthContextInitializer:   ethContextInitializer,
	}
}

func (act *AnchoringConfirmationTask) Name() string {
	return AnchorRepositoryConfirmationTaskName
}

func (act *AnchoringConfirmationTask) Init() error {
	queue.Queue.Register(act.Name(), act)
	return nil
}

func (act *AnchoringConfirmationTask) Copy() (gocelery.CeleryTask, error) {
	return &AnchoringConfirmationTask{
		act.From,
		act.AnchorID,
		act.CentrifugeID,
		act.BlockHeight,
		act.EthContextInitializer,
		act.EthContext,
		act.AnchorCommittedFilterer,
	}, nil
}

// ParseKwargs - define a method to parse AnchorID, Address and RootHash
func (act *AnchoringConfirmationTask) ParseKwargs(kwargs map[string]interface{}) error {
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
	centrifugeId, ok := kwargs[CentrifugeIDParam]
	if !ok {
		return fmt.Errorf("undefined kwarg " + CentrifugeIDParam)
	}

	centrifugeIdBytes, err := getBytesCentrifugeID(centrifugeId)
	if err != nil {
		return fmt.Errorf("malformed kwarg [%s] because [%s]", CentrifugeIDParam, err.Error())
	}

	act.CentrifugeID = centrifugeIdBytes

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

// RunTask calls listens to events from geth related to AnchoringConfirmationTask#AnchorID and records result.
func (act *AnchoringConfirmationTask) RunTask() (interface{}, error) {
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
			[]*big.Int{act.AnchorID.toBigInt()},
			[]*big.Int{act.CentrifugeID.BigInt()},
		)
		if err != nil {
			return nil, centerrors.Wrap(err, "failed to start filtering anchor event logs")
		}

		err = utils.LookForEvent(iter)
		if err == nil {
			return iter.Event, nil
		}

		if err != utils.EventNotFound {
			return nil, err
		}
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
	var fixed [identity.CentIDByteLength]byte
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
