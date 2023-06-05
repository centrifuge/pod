//go:build unit

package access

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"

	"github.com/centrifuge/pod/contextutil"

	"github.com/centrifuge/pod/testingutils/keyrings"

	"github.com/centrifuge/pod/utils"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	authToken "github.com/centrifuge/pod/http/auth/token"

	"github.com/centrifuge/pod/errors"

	"github.com/stretchr/testify/assert"

	configMocks "github.com/centrifuge/pod/config"
)

func TestValidationWrapperFactory(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)
	proxyAccessValidatorMock := NewValidatorMock(t)
	adminAccessValidatorMock := NewValidatorMock(t)
	investorAccessValidatorMock := NewValidatorMock(t)

	factory := NewValidationWrapperFactory(
		configSrvMock,
		proxyAccessValidatorMock,
		adminAccessValidatorMock,
		investorAccessValidatorMock,
	)

	configMock := configMocks.NewConfigurationMock(t)
	configMock.On("IsAuthenticationEnabled").
		Return(false).
		Once()

	configSrvMock.On("GetConfig").
		Return(configMock, nil)

	res, err := factory.GetValidationWrappers()
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.IsType(t, &partialValidationWrapper{}, res[0])

	configMock.On("IsAuthenticationEnabled").
		Return(true).
		Once()

	res, err = factory.GetValidationWrappers()
	assert.NoError(t, err)
	assert.Len(t, res, 4)
	assert.IsType(t, &adminValidationWrapper{}, res[0])
	assert.IsType(t, &investValidationWrapper{}, res[1])
	assert.IsType(t, &noopValidationWrapper{}, res[2])
	assert.IsType(t, &proxyValidationWrapper{}, res[3])
}

func TestValidationWrappers_Validate(t *testing.T) {
	validationWrapperMock1 := NewValidationWrapperMock(t)
	validationWrapperMock2 := NewValidationWrapperMock(t)

	validationWrappers := ValidationWrappers{
		validationWrapperMock1,
		validationWrapperMock2,
	}

	testPath := "path"
	req := &http.Request{URL: &url.URL{Path: testPath}}

	validationWrapperMock1.On("Matches", testPath).
		Return(true).
		Once()

	validationWrapperMock1.On("Validate", req).
		Return(nil).
		Once()

	err := validationWrappers.Validate(req)
	assert.NoError(t, err)

	validationWrapperMock1.On("Matches", testPath).
		Return(true).
		Once()

	validationWrapperErr := errors.New("error")

	validationWrapperMock1.On("Validate", req).
		Return(validationWrapperErr).
		Once()

	err = validationWrappers.Validate(req)
	assert.ErrorIs(t, err, validationWrapperErr)

	validationWrapperMock1.On("Matches", testPath).
		Return(false).
		Once()

	validationWrapperMock2.On("Matches", testPath).
		Return(true).
		Once()

	validationWrapperMock2.On("Validate", req).
		Return(validationWrapperErr).
		Once()

	err = validationWrappers.Validate(req)
	assert.ErrorIs(t, err, validationWrapperErr)

	validationWrapperMock1.On("Matches", testPath).
		Return(false).
		Once()

	validationWrapperMock2.On("Matches", testPath).
		Return(false).
		Once()

	err = validationWrappers.Validate(req)
	assert.ErrorIs(t, err, ErrNoValidationWrapperForPath)
}

func TestValidationWrapper_Validate(t *testing.T) {
	testReq := &http.Request{}
	testToken := &authToken.JW3Token{}
	testAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	testValidationWrapper := &validationWrapper{
		pathMatchingFn: func(path string) bool {
			assert.Equal(t, path, testReq.URL.Path)
			return true
		},
		tokenValidationFn: func(r *http.Request) (*authToken.JW3Token, error) {
			assert.Equal(t, testReq, r)

			return testToken, nil
		},
		accessValidationFn: func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
			assert.Equal(t, testReq, r)
			assert.Equal(t, testToken, token)

			return testAccountID, nil
		},
		postValidationFn: func(r *http.Request, accountID *types.AccountID) error {
			assert.Equal(t, testReq, r)
			assert.True(t, testAccountID.Equal(accountID))
			return nil
		},
	}

	err = testValidationWrapper.Validate(testReq)
	assert.NoError(t, err)
}

