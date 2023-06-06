//go:build unit

package access

import (
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"testing"

	"github.com/centrifuge/pod/errors"

	"github.com/ethereum/go-ethereum/common/hexutil"

	authToken "github.com/centrifuge/pod/http/auth/token"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	"github.com/centrifuge/pod/utils"
	"github.com/vedhavyas/go-subkey"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/http/coreapi"
	"github.com/centrifuge/pod/pallets/loans"
	"github.com/centrifuge/pod/pallets/permissions"
	"github.com/centrifuge/pod/pallets/uniques"
	"github.com/stretchr/testify/assert"
)

func TestInvestorAccessValidator_Validate(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	poolID := types.U64(rand.Uint32())
	loanID := types.U64(rand.Uint32())
	documentID := "document_id"

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%d&%s=%s",
		coreapi.PoolIDQueryParam, poolID,
		coreapi.LoanIDQueryParam, loanID,
		coreapi.AssetIDQueryParam, hexutil.Encode([]byte(documentID)),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	investorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	investSSS58Address, err := subkey.SS58Address(investorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: investSSS58Address,
		},
	}

	permissionRoles := &permissions.PermissionRoles{PoolAdmin: permissions.PodReadAccess}

	permissionsAPIMock.On("GetPermissionRoles", investorAccountID, poolID).
		Return(permissionRoles, nil).
		Once()

	borrowerAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	collectionID := types.U64(rand.Uint32())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	loan := &loans.CreatedLoanStorageEntry{
		Info: loans.LoanInfo{
			Collateral: loans.Asset{
				CollectionID: collectionID,
				ItemID:       itemID,
			},
		},
		Borrower: *borrowerAccountID,
	}

	loansAPIMock.On("GetCreatedLoan", poolID, loanID).
		Return(loan, nil).
		Once()

	uniquesAPIMock.On(
		"GetItemAttribute",
		collectionID,
		itemID,
		[]byte(nftv3.DocumentIDAttributeKey),
	).
		Return([]byte(documentID), nil).
		Once()

	res, err := investorAccessValidator.Validate(req, token)
	assert.NoError(t, err)
	assert.Equal(t, borrowerAccountID, res)
}

func TestInvestorAccessValidator_Validate_InvalidPoolID(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	loanID := types.U64(rand.Uint32())
	documentID := "document_id"

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%s",
		coreapi.LoanIDQueryParam, loanID,
		coreapi.AssetIDQueryParam, hexutil.Encode([]byte(documentID)),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	investorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	investSSS58Address, err := subkey.SS58Address(investorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: investSSS58Address,
		},
	}

	res, err := investorAccessValidator.Validate(req, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrInvestorAccessParamsRetrieval)
}

func TestInvestorAccessValidator_Validate_InvalidLoanID(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	poolID := types.U64(rand.Uint32())
	documentID := "document_id"

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%s",
		coreapi.PoolIDQueryParam, poolID,
		coreapi.AssetIDQueryParam, hexutil.Encode([]byte(documentID)),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	investorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	investSSS58Address, err := subkey.SS58Address(investorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: investSSS58Address,
		},
	}

	res, err := investorAccessValidator.Validate(req, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrInvestorAccessParamsRetrieval)
}

func TestInvestorAccessValidator_Validate_InvalidAssetID(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	poolID := types.U64(rand.Uint32())
	loanID := types.U64(rand.Uint32())

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%d",
		coreapi.PoolIDQueryParam, poolID,
		coreapi.LoanIDQueryParam, loanID,
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	investorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	investSSS58Address, err := subkey.SS58Address(investorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: investSSS58Address,
		},
	}

	res, err := investorAccessValidator.Validate(req, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrInvestorAccessParamsRetrieval)
}

func TestInvestorAccessValidator_Validate_InvalidInvestorAccountID(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	poolID := types.U64(rand.Uint32())
	loanID := types.U64(rand.Uint32())
	documentID := "document_id"

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%d&%s=%s",
		coreapi.PoolIDQueryParam, poolID,
		coreapi.LoanIDQueryParam, loanID,
		coreapi.AssetIDQueryParam, hexutil.Encode([]byte(documentID)),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: "",
		},
	}

	res, err := investorAccessValidator.Validate(req, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrSS58AddressDecode)
}

func TestInvestorAccessValidator_Validate_PermissionRolesRetrievalError(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	poolID := types.U64(rand.Uint32())
	loanID := types.U64(rand.Uint32())
	documentID := "document_id"

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%d&%s=%s",
		coreapi.PoolIDQueryParam, poolID,
		coreapi.LoanIDQueryParam, loanID,
		coreapi.AssetIDQueryParam, hexutil.Encode([]byte(documentID)),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	investorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	investSSS58Address, err := subkey.SS58Address(investorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: investSSS58Address,
		},
	}

	permissionsAPIMock.On("GetPermissionRoles", investorAccountID, poolID).
		Return(nil, errors.New("error")).
		Once()

	res, err := investorAccessValidator.Validate(req, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrPermissionRolesRetrievalError)
}

