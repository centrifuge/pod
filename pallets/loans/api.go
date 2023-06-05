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
	ErrPoolIDEncoding       = errors.Error("pool ID encoding")
	ErrLoanIDEncoding       = errors.Error("loan ID encoding")
	ErrActiveLoansRetrieval = errors.Error("active loans retrieval")
	ErrCreatedLoanRetrieval = errors.Error("created loans retrieval")
	ErrActiveLoansNotFound  = errors.Error("active loans not found")
	ErrCreatedLoanNotFound  = errors.Error("created loans not found")
)

const (
	PalletName             = "Loans"
	ActiveLoansStorageName = "ActiveLoans"
	CreatedLoanStorageName = "CreatedLoan"
)

//go:generate mockery --name API --structname APIMock --filename api_mock.go --inpackage

type API interface {
	GetCreatedLoan(poolID types.U64, loanID types.U64) (*CreatedLoanStorageEntry, error)
	GetActiveLoans(poolID types.U64) ([]ActiveLoanStorageEntry, error)
}

type api struct {
	centAPI centchain.API
}

func NewAPI(centAPI centchain.API) API {
	return &api{centAPI: centAPI}
}

func (a *api) GetCreatedLoan(poolID types.U64, loanID types.U64) (*CreatedLoanStorageEntry, error) {
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

	encodedPoolID, err := codec.Encode(poolID)

	if err != nil {
		log.Errorf("Couldn't encode pool ID: %s", err)

		return nil, ErrPoolIDEncoding
	}

	encodedLoanID, err := codec.Encode(loanID)

	if err != nil {
		log.Errorf("Couldn't encode loan ID: %s", err)

		return nil, ErrLoanIDEncoding
	}

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		CreatedLoanStorageName,
		encodedPoolID,
		encodedLoanID,
	)

	if err != nil {
		log.Errorf("Couldn't create storage key: %s", err)

		return nil, errors.ErrStorageKeyCreation
	}

	var createdLoan CreatedLoanStorageEntry

	ok, err := a.centAPI.GetStorageLatest(storageKey, &createdLoan)

	if err != nil {
		log.Errorf("Couldn't retrieve created loan from storage: %s", err)

		return nil, ErrCreatedLoanRetrieval
	}

	if !ok {
		log.Error("Created loan not found for pool ID %d and loan ID %d", poolID, loanID)

		return nil, ErrCreatedLoanNotFound
	}

	return &createdLoan, nil
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

		return nil, ErrPoolIDEncoding
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