func TestValidationWrapper_Validate_TokenValidationFnError(t *testing.T) {
	testReq := &http.Request{}
	testAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	tokenValidationErr := errors.New("error")

	testValidationWrapper := &validationWrapper{
		pathMatchingFn: func(path string) bool {
			assert.Equal(t, path, testReq.URL.Path)
			return true
		},
		tokenValidationFn: func(r *http.Request) (*authToken.JW3Token, error) {
			assert.Equal(t, testReq, r)

			return nil, tokenValidationErr
		},
		accessValidationFn: func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
			t.Fatal("should not reach the access validation function")
			return testAccountID, nil
		},
		postValidationFn: func(r *http.Request, accountID *types.AccountID) error {
			t.Fatal("should not reach the post validation function")
			return nil
		},
	}

	err = testValidationWrapper.Validate(testReq)
	assert.ErrorIs(t, err, tokenValidationErr)
}

func TestValidationWrapper_Validate_AccessValidationError(t *testing.T) {
	testReq := &http.Request{}
	testToken := &authToken.JW3Token{}

	accessValidationErr := errors.New("error")

	testValidationWrapper := &validationWrapper{
		pathMatchingFn: func(path string) bool {
			assert.Equal(t, path, testReq.URL.Path)
			return true
		},
		tokenValidationFn: func(r *http.Request) (*authToken.JW3Token, error) {
			assert.Equal(t, testReq, r)

			return testToken, nil
		},
		accessValidationFn: func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
			assert.Equal(t, testReq, r)
			assert.Equal(t, testToken, token)

			return nil, accessValidationErr
		},
		postValidationFn: func(r *http.Request, accountID *types.AccountID) error {
			t.Fatal("should not reach the post validation function")
			return nil
		},
	}

	err := testValidationWrapper.Validate(testReq)
	assert.ErrorIs(t, err, accessValidationErr)
}

func TestValidationWrapper_Validate_PostValidationError(t *testing.T) {
	testReq := &http.Request{}
	testToken := &authToken.JW3Token{}
	testAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	postValidationErr := errors.New("error")

	testValidationWrapper := &validationWrapper{
		pathMatchingFn: func(path string) bool {
			assert.Equal(t, path, testReq.URL.Path)
			return true
		},
		tokenValidationFn: func(r *http.Request) (*authToken.JW3Token, error) {
			assert.Equal(t, testReq, r)

			return testToken, nil
		},
		accessValidationFn: func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
			assert.Equal(t, testReq, r)
			assert.Equal(t, testToken, token)

			return testAccountID, nil
		},
		postValidationFn: func(r *http.Request, accountID *types.AccountID) error {
			assert.Equal(t, testReq, r)
			assert.True(t, testAccountID.Equal(accountID))
			return postValidationErr
		},
	}

	err = testValidationWrapper.Validate(testReq)
	assert.ErrorIs(t, err, postValidationErr)
}

func Test_adminValidationWrapper_Matches(t *testing.T) {
	validatorMock := NewValidatorMock(t)

	adminValidationWrapper := getAdminValidationWrapper(validatorMock)

	tests := []struct {
		Path          string
		MatchExpected bool
	}{
		{
			Path:          "/v2/accounts",
			MatchExpected: true,
		},
		{
			Path:          "/v2/accounts/generate",
			MatchExpected: true,
		},
		{
			Path:          "/v2/accounts/0xabc0123",
			MatchExpected: true,
		},
		{
			Path:          "/v2/accounts/0xabc0123/sign",
			MatchExpected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Path, func(t *testing.T) {
			assert.Equal(t, test.MatchExpected, adminValidationWrapper.Matches(test.Path))
		})
	}
}

func Test_adminValidationWrapper_Validate(t *testing.T) {
	validatorMock := NewValidatorMock(t)

	adminValidationWrapper := getAdminValidationWrapper(validatorMock)

	token, tokenStr := getValidTestToken(t, authToken.PodAdminProxyType)

	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	validatorMock.On("Validate", req, token).
		Return(accountID, nil).
		Once()

	err = adminValidationWrapper.Validate(req)
	assert.NoError(t, err)

	// Validation error

	validationErr := errors.New("error")

	validatorMock.On("Validate", req, token).
		Return(nil, validationErr).
		Once()

	err = adminValidationWrapper.Validate(req)
	assert.ErrorIs(t, err, validationErr)

	// Invalid proxy type

	_, tokenStr = getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	err = adminValidationWrapper.Validate(req)
	assert.ErrorIs(t, err, ErrInvalidProxyType)
}

func Test_noopValidationWrapper_Validate(t *testing.T) {
	noopValidationWrapper := getNoopValidationWrapper()

	req := &http.Request{}

	err := noopValidationWrapper.Validate(req)
	assert.NoError(t, err)
}

