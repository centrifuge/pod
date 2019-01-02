package documents

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

const (
	modelIDParam           = "modelID"
	tenantIDParam          = "tenantID"
	documentAnchorTaskName = "Document Anchoring"
)

var log = logging.Logger("anchor_task")

type documentAnchorTask struct {
	transactions.BaseTask

	id       []byte
	tenantID common.Address

	// state
	config        config.Configuration
	processor     anchorProcessor
	modelGetFunc  func(tenantID, id []byte) (Model, error)
	modelSaveFunc func(tenantID, id []byte, model Model) error
}

// TaskTypeName returns the name of the task.
func (d *documentAnchorTask) TaskTypeName() string {
	return documentAnchorTaskName
}

// ParseKwargs parses the kwargs.
func (d *documentAnchorTask) ParseKwargs(kwargs map[string]interface{}) error {
	err := d.ParseTransactionID(kwargs)
	if err != nil {
		return err
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

// Copy returns a new task with state.
func (d *documentAnchorTask) Copy() (gocelery.CeleryTask, error) {
	return &documentAnchorTask{
		BaseTask:      transactions.BaseTask{TxRepository: d.TxRepository},
		config:        d.config,
		processor:     d.processor,
		modelGetFunc:  d.modelGetFunc,
		modelSaveFunc: d.modelSaveFunc,
	}, nil
}

// RunTask anchors the document.
func (d *documentAnchorTask) RunTask() (res interface{}, err error) {
	log.Infof("starting anchor task: %v\n", d.TxID.String())
	defer func() {
		if err == nil {
			log.Infof("anchor task successful: %v\n", d.TxID.String())
		} else {
			log.Infof("anchor task failed: %v\n", err)
		}

		err = d.UpdateTransaction(d.tenantID, d.TaskTypeName(), err)
	}()

	ctxh, err := contextutil.NewCentrifugeContext(context.Background(), d.config)
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

// taskQueuer can be implemented by any queueing system
type taskQueuer interface {
	EnqueueJob(taskTypeName string, params map[string]interface{}) (queue.TaskResult, error)
}

// InitDocumentAnchorTask enqueues a new document anchor task and returns the txID.
func InitDocumentAnchorTask(tq taskQueuer, txRepo transactions.Repository, tenantID common.Address, modelID []byte) (uuid.UUID, error) {
	tx := transactions.NewTransaction(tenantID, documentAnchorTaskName)
	err := txRepo.Save(tx)
	if err != nil {
		return uuid.Nil, err
	}

	params := map[string]interface{}{
		transactions.TxIDParam: tx.ID.String(),
		modelIDParam:           hexutil.Encode(modelID),
		tenantIDParam:          tenantID,
	}

	_, err = tq.EnqueueJob(documentAnchorTaskName, params)
	if err != nil {
		return uuid.Nil, err
	}

	return tx.ID, nil
}
