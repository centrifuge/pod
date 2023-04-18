package access

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/errors"
	authToken "github.com/centrifuge/pod/http/auth/token"
)

type ValidationWrapperFactory interface {
	GetValidationWrappers() (ValidationWrappers, error)
}

type validationWrapperFactory struct {
	configSrv               config.Service
	proxyAccessValidator    Validator
	adminAccessValidator    Validator
	investorAccessValidator Validator
}

func NewValidationWrapperFactory(
	configSrv config.Service,
	proxyAccessValidator Validator,
	adminAccessValidator Validator,
	investorAccessValidator Validator,
) ValidationWrapperFactory {
	return &validationWrapperFactory{
		configSrv,
		proxyAccessValidator,
		adminAccessValidator,
		investorAccessValidator,
	}
}

func (v *validationWrapperFactory) GetValidationWrappers() (ValidationWrappers, error) {
	cfg, err := v.configSrv.GetConfig()

	if err != nil {
		return nil, err
	}

	if !cfg.IsAuthenticationEnabled() {
		return ValidationWrappers{
			getPartialValidationWrapper(v.configSrv),
		}, nil
	}

	return DefaultValidationWrappers(
		v.configSrv,
		v.proxyAccessValidator,
		v.adminAccessValidator,
		v.investorAccessValidator,
	), nil
}

type pathMatcherFn func(path string) bool
type tokenValidationFn func(r *http.Request) (*authToken.JW3Token, error)
type accessValidationFn func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error)
type postAccessValidationFn func(r *http.Request, accountID *types.AccountID) error

type ValidationWrapper interface {
	Validate(r *http.Request) error
	Matches(path string) bool
}

type ValidationWrappers []ValidationWrapper

func (a ValidationWrappers) Validate(r *http.Request) error {
	path := r.URL.Path

	for _, wrapper := range a {
		if !wrapper.Matches(path) {
			continue
		}

		return wrapper.Validate(r)
	}

	return fmt.Errorf("no access validation wrapper found for the provided path - %s", path)
}

type validationWrapper struct {
	pathMatchingFn     pathMatcherFn
	tokenValidationFn  tokenValidationFn
	accessValidationFn accessValidationFn
	postValidationFn   postAccessValidationFn
}

func NewValidationWrapper(
	pathMatchingFn pathMatcherFn,
	tokenValidationFn tokenValidationFn,
	accessValidationFn accessValidationFn,
	postValidationFn postAccessValidationFn,
) ValidationWrapper {
	return &validationWrapper{
		pathMatchingFn,
		tokenValidationFn,
		accessValidationFn,
		postValidationFn,
	}
}

func (a *validationWrapper) Validate(r *http.Request) error {
	token, err := a.tokenValidationFn(r)

	if err != nil {
		return err
	}

	identity, err := a.accessValidationFn(r, token)

	if err != nil {
		return err
	}

	if err := a.postValidationFn(r, identity); err != nil {
		return err
	}

	return nil
}

func (a *validationWrapper) Matches(path string) bool {
	return a.pathMatchingFn(path)
}

func DefaultValidationWrappers(
	configSrv config.Service,
	proxyAccessValidator Validator,
	adminAccessValidator Validator,
	investorAccessValidator Validator,
) ValidationWrappers {
	return []ValidationWrapper{
		getAdminValidationWrapper(adminAccessValidator),
		getInvestorValidationWrapper(configSrv, investorAccessValidator),
		getNoopValidationWrapper(),
		getProxyValidationWrapper(configSrv, proxyAccessValidator),
	}
}

var (
	adminPathRegex = regexp.MustCompile(`^/v2/accounts(|/generate|/0x[a-fA-F0-9]+)$`)
)

func getAdminValidationWrapper(
	adminAccessValidator Validator,
) ValidationWrapper {
	return NewValidationWrapper(
		func(path string) bool {
			return adminPathRegex.MatchString(path)
		},
		getDefaultTokenValidationFn(),
		func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
			return adminAccessValidator.Validate(r, token)
		},
		func(_ *http.Request, _ *types.AccountID) error {
			return nil
		},
	)
}

