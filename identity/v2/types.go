package v2

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type CreateIdentityRequest struct {
	Identity         *types.AccountID
	WebhookURL       string
	PrecommitEnabled bool
}
