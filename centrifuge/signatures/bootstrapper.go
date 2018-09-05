package signatures

import "github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	NewSigningService(SigningService{IdentityService: &identity.EthereumIdentityService{}})
	return nil
}
