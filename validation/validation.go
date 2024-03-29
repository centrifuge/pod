package validation

import (
	"net/url"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type Validator interface {
	Validate() error
}

type genericValidator[T any] struct {
	t   T
	fns []func(T) error
}

func (g *genericValidator[T]) Validate() error {
	for _, fn := range g.fns {
		if err := fn(g.t); err != nil {
			return err
		}
	}

	return nil
}

func NewValidator[T any](t T, validatorFns ...func(T) error) Validator {
	return &genericValidator[T]{
		t:   t,
		fns: validatorFns,
	}
}

func Validate(validators ...Validator) error {
	for _, validator := range validators {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}

var (
	URLValidationFn = func(u string) error {
		if u == "" {
			return ErrMissingURL
		}

		if _, err := url.ParseRequestURI(u); err != nil {
			return ErrInvalidURL
		}

		return nil
	}

	AccountIDValidationFn = func(accountID *types.AccountID) error {
		if accountID == nil {
			return ErrInvalidAccountID
		}

		return nil
	}

	U64ValidationFn = func(v types.U64) error {
		if v == 0 {
			return ErrInvalidU64
		}

		return nil
	}
)
