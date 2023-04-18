package access

import (
	"net/http"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	authToken "github.com/centrifuge/pod/http/auth/token"
)

type Validator interface {
	Validate(req *http.Request, token *authToken.JW3Token) (*types.AccountID, error)
}