var (
	noopAccessValidationPaths = map[string]struct{}{
		"/ping": {},
	}
)

func getNoopValidationWrapper() ValidationWrapper {
	return NewValidationWrapper(
		func(path string) bool {
			_, ok := noopAccessValidationPaths[path]

			return ok
		},
		func(_ *http.Request) (*authToken.JW3Token, error) {
			return nil, nil
		},
		func(_ *http.Request, _ *authToken.JW3Token) (*types.AccountID, error) {
			return nil, nil
		},
		func(r *http.Request, _ *types.AccountID) error {
			return nil
		},
	)
}

var (
	investorPathRegex = regexp.MustCompile(`^/v3/investors.*$`)
)

func getInvestorValidationWrapper(
	configSrv config.Service,
	investorAccessValidator Validator,
) ValidationWrapper {
	return NewValidationWrapper(
		func(path string) bool {
			return investorPathRegex.MatchString(path)
		},
		getTokenValidationFn(authToken.NewSR25519TokenValidator(
			[]func(header *authToken.JW3THeader) error{
				authToken.BasicHeaderValidationFn,
				authToken.SR25519HeaderValidationFn,
			},
			[]func(payload *authToken.JW3TPayload) error{
				authToken.BasicPayloadValidationFn,
			},
		)),
		func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
			return investorAccessValidator.Validate(r, token)
		},
		getDefaultPostValidationFn(configSrv),
	)
}

func getProxyValidationWrapper(
	configSrv config.Service,
	proxyAccessValidator Validator,
) ValidationWrapper {
	return NewValidationWrapper(
		func(_ string) bool {
			return true
		},
		getDefaultTokenValidationFn(),
		func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
			return proxyAccessValidator.Validate(r, token)
		},
		getDefaultPostValidationFn(configSrv),
	)
}

func getPartialValidationWrapper(
	configSrv config.Service,
) ValidationWrapper {
	return NewValidationWrapper(
		func(path string) bool {
			return true
		},
		func(r *http.Request) (*authToken.JW3Token, error) {
			token, err := parseToken(r)

			if err != nil {
				return nil, nil
			}

			return token, nil
		},
		func(_ *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
			if token == nil {
				return nil, nil
			}

			return authToken.DecodeSS58Address(token.Payload.Address)
		},
		func(r *http.Request, accountID *types.AccountID) error {
			if accountID == nil {
				return nil
			}

			defaultPostValidationFn := getDefaultPostValidationFn(configSrv)

			return defaultPostValidationFn(r, accountID)
		},
	)
}

const (
	authorizationHeaderName      = "Authorization"
	authorizationHeaderSeparator = " "
)

func parseToken(r *http.Request) (*authToken.JW3Token, error) {
	authHeader := r.Header.Get(authorizationHeaderName)
	bearer := strings.Split(authHeader, authorizationHeaderSeparator)

	if len(bearer) != 2 {
		return nil, errors.New("invalid authorization header")
	}

	return authToken.DecodeJW3Token(bearer[1])
}

func getTokenValidationFn(tokenValidator authToken.Validator) tokenValidationFn {
	return func(r *http.Request) (*authToken.JW3Token, error) {
		token, err := parseToken(r)

		if err != nil {
			return nil, err
		}

		if err := tokenValidator.Validate(token); err != nil {
			return nil, err
		}

		return token, nil
	}
}

func getDefaultTokenValidationFn() tokenValidationFn {
	return getTokenValidationFn(authToken.DefaultSR25519TokenValidator())
}

func getDefaultPostValidationFn(configSrv config.Service) postAccessValidationFn {
	return func(r *http.Request, accountID *types.AccountID) error {
		acc, err := configSrv.GetAccount(accountID.ToBytes())
		if err != nil {
			return errors.New("identity not found")
		}

		ctx := contextutil.WithAccount(r.Context(), acc)

		*r = *r.WithContext(ctx)

		return nil
	}
}
