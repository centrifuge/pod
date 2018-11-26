package p2p

import (
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/version"
)

// Validator defines method that must be implemented by any validator type.
type Validator interface {
	// Validate validates p2p requests
	Validate(header *p2ppb.CentrifugeHeader) error
}

// ValidatorGroup implements Validator for validating a set of validators.
type ValidatorGroup []Validator

// Validate will execute all group specific atomic validations
func (group ValidatorGroup) Validate(header *p2ppb.CentrifugeHeader) (errors error) {
	for _, v := range group {
		if err := v.Validate(header); err != nil {
			errors = documents.AppendError(errors, err)
		}
	}
	return errors
}

// ValidatorFunc implements Validator and can be used as a adaptor for functions
// with specific function signature
type ValidatorFunc func(header *p2ppb.CentrifugeHeader) error

// Validate passes the arguments to the underlying validator
// function and returns the results
func (vf ValidatorFunc) Validate(header *p2ppb.CentrifugeHeader) error {
	return vf(header)
}

func versionValidator() Validator {
	return ValidatorFunc(func(header *p2ppb.CentrifugeHeader) error {
		if header == nil {
			return fmt.Errorf("nil header")
		}
		if !version.CheckVersion(header.CentNodeVersion) {
			return version.IncompatibleVersionError(header.CentNodeVersion)
		}
		return nil
	})
}

func networkValidator(networkID uint32) Validator {
	return ValidatorFunc(func(header *p2ppb.CentrifugeHeader) error {
		if header == nil {
			return fmt.Errorf("nil header")
		}
		if networkID != header.NetworkIdentifier {
			return incompatibleNetworkError(networkID, header.NetworkIdentifier)
		}
		return nil
	})
}

func handshakeValidator(networkID uint32) ValidatorGroup {
	return ValidatorGroup{
		versionValidator(),
		networkValidator(networkID),
	}
}

func incompatibleNetworkError(configNetwork uint32, nodeNetwork uint32) error {
	return centerrors.New(code.NetworkMismatch, fmt.Sprintf("Incompatible network id: node network: %d, client network: %d", configNetwork, nodeNetwork))
}
