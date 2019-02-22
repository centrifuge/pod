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
	"github.com/centrifuge/go-centrifuge/transactions/txv1"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
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
	txv1.BaseTask

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
	err := d.ParseTransactionID(d.TaskTypeName(), kwargs)
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
		BaseTask:      txv1.BaseTask{TxManager: d.TxManager},
		config:        d.config,
		processor:     d.processor,
		modelGetFunc:  d.modelGetFunc,
		modelSaveFunc: d.modelSaveFunc,
	}, nil
}

// RunTask anchors the Document.
func (d *documentAnchorTask) RunTask() (res interface{}, err error) {
	log.Infof("starting anchor task for transaction: %s\n", d.TxID)
	defer func() {
		err = d.UpdateTransaction(d.accountID, d.TaskTypeName(), err)
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
		return false, errors.New("failed to anchor Document: %v", err)
	}

	return true, nil
}

// InitDocumentAnchorTask enqueues a new Document anchor task for a given combination of accountID/modelID/txID.
func InitDocumentAnchorTask(txMan transactions.Manager, tq queue.TaskQueuer, accountID identity.CentID, modelID []byte, txID transactions.TxID) (queue.TaskResult, error) {
	params := map[string]interface{}{
		transactions.TxIDParam: txID.String(),
		DocumentIDParam:        hexutil.Encode(modelID),
		AccountIDParam:         accountID.String(),
	}

	err := txMan.UpdateTaskStatus(accountID, txID, transactions.Pending, documentAnchorTaskName, "init")
	if err != nil {
		return nil, err
	}

	tr, err := tq.EnqueueJob(documentAnchorTaskName, params)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// CreateAnchorTransaction creates a transaction for anchoring a Document using transaction manager
func CreateAnchorTransaction(txMan transactions.Manager, tq queue.TaskQueuer, self identity.CentID, txID transactions.TxID, documentID []byte) (transactions.TxID, chan bool, error) {
	txID, done, err := txMan.ExecuteWithinTX(context.Background(), self, txID, "anchor Document", func(accountID identity.CentID, TID transactions.TxID, txMan transactions.Manager, errChan chan<- error) {
		tr, err := InitDocumentAnchorTask(txMan, tq, accountID, documentID, TID)
		if err != nil {
			errChan <- err
			return
		}
		_, err = tr.Get(txMan.GetDefaultTaskTimeout())
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	})
	return txID, done, err
}
