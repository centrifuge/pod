package v3

import (
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	nftv3 "github.com/centrifuge/pod/nft/v3"
)

// BootstrappedService key maps to the Service implementation in Bootstrap context.
const BootstrappedService = "V3 Service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

func (b *Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	nftSrvV3, ok := ctx[nftv3.BootstrappedNFTV3Service].(nftv3.Service)

	if !ok {
		return errors.New("nft V3 service not initialised")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)

	if !ok {
		return errors.New("document service not initialised")
	}

	ctx[BootstrappedService] = &Service{
		docSrv:   docSrv,
		nftSrvV3: nftSrvV3,
	}

	return nil
}
