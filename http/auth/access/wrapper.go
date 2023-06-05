package access

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	authToken "github.com/centrifuge/pod/http/auth/token"
)

//go:generate mockery --name ValidationWrapperFactory --structname ValidationWrapperFactoryMock --filename factory_mock.go --inpackage

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

	return []ValidationWrapper{
		getAdminValidationWrapper(v.adminAccessValidator),
		getInvestorValidationWrapper(v.configSrv, v.investorAccessValidator),
		getNoopValidationWrapper(),
		getProxyValidationWrapper(v.configSrv, v.proxyAccessValidator),
	}, nil
}

type pathMatcherFn func(path string) bool
type tokenValidationFn func(r *http.Request) (*authToken.JW3Token, error)
type accessValidationFn func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error)
type postAccessValidationFn func(r *http.Request, accountID *types.AccountID) error

//go:generate mockery --name ValidationWrapper --structname ValidationWrapperMock --filename wrapper_mock.go --inpackage

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

	return ErrNoValidationWrapperForPath
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

type adminValidationWrapper struct {
	ValidationWrapper
}

var (
	adminPathRegex = regexp.MustCompile(`^/v2/accounts(|/generate|/0x[a-fA-F0-9]+)$`)
)

func getAdminValidationWrapper(
	adminAccessValidator Validator,
) ValidationWrapper {
	wrapper := NewValidationWrapper(
		func(path string) bool {
			return adminPathRegex.MatchString(path)
		},

		func(r *http.Request) (*authToken.JW3Token, error) {
			token, err := getDefaultTokenValidationFn()(r)

			if err != nil {
				return nil, err
			}

			if token.Payload.ProxyType != authToken.PodAdminProxyType {
				return nil, ErrInvalidProxyType
			}

			return token, nil
		},
		adminAccessValidator.Validate,
		func(_ *http.Request, _ *types.AccountID) error {
			return nil
		},
	)

	return &adminValidationWrapper{wrapper}
}

type noopValidationWrapper struct {
	ValidationWrapper
}

var (
	noopAccessValidationPaths = map[string]struct{}{
		"/ping": {},
	}
)

func getNoopValidationWrapper() ValidationWrapper {
	wrapper := NewValidationWrapper(
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

	return &noopValidationWrapper{wrapper}
}

type investValidationWrapper struct {
	ValidationWrapper
}

var (
	investorPathRegex = regexp.MustCompile(`^/v3/investor.*$`)
)

func getInvestorValidationWrapper(
	configSrv config.Service,
	investorAccessValidator Validator,
) ValidationWrapper {
	wrapper := NewValidationWrapper(
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
		investorAccessValidator.Validate,
		getDefaultPostValidationFn(configSrv),
	)

	return &investValidationWrapper{wrapper}
}

type proxyValidationWrapper struct {
	ValidationWrapper
}

func getProxyValidationWrapper(
	configSrv config.Service,
	proxyAccessValidator Validator,
) ValidationWrapper {
	wrapper := NewValidationWrapper(
		func(_ string) bool {
			return true
		},
		getDefaultTokenValidationFn(),
		proxyAccessValidator.Validate,
		getDefaultPostValidationFn(configSrv),
	)

	return &proxyValidationWrapper{wrapper}
}

type partialValidationWrapper struct {
	ValidationWrapper
}

func getPartialValidationWrapper(
	configSrv config.Service,
) ValidationWrapper {
	wrapper := NewValidationWrapper(
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

			return getDefaultPostValidationFn(configSrv)(r, accountID)
		},
	)

	return &partialValidationWrapper{wrapper}
}

const (
	authorizationHeaderName      = "Authorization"
	authorizationHeaderSeparator = " "
)

func parseToken(r *http.Request) (*authToken.JW3Token, error) {
	authHeader := r.Header.Get(authorizationHeaderName)
	bearer := strings.Split(authHeader, authorizationHeaderSeparator)

	if len(bearer) != 2 {
		return nil, ErrInvalidAuthorizationHeader
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
			return ErrIdentityNotFound
		}

		ctx := contextutil.WithAccount(r.Context(), acc)

		*r = *r.WithContext(ctx)

		return nil
	}
}