func Test_investorValidationWrapper_Matches(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)
	validatorMock := NewValidatorMock(t)

	investorValidationWrapper := getInvestorValidationWrapper(configSrvMock, validatorMock)

	tests := []struct {
		Path          string
		MatchExpected bool
	}{
		{
			Path:          "/v3/investor",
			MatchExpected: true,
		},
		{
			Path:          "/v3/investor/asset",
			MatchExpected: true,
		},
		{
			Path:          "/v3/investor/test",
			MatchExpected: true,
		},
		{
			Path:          "/v2/investor",
			MatchExpected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Path, func(t *testing.T) {
			assert.Equal(t, test.MatchExpected, investorValidationWrapper.Matches(test.Path))
		})
	}
}

func Test_investorValidationWrapper_Validate(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)
	validatorMock := NewValidatorMock(t)

	investorValidationWrapper := getInvestorValidationWrapper(configSrvMock, validatorMock)

	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	validatorMock.On("Validate", req, token).
		Return(accountID, nil).
		Once()

	accountMock := configMocks.NewAccountMock(t)

	configSrvMock.On("GetAccount", accountID.ToBytes()).
		Return(accountMock, nil).
		Once()

	err = investorValidationWrapper.Validate(req)
	assert.NoError(t, err)

	reqContextAccount, err := contextutil.Account(req.Context())
	assert.NoError(t, err)

	assert.Equal(t, accountMock, reqContextAccount)
}

func Test_investorValidationWrapper_Validate_ValidationError(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)
	validatorMock := NewValidatorMock(t)

	investorValidationWrapper := getInvestorValidationWrapper(configSrvMock, validatorMock)

	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	validationErr := errors.New("error")

	validatorMock.On("Validate", req, token).
		Return(nil, validationErr).
		Once()

	err := investorValidationWrapper.Validate(req)
	assert.ErrorIs(t, err, validationErr)
}

func Test_investorValidationWrapper_Validate_PostValidationError(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)
	validatorMock := NewValidatorMock(t)

	investorValidationWrapper := getInvestorValidationWrapper(configSrvMock, validatorMock)

	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	validatorMock.On("Validate", req, token).
		Return(accountID, nil).
		Once()

	configSrvMock.On("GetAccount", accountID.ToBytes()).
		Return(nil, errors.New("error")).
		Once()

	err = investorValidationWrapper.Validate(req)
	assert.ErrorIs(t, err, ErrIdentityNotFound)
}

func Test_proxyValidationWrapper_Validate(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)
	validatorMock := NewValidatorMock(t)

	proxyValidationWrapper := getProxyValidationWrapper(configSrvMock, validatorMock)

	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	validatorMock.On("Validate", req, token).
		Return(accountID, nil).
		Once()

	accountMock := configMocks.NewAccountMock(t)

	configSrvMock.On("GetAccount", accountID.ToBytes()).
		Return(accountMock, nil).
		Once()

	err = proxyValidationWrapper.Validate(req)
	assert.NoError(t, err)

	reqContextAccount, err := contextutil.Account(req.Context())
	assert.NoError(t, err)

	assert.Equal(t, accountMock, reqContextAccount)
}

func Test_proxyValidationWrapper_Validate_InvalidProxyType(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)
	validatorMock := NewValidatorMock(t)

	proxyValidationWrapper := getProxyValidationWrapper(configSrvMock, validatorMock)

	_, tokenStr := getValidTestToken(t, "invalid_proxy_type")

	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	err := proxyValidationWrapper.Validate(req)
	assert.Error(t, err)
}

func Test_proxyValidationWrapper_Validate_ValidationError(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)
	validatorMock := NewValidatorMock(t)

	proxyValidationWrapper := getProxyValidationWrapper(configSrvMock, validatorMock)

	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	validationError := errors.New("error")

	validatorMock.On("Validate", req, token).
		Return(nil, validationError).
		Once()

	err := proxyValidationWrapper.Validate(req)
	assert.ErrorIs(t, err, validationError)
}

func Test_proxyValidationWrapper_Validate_PostValidationError(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)
	validatorMock := NewValidatorMock(t)

	proxyValidationWrapper := getProxyValidationWrapper(configSrvMock, validatorMock)

	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	validatorMock.On("Validate", req, token).
		Return(accountID, nil).
		Once()

	configSrvMock.On("GetAccount", accountID.ToBytes()).
		Return(nil, errors.New("error")).
		Once()

	err = proxyValidationWrapper.Validate(req)
	assert.ErrorIs(t, err, ErrIdentityNotFound)
}

func Test_partialValidationWrapper_Validate_NoToken(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)

	partialValidationWrapper := getPartialValidationWrapper(configSrvMock)

	req := &http.Request{Header: make(map[string][]string)}

	err := partialValidationWrapper.Validate(req)
	assert.NoError(t, err)
}