func TestInvestorAccessValidator_Validate_InvalidPoolPermissions(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	poolID := types.U64(rand.Uint32())
	loanID := types.U64(rand.Uint32())
	documentID := "document_id"

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%d&%s=%s",
		coreapi.PoolIDQueryParam, poolID,
		coreapi.LoanIDQueryParam, loanID,
		coreapi.AssetIDQueryParam, hexutil.Encode([]byte(documentID)),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	investorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	investSSS58Address, err := subkey.SS58Address(investorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: investSSS58Address,
		},
	}

	permissionRoles := &permissions.PermissionRoles{PoolAdmin: permissions.Borrower}

	permissionsAPIMock.On("GetPermissionRoles", investorAccountID, poolID).
		Return(permissionRoles, nil).
		Once()

	res, err := investorAccessValidator.Validate(req, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrInvalidPoolPermissions)
}

func TestInvestorAccessValidator_Validate_CreatedLoanRetrievalError(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	poolID := types.U64(rand.Uint32())
	loanID := types.U64(rand.Uint32())
	documentID := "document_id"

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%d&%s=%s",
		coreapi.PoolIDQueryParam, poolID,
		coreapi.LoanIDQueryParam, loanID,
		coreapi.AssetIDQueryParam, hexutil.Encode([]byte(documentID)),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	investorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	investSSS58Address, err := subkey.SS58Address(investorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: investSSS58Address,
		},
	}

	permissionRoles := &permissions.PermissionRoles{PoolAdmin: permissions.PodReadAccess}

	permissionsAPIMock.On("GetPermissionRoles", investorAccountID, poolID).
		Return(permissionRoles, nil).
		Once()

	loansAPIMock.On("GetCreatedLoan", poolID, loanID).
		Return(nil, errors.New("error")).
		Once()

	res, err := investorAccessValidator.Validate(req, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrCreatedLoanRetrieval)
}

func TestInvestorAccessValidator_Validate_DocumentIDRetrievalError(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	poolID := types.U64(rand.Uint32())
	loanID := types.U64(rand.Uint32())
	documentID := "document_id"

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%d&%s=%s",
		coreapi.PoolIDQueryParam, poolID,
		coreapi.LoanIDQueryParam, loanID,
		coreapi.AssetIDQueryParam, hexutil.Encode([]byte(documentID)),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	investorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	investSSS58Address, err := subkey.SS58Address(investorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: investSSS58Address,
		},
	}

	permissionRoles := &permissions.PermissionRoles{PoolAdmin: permissions.PodReadAccess}

	permissionsAPIMock.On("GetPermissionRoles", investorAccountID, poolID).
		Return(permissionRoles, nil).
		Once()

	borrowerAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	collectionID := types.U64(rand.Uint32())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	loan := &loans.CreatedLoanStorageEntry{
		Info: loans.LoanInfo{
			Collateral: loans.Asset{
				CollectionID: collectionID,
				ItemID:       itemID,
			},
		},
		Borrower: *borrowerAccountID,
	}

	loansAPIMock.On("GetCreatedLoan", poolID, loanID).
		Return(loan, nil).
		Once()

	uniquesAPIMock.On(
		"GetItemAttribute",
		collectionID,
		itemID,
		[]byte(nftv3.DocumentIDAttributeKey),
	).
		Return(nil, errors.New("error")).
		Once()

	res, err := investorAccessValidator.Validate(req, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrDocumentIDRetrieval)
}

func TestInvestorAccessValidator_Validate_DocumentIDMismatch(t *testing.T) {
	loansAPIMock := loans.NewAPIMock(t)
	permissionsAPIMock := permissions.NewAPIMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

	investorAccessValidator := NewInvestorAccessValidator(loansAPIMock, permissionsAPIMock, uniquesAPIMock)

	poolID := types.U64(rand.Uint32())
	loanID := types.U64(rand.Uint32())
	documentID := "document_id"

	reqURL, err := url.Parse("http://localhost/v3/investors")
	assert.NoError(t, err)

	reqURL.RawQuery = fmt.Sprintf(
		"%s=%d&%s=%d&%s=%s",
		coreapi.PoolIDQueryParam, poolID,
		coreapi.LoanIDQueryParam, loanID,
		coreapi.AssetIDQueryParam, hexutil.Encode([]byte(documentID)),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	assert.NoError(t, err)

	investorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	investSSS58Address, err := subkey.SS58Address(investorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: investSSS58Address,
		},
	}

	permissionRoles := &permissions.PermissionRoles{PoolAdmin: permissions.PodReadAccess}

	permissionsAPIMock.On("GetPermissionRoles", investorAccountID, poolID).
		Return(permissionRoles, nil).
		Once()

	borrowerAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	collectionID := types.U64(rand.Uint32())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	loan := &loans.CreatedLoanStorageEntry{
		Info: loans.LoanInfo{
			Collateral: loans.Asset{
				CollectionID: collectionID,
				ItemID:       itemID,
			},
		},
		Borrower: *borrowerAccountID,
	}

	loansAPIMock.On("GetCreatedLoan", poolID, loanID).
		Return(loan, nil).
		Once()

	uniquesAPIMock.On(
		"GetItemAttribute",
		collectionID,
		itemID,
		[]byte(nftv3.DocumentIDAttributeKey),
	).
		Return([]byte("some_other_document_id"), nil).
		Once()

	res, err := investorAccessValidator.Validate(req, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrDocumentIDMismatch)
}
