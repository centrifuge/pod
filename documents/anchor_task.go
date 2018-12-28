package documents

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

const (
	txIDParam              = "transactionID"
	modelIDParam           = "modelID"
	tenantIDParam          = "tenantID"
	documentAnchorTaskName = "Document Anchoring"
)

var log = logging.Logger("nft")

type documentAnchorTask struct {
	txID     uuid.UUID
	id       []byte
	tenantID common.Address

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

	var err error
	d.txID, err = uuid.FromString(txID)
	if err != nil {
		return errors.New("invalid transaction ID")
	}

	modelID, ok := kwargs[modelIDParam].(string)
	if !ok {
		return errors.New("missing model ID")
	}

	d.id, err = hexutil.Decode(modelID)
	if err != nil {
		return errors.New("invalid model ID")
	}

	tenantID, ok := kwargs[tenantIDParam].(string)
	if !ok {
		return errors.New("missing tenant ID")
	}

	d.tenantID = common.HexToAddress(tenantID)
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
		tx, erri := d.txRepository.Get(d.tenantID, d.txID)
		if erri != nil {
			log.Infof("transaction not found: %v\n", erri)
			return
		}

		if err == nil {
			log.Infof("anchor task successful: %v\n", d.txID.String())
			tx.Status = transactions.Success
			tx.Logs = append(tx.Logs, transactions.NewLog(documentAnchorTaskName, ""))
			err = d.txRepository.Save(tx)
			return
		}

		log.Infof("anchor task failed: %v\n", err)
		tx.Status = transactions.Failed
		tx.Logs = append(tx.Logs, transactions.NewLog(documentAnchorTaskName, err.Error()))
		erri = d.txRepository.Save(tx)
		if erri != nil {
			err = errors.AppendError(err, erri)
		}
	}()

	ctxh, err := header.NewContextHeader(context.Background(), d.config)
	if err != nil {
		return false, errors.New("failed to get context header: %v", err)
	}

	model, err := d.modelGetFunc(d.tenantID[:], d.id)
	if err != nil {
		return false, errors.New("failed to get model: %v", err)
	}

	if _, err = AnchorDocument(ctxh, model, d.processor, func(id []byte, model Model) error {
		return d.modelSaveFunc(d.tenantID[:], id, model)
	}); err != nil {
		return false, errors.New("failed to anchor document: %v", err)
	}

	return true, nil
}
