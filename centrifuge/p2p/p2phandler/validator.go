package p2phandler

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
)

type Validator interface {
	// Validate validates p2p requests
	Validate(param interface{}) error
}

// ValidatorGroup implements Validator for validating a set of validators.
type ValidatorGroup []Validator

// Validate will execute all group specific atomic validations
func (group ValidatorGroup) Validate(params []interface{}) (errors error) {

	if len(params) != len(group) {
		return centerrors.New(code.Unknown, fmt.Sprintf("Mismatched validator params, expected [%d], actual [%d]", len(group), len(params)))
	}

	for i, v := range group {
		if err := v.Validate(params[i]); err != nil {
			errors = documents.AppendError(errors, err)
		}
	}
	return errors
}

// ValidatorFunc implements Validator and can be used as a adaptor for functions
// with specific function signature
type ValidatorFunc func(param interface{}) error

// Validate passes the arguments to the underlying validator
// function and returns the results
func (vf ValidatorFunc) Validate(param interface{}) error {
	return vf(param)
}

func versionValidator() Validator {
	return ValidatorFunc(func(param interface{}) error {
		nodeVersion, ok := param.(string)
		if !ok {
			return centerrors.New(code.Unknown, fmt.Sprintf("Cannot convert param [%v] to string", param))
		}
		if !version.CheckVersion(nodeVersion) {
			return version.IncompatibleVersionError(nodeVersion)
		}
		return nil
	})
}

func networkValidator() Validator {
	return ValidatorFunc(func(param interface{}) error {
		networkID, ok := param.(uint32)
		if !ok {
			return centerrors.New(code.Unknown, fmt.Sprintf("Cannot convert param [%v] to uint32", param))
		}
		if config.Config.GetNetworkID() != networkID {
			return incompatibleNetworkError(config.Config.GetNetworkID(), networkID)
		}
		return nil
	})
}

func handshakeValidator() ValidatorGroup {
	return ValidatorGroup{
		versionValidator(),
		networkValidator(),
	}
}

func incompatibleNetworkError(configNetwork uint32, nodeNetwork uint32) error {
	return centerrors.New(code.NetworkMismatch, fmt.Sprintf("Incompatible network id: node network: %d, client network: %d", configNetwork, nodeNetwork))
}
