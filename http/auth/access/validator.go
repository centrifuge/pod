package access

import (
	"net/http"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	authToken "github.com/centrifuge/pod/http/auth/token"
)

//go:generate mockery --name Validator --structname ValidatorMock --filename validator_mock.go --inpackage

type Validator interface {
	Validate(req *http.Request, token *authToken.JW3Token) (*types.AccountID, error)
}
