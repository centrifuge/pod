package access

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	authToken "github.com/centrifuge/pod/http/auth/token"
	"github.com/centrifuge/pod/http/coreapi"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	"github.com/centrifuge/pod/pallets/loans"
	"github.com/centrifuge/pod/pallets/permissions"
	"github.com/centrifuge/pod/pallets/uniques"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	logging "github.com/ipfs/go-log"
)

type investorAccessValidator struct {
	log            *logging.ZapEventLogger
	loansAPI       loans.API
	permissionsAPI permissions.API
	uniquesAPI     uniques.API
}

func NewInvestorAccessValidator(
	loansAPI loans.API,
	permissionsAPI permissions.API,
	uniquesAPI uniques.API,
) Validator {
	log := logging.Logger("http-investor-access-validator")

	return &investorAccessValidator{
		log:            log,
		loansAPI:       loansAPI,
		permissionsAPI: permissionsAPI,
		uniquesAPI:     uniquesAPI,
	}
}

func (i *investorAccessValidator) Validate(req *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
	params, err := getInvestorAccessParams(req)

	if err != nil {
		i.log.Errorf("Couldn't get investor access params: %s", err)

		return nil, ErrInvestorAccessParamsRetrieval
	}

	investorID, err := authToken.DecodeSS58Address(token.Payload.Address)

	if err != nil {
		i.log.Errorf("Couldn't decode investor account ID: %s", err)

		return nil, ErrSS58AddressDecode
	}

	if err := i.validatePoolPermissions(investorID, params.PoolID, permissions.POD_READ_ACCESS); err != nil {
		i.log.Errorf("Couldn't validate investor pool permissions: %s", err)

		return nil, err
	}

	loan, err := i.getActiveLoan(params.PoolID, params.LoanID)

	if err != nil {
		i.log.Errorf("Couldn't get active loan: %s", err)

		return nil, err
	}

	documentID, err := i.uniquesAPI.GetItemAttribute(
		loan.Info.Collateral.CollectionID,
		loan.Info.Collateral.ItemID,
		[]byte(nftv3.DocumentIDAttributeKey),
	)

	if err != nil {
		i.log.Errorf("Couldn't get document ID from NFT attribute: %s", err)

		return nil, ErrDocumentIDRetrieval
	}

	if !bytes.Equal(params.AssetID, documentID) {
		i.log.Error("Document IDs do not match")

		return nil, ErrDocumentIDMismatch
	}

	return &loan.Borrower, nil
}

func (i *investorAccessValidator) getActiveLoan(poolID types.U64, loanID types.U64) (*loans.ActiveLoan, error) {
	activeLoans, err := i.loansAPI.GetActiveLoans(poolID)

	if err != nil {
		i.log.Errorf("Couldn't retrieve active loans for pool ID %d: %s", poolID, err)

		return nil, ErrActiveLoansRetrieval
	}

	for _, activeLoan := range activeLoans {
		if activeLoan.ActiveLoan.LoanID == loanID {
			loan := activeLoan.ActiveLoan
			return &loan, nil
		}
	}

	i.log.Errorf("Active loan not found for pool ID %d, loan ID %d: %s", poolID, loanID, err)

	return nil, ErrActiveLoanNotFound
}

func (i *investorAccessValidator) validatePoolPermissions(
	accountID *types.AccountID,
	poolID types.U64,
	expectPoolPermissions permissions.PoolAdminRole,
) error {
	permissionRoles, err := i.permissionsAPI.GetPermissionRoles(accountID, poolID)

	if err != nil {
		i.log.Errorf("Couldn't get permission roles: %s", err)

		return err
	}

	if permissionRoles.PoolAdmin&expectPoolPermissions != expectPoolPermissions {
		i.log.Errorf("Invalid pool permissions: %d", permissionRoles.PoolAdmin)

		return ErrInvalidPoolPermissions
	}

	return nil
}

type InvestorAccessParams struct {
	AssetID []byte
	PoolID  types.U64
	LoanID  types.U64
}

func getInvestorAccessParams(req *http.Request) (*InvestorAccessParams, error) {
	poolID, err := strconv.Atoi(chi.URLParam(req, coreapi.PoolIDNameParam))

	if err != nil {
		return nil, fmt.Errorf("pool ID parsing: %w", err)
	}

	loanID, err := strconv.Atoi(chi.URLParam(req, coreapi.LoanIDNameParam))

	if err != nil {
		return nil, fmt.Errorf("loan ID parsing: %w", err)
	}

	assetID, err := hexutil.Decode(chi.URLParam(req, coreapi.AssetIDNameParam))

	if err != nil {
		return nil, fmt.Errorf("asset ID parsing: %w", err)
	}

	return &InvestorAccessParams{
		AssetID: assetID,
		PoolID:  types.U64(poolID),
		LoanID:  types.U64(loanID),
	}, nil
}
