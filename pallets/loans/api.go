package loans

import (
	"github.com/centrifuge/chain-custom-types/pkg/loans"
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
	ErrCreatedLoanRetrieval = errors.Error("created loan retrieval")
	ErrCreatedLoanNotFound  = errors.Error("created loan not found")
)

const (
	PalletName             = "Loans"
	CreatedLoanStorageName = "CreatedLoan"
)

type CreatedLoanStorageEntry struct {
	Info     loans.LoanInfo
	Borrower types.AccountID
}

//go:generate mockery --name API --structname APIMock --filename api_mock.go --inpackage

type API interface {
	GetCreatedLoan(poolID types.U64, loanID types.U64) (*CreatedLoanStorageEntry, error)
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
		log.Errorf("Created loan not found for pool ID %d and loan ID %d", poolID, loanID)

		return nil, ErrCreatedLoanNotFound
	}

	return &createdLoan, nil
}
