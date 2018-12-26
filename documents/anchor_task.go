package documents

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

const (
	txIDParam     = "transactionID"
	modelIDParam  = "modelID"
	tenantIDParam = "tenantID"
)

var log = logging.Logger("nft")

type documentAnchorTask struct {
	txID         uuid.UUID
	id, tenantID []byte

	txRepository  transactions.Repository
	config        config.Configuration
	processor     anchorProcessor
	modelGetFunc  func(tenantID, id []byte) (Model, error)
	modelSaveFunc func(tenantID, id []byte, model Model) error
}

func (d *documentAnchorTask) ParseKwargs(kwargs map[string]interface{}) error {
	txID, ok := kwargs[txIDParam].(string)
	if !ok {
		return errors.New("missing transaction ID")
	}

	d.txID = uuid.Must(uuid.FromString(txID))
	if d.id, ok = kwargs[modelIDParam].([]byte); !ok {
		return errors.New("missing model ID")
	}

	if d.tenantID, ok = kwargs[tenantIDParam].([]byte); !ok {
		return errors.New("missing tenant ID")
	}

	return nil
}

func (d *documentAnchorTask) Copy() (gocelery.CeleryTask, error) {
	return &documentAnchorTask{
		txRepository:  d.txRepository,
		config:        d.config,
		processor:     d.processor,
		modelGetFunc:  d.modelGetFunc,
		modelSaveFunc: d.modelSaveFunc,
	}, nil
}

func (d *documentAnchorTask) RunTask() (res interface{}, err error) {
	log.Infof("starting anchor task: %v\n", d.txID.String())
	defer func() {
		if err != nil {
			log.Infof("anchor task failed: %v\n", err)
			return
		}

		log.Infof("anchor task successful: %v\n", d.txID.String())
	}()

	ctxh, err := header.NewContextHeader(context.Background(), d.config)
	if err != nil {
		return false, errors.New("failed to get context header: %v", err)
	}

	model, err := d.modelGetFunc(d.tenantID, d.id)
	if err != nil {
		return false, errors.New("failed to get model: %v", err)
	}

	if _, err = AnchorDocument(ctxh, model, d.processor, func(id []byte, model Model) error {
		return d.modelSaveFunc(d.tenantID, id, model)
	}); err != nil {
		return false, errors.New("failed to anchor document: %v", err)
	}

	return true, nil
}
