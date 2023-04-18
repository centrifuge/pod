package loans

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/validation"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("loans_api")
)

const (
	ErrClassIDEncoding      = errors.Error("class ID encoding")
	ErrActiveLoansRetrieval = errors.Error("active loans retrieval")
	ErrActiveLoansNotFound  = errors.Error("active loans not found")
)

const (
	PalletName             = "Loans"
	ActiveLoansStorageName = "ActiveLoans"
)

type API interface {
	GetActiveLoans(poolID types.U64) ([]ActiveLoanStorageEntry, error)
}

type api struct {
	centAPI centchain.API
}

func NewAPI(centAPI centchain.API) API {
	return &api{centAPI: centAPI}
}

func (a *api) GetActiveLoans(poolID types.U64) ([]ActiveLoanStorageEntry, error) {
	err := validation.Validate(
		validation.NewValidator(poolID, validation.U64ValidationFn),
	)

	if err != nil {
		log.Errorf("Validation error: %s", err)

		return nil, err
	}

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	encodedClassID, err := codec.Encode(poolID)

	if err != nil {
		log.Errorf("Couldn't encode class ID: %s", err)

		return nil, ErrClassIDEncoding
	}

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		ActiveLoansStorageName,
		encodedClassID,
	)

	if err != nil {
		log.Errorf("Couldn't create storage key: %s", err)

		return nil, errors.ErrStorageKeyCreation
	}

	var activeLoans []ActiveLoanStorageEntry

	ok, err := a.centAPI.GetStorageLatest(storageKey, &activeLoans)

	if err != nil {
		log.Errorf("Couldn't retrieve active loans from storage: %s", err)

		return nil, ErrActiveLoansRetrieval
	}

	if !ok {
		log.Error("Active loans not found for pool ID %d", poolID)

		return nil, ErrActiveLoansNotFound
	}

	return activeLoans, nil
}
