package v2

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
)

type CreateIdentityRequest struct {
	Identity         identity.DID
	WebhookURL       string
	PrecommitEnabled bool
	AccountProxies   config.AccountProxies
}

type CreateIdentityResponse struct {
	JobID    string
	Identity identity.DID
}
