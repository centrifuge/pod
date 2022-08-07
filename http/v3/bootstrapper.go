package v3

import (
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
)

// BootstrappedService key maps to the Service implementation in Bootstrap context.
const BootstrappedService = "V3 Service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b *Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	nftSrvV3 := ctx[nftv3.BootstrappedNFTV3Service].(nftv3.Service)
	ctx[BootstrappedService] = Service{
		nftSrvV3: nftSrvV3,
	}
	return nil
}
