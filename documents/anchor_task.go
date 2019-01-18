package documents

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

const (
	// DocumentIDParam maps to model ID in the kwargs
	DocumentIDParam = "documentID"

	// AccountIDParam maps to account ID in the kwargs
	AccountIDParam = "accountID"

	documentAnchorTaskName = "Document Anchoring"
)

var log = logging.Logger("anchor_task")

type documentAnchorTask struct {
	transactions.BaseTask

	id        []byte
	accountID identity.CentID

	// state
	config        config.Service
	processor     AnchorProcessor
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

	modelID, ok := kwargs[DocumentIDParam].(string)
	if !ok {
		return errors.New("missing model ID")
	}

	d.id, err = hexutil.Decode(modelID)
	if err != nil {
		return errors.New("invalid model ID")
	}

	accountID, ok := kwargs[AccountIDParam].(string)
	if !ok {
		return errors.New("missing account ID")
	}

	d.accountID, err = identity.CentIDFromString(accountID)
	if err != nil {
		return errors.New("invalid cent ID")
	}
	return nil
}

// Copy returns a new task with state.
func (d *documentAnchorTask) Copy() (gocelery.CeleryTask, error) {
	return &documentAnchorTask{
		BaseTask:      transactions.BaseTask{TxService: d.TxService},
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
		err = d.UpdateTransaction(d.accountID, d.TaskTypeName(), err, false)
	}()

	tc, err := d.config.GetAccount(d.accountID[:])
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}
	ctxh, err := contextutil.New(context.Background(), tc)
	if err != nil {
		return false, errors.New("failed to get context header: %v", err)
	}

	model, err := d.modelGetFunc(d.accountID[:], d.id)
	if err != nil {
		return false, errors.New("failed to get model: %v", err)
	}

	if _, err = AnchorDocument(ctxh, model, d.processor, func(id []byte, model Model) error {
		return d.modelSaveFunc(d.accountID[:], id, model)
	}); err != nil {
		return false, errors.New("failed to anchor document: %v", err)
	}

	return true, nil
}

// InitDocumentAnchorTask enqueues a new document anchor task and returns the txID.
func InitDocumentAnchorTask(tq queue.TaskQueuer, txService transactions.Service, accountID identity.CentID, modelID []byte, txID uuid.UUID) (uuid.UUID, error) {
	var tx *transactions.Transaction
	var err error
	if txID != uuid.Nil {
		tx, err = txService.GetTransaction(accountID, txID)
	} else {
		tx, err = txService.CreateTransaction(accountID, documentAnchorTaskName)
	}

	if err != nil {
		return uuid.Nil, err
	}

	params := map[string]interface{}{
		transactions.TxIDParam: tx.ID.String(),
		DocumentIDParam:        hexutil.Encode(modelID),
		AccountIDParam:         accountID.String(),
	}

	_, err = tq.EnqueueJob(documentAnchorTaskName, params)
	if err != nil {
		return uuid.Nil, err
	}

	return tx.ID, nil
}