func Test_partialValidationWrapper_Validate_WithToken(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)

	partialValidationWrapper := getPartialValidationWrapper(configSrvMock)

	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{Header: make(map[string][]string)}
	req.Header.Add("Authorization", "Bearer "+tokenStr)

	accountID, err := authToken.DecodeSS58Address(token.Payload.Address)
	assert.NoError(t, err)

	accountMock := configMocks.NewAccountMock(t)

	configSrvMock.On("GetAccount", accountID.ToBytes()).
		Return(accountMock, nil).
		Once()

	err = partialValidationWrapper.Validate(req)
	assert.NoError(t, err)

	reqContextAccount, err := contextutil.Account(req.Context())
	assert.NoError(t, err)

	assert.Equal(t, accountMock, reqContextAccount)
}

func Test_partialValidationWrapper_Validate_WithInvalidToken(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)

	partialValidationWrapper := getPartialValidationWrapper(configSrvMock)

	req := &http.Request{Header: make(map[string][]string)}
	req.Header.Add("Authorization", "Bearer "+"invalid_token")

	err := partialValidationWrapper.Validate(req)
	assert.NoError(t, err)

	reqContextAccount, err := contextutil.Account(req.Context())
	assert.Error(t, err)
	assert.Nil(t, reqContextAccount)
}

func Test_parseToken(t *testing.T) {
	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{Header: make(map[string][]string)}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer%s%s", authorizationHeaderSeparator, tokenStr))

	res, err := parseToken(req)
	assert.NoError(t, err)
	assert.Equal(t, token, res)
}

func Test_parseToken_InvalidAuthHeaders(t *testing.T) {
	tests := []struct {
		AuthHeader    string
		ExpectedError error
	}{
		{
			AuthHeader:    "",
			ExpectedError: ErrInvalidAuthorizationHeader,
		},
		{
			AuthHeader:    "Bearer",
			ExpectedError: ErrInvalidAuthorizationHeader,
		},
		{
			AuthHeader:    "Bearer ",
			ExpectedError: authToken.ErrInvalidJW3Token,
		},
		{
			AuthHeader:    "Bearer test",
			ExpectedError: authToken.ErrInvalidJW3Token,
		},
	}

	for _, test := range tests {
		t.Run(test.AuthHeader, func(t *testing.T) {
			req := &http.Request{Header: make(map[string][]string)}
			req.Header.Add("Authorization", test.AuthHeader)

			res, err := parseToken(req)
			assert.ErrorIs(t, err, test.ExpectedError)
			assert.Nil(t, res)
		})
	}
}

func Test_getTokenValidationFn(t *testing.T) {
	tokenValidatorMock := authToken.NewValidatorMock(t)

	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{Header: make(map[string][]string)}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer%s%s", authorizationHeaderSeparator, tokenStr))

	tokenValidationFn := getTokenValidationFn(tokenValidatorMock)

	tokenValidatorMock.On("Validate", token).
		Return(nil).
		Once()

	res, err := tokenValidationFn(req)
	assert.NoError(t, err)
	assert.Equal(t, token, res)
}

func Test_getTokenValidationFn_ParseTokenError(t *testing.T) {
	tokenValidatorMock := authToken.NewValidatorMock(t)

	req := &http.Request{Header: make(map[string][]string)}

	tokenValidationFn := getTokenValidationFn(tokenValidatorMock)

	res, err := tokenValidationFn(req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func Test_getTokenValidationFn_TokenValidatorError(t *testing.T) {
	tokenValidatorMock := authToken.NewValidatorMock(t)

	token, tokenStr := getValidTestToken(t, proxyType.ProxyTypeName[proxyType.Any])

	req := &http.Request{Header: make(map[string][]string)}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer%s%s", authorizationHeaderSeparator, tokenStr))

	tokenValidationFn := getTokenValidationFn(tokenValidatorMock)

	tokenValidatorError := errors.New("error")

	tokenValidatorMock.On("Validate", token).
		Return(tokenValidatorError).
		Once()

	res, err := tokenValidationFn(req)
	assert.ErrorIs(t, err, tokenValidatorError)
	assert.Nil(t, res)
}

func getValidTestToken(t *testing.T, proxyType string) (*authToken.JW3Token, string) {
	accountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	tokenStr, err := authToken.CreateJW3Token(
		accountID,
		accountID,
		keyrings.AliceKeyRingPair.URI,
		proxyType,
	)
	assert.NoError(t, err)

	token, err := authToken.DecodeJW3Token(tokenStr)
	assert.NoError(t, err)

	return token, tokenStr
}
