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

//go:generate mockery --name ValidationServiceFactory --structname ValidationServiceFactoryMock --filename factory_mock.go --inpackage

type ValidationServiceFactory interface {
	GetValidationServices() (ValidationServices, error)
}

type validationServiceFactory struct {
	configSrv               config.Service
	proxyAccessValidator    Validator
	adminAccessValidator    Validator
	investorAccessValidator Validator
}

func NewValidationServiceFactory(
	configSrv config.Service,
	proxyAccessValidator Validator,
	adminAccessValidator Validator,
	investorAccessValidator Validator,
) ValidationServiceFactory {
	return &validationServiceFactory{
		configSrv,
		proxyAccessValidator,
		adminAccessValidator,
		investorAccessValidator,
	}
}

func (v *validationServiceFactory) GetValidationServices() (ValidationServices, error) {
	cfg, err := v.configSrv.GetConfig()

	if err != nil {
		return nil, err
	}

	if !cfg.IsAuthenticationEnabled() {
		return ValidationServices{
			getPartialValidationService(v.configSrv),
		}, nil
	}

	return []ValidationService{
		getAdminValidationService(v.adminAccessValidator),
		getInvestorValidationService(v.configSrv, v.investorAccessValidator),
		getNoopValidationService(),
		getProxyValidationService(v.configSrv, v.proxyAccessValidator),
	}, nil
}

type pathMatcherFn func(path string) bool
type tokenValidationFn func(r *http.Request) (*authToken.JW3Token, error)
type accessValidationFn func(r *http.Request, token *authToken.JW3Token) (*types.AccountID, error)
type postAccessValidationFn func(r *http.Request, accountID *types.AccountID) error

//go:generate mockery --name ValidationService --structname ValidationServiceMock --filename service_mock.go --inpackage

type ValidationService interface {
	Validate(r *http.Request) error
	Matches(path string) bool
}

type ValidationServices []ValidationService

func (a ValidationServices) Validate(r *http.Request) error {
	path := r.URL.Path

	for _, service := range a {
		if !service.Matches(path) {
			continue
		}

		return service.Validate(r)
	}

	return ErrNoValidationServiceForPath
}

type validationService struct {
	pathMatchingFn     pathMatcherFn
	tokenValidationFn  tokenValidationFn
	accessValidationFn accessValidationFn
	postValidationFn   postAccessValidationFn
}

func NewValidationService(
	pathMatchingFn pathMatcherFn,
	tokenValidationFn tokenValidationFn,
	accessValidationFn accessValidationFn,
	postValidationFn postAccessValidationFn,
) ValidationService {
	return &validationService{
		pathMatchingFn,
		tokenValidationFn,
		accessValidationFn,
		postValidationFn,
	}
}

func (a *validationService) Validate(r *http.Request) error {
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

func (a *validationService) Matches(path string) bool {
	return a.pathMatchingFn(path)
}

type adminValidationService struct {
	ValidationService
}

var (
	adminPathRegex = regexp.MustCompile(`^/v2/accounts(|/generate|/0x[a-fA-F0-9]+)$`)
)

func getAdminValidationService(
	adminAccessValidator Validator,
) ValidationService {
	service := NewValidationService(
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

	return &adminValidationService{service}
}

type noopValidationService struct {
	ValidationService
}

var (
	noopAccessValidationPaths = map[string]struct{}{
		"/ping": {},
	}
)

func getNoopValidationService() ValidationService {
	service := NewValidationService(
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

	return &noopValidationService{service}
}

type investValidationService struct {
	ValidationService
}

var (
	investorPathRegex = regexp.MustCompile(`^/v3/investor.*$`)
)

func getInvestorValidationService(
	configSrv config.Service,
	investorAccessValidator Validator,
) ValidationService {
	service := NewValidationService(
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

	return &investValidationService{service}
}

type proxyValidationService struct {
	ValidationService
}

func getProxyValidationService(
	configSrv config.Service,
	proxyAccessValidator Validator,
) ValidationService {
	service := NewValidationService(
		func(_ string) bool {
			return true
		},
		getDefaultTokenValidationFn(),
		proxyAccessValidator.Validate,
		getDefaultPostValidationFn(configSrv),
	)

	return &proxyValidationService{service}
}

type partialValidationService struct {
	ValidationService
}

func getPartialValidationService(
	configSrv config.Service,
) ValidationService {
	service := NewValidationService(
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

	return &partialValidationService{service}
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
