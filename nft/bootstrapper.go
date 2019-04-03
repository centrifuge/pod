package nft

import (
	"context"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
)

const (
	// BootstrappedInvoiceUnpaid is the key to InvoiceUnpaid NFT in bootstrap context.
	BootstrappedInvoiceUnpaid = "BootstrappedInvoiceUnpaid"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initializes the invoice unpaid contract
func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, err := configstore.RetrieveConfig(false, ctx)
	if err != nil {
		return err
	}

	if _, ok := ctx[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialised")
	}

	idService, ok := ctx[identity.BootstrappedDIDService].(identity.ServiceDID)
	if !ok {
		return errors.New("identity service not initialised")
	}

	if _, ok := ctx[bootstrap.BootstrappedQueueServer]; !ok {
		return errors.New("queue hasn't been initialized")
	}
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)

	txManager, ok := ctx[transactions.BootstrappedService].(transactions.Manager)
	if !ok {
		return errors.New("transactions repository not initialised")
	}

	client := ethereum.GetClient()
	InvoiceUnpaid := newEthInvoiceUnpaid(
		cfg,
		idService,
		client,
		queueSrv,
		docSrv,
		bindContract,
		txManager, func() (uint64, error) {
			h, err := client.GetEthClient().HeaderByNumber(context.Background(), nil)
			if err != nil {
				return 0, err
			}

			return h.Number.Uint64(), nil
		})
	ctx[BootstrappedInvoiceUnpaid] = InvoiceUnpaid
	return nil
}
