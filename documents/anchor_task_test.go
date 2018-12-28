// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/transactions"

	cc "github.com/centrifuge/go-centrifuge/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestDocumentAnchorTask_ParseKwargs(t *testing.T) {
	tests := []struct {
		name   string
		kwargs map[string]interface{}
		err    string
	}{
		{
			name: "nil kwargs",
			err:  "missing transaction ID",
		},

		{
			kwargs: map[string]interface{}{
				txIDParam: "some string",
			},
			err: "invalid transaction ID",
		},

		// missing model ID
		{
			kwargs: map[string]interface{}{
				txIDParam: uuid.Must(uuid.NewV4()).String(),
			},
			err: "missing model ID",
		},

		// missing tenantID
		{
			kwargs: map[string]interface{}{
				txIDParam:    uuid.Must(uuid.NewV4()).String(),
				modelIDParam: hexutil.Encode(utils.RandomSlice(32)),
			},

			err: "missing tenant ID",
		},

		// all good
		{
			name: "success",
			kwargs: map[string]interface{}{
				txIDParam:     uuid.Must(uuid.NewV4()).String(),
				modelIDParam:  hexutil.Encode(utils.RandomSlice(32)),
				tenantIDParam: cc.DummyIdentity,
			},
		},
	}

	for _, c := range tests {
		name := c.name
		if name == "" {
			name = c.err
		}

		t.Run(name, func(t *testing.T) {
			d, err := utils.SimulateJSONDecodeForGocelery(c.kwargs)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			task := new(documentAnchorTask)
			err = task.ParseKwargs(d)
			if c.err == "" {
				assert.Equal(t, task.txID.String(), c.kwargs[txIDParam])
				assert.Equal(t, hexutil.Encode(task.id), c.kwargs[modelIDParam])
				assert.Equal(t, task.tenantID, c.kwargs[tenantIDParam])
				return
			}

			assert.EqualError(t, err, c.err)
		})
	}
}

func TestDocumentAnchorTask_updateTransaction(t *testing.T) {
	task := new(documentAnchorTask)
	task.tenantID = cc.DummyIdentity
	task.id = utils.RandomSlice(32)
	task.txID = uuid.Must(uuid.NewV4())
	task.txRepository = ctx[transactions.BootstrappedRepo].(transactions.Repository)

	// missing transaction with nil error
	err := task.updateTransaction(nil)
	err = errors.GetErrs(err)[0]
	assert.True(t, errors.IsOfType(transactions.ErrTransactionMissing, err))

	// missing transaction with error
	err = task.updateTransaction(errors.New("anchor error"))
	err = errors.GetErrs(err)[1]
	assert.True(t, errors.IsOfType(transactions.ErrTransactionMissing, err))

	// no error and success
	tx := transactions.NewTransaction(task.tenantID, "")
	assert.NoError(t, task.txRepository.Save(tx))
	task.txID = tx.ID
	assert.NoError(t, task.updateTransaction(nil))
	tx, err = task.txRepository.Get(task.tenantID, task.txID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, transactions.Success)
	assert.Len(t, tx.Logs, 1)

	// failed task
	tx = transactions.NewTransaction(task.tenantID, "")
	assert.NoError(t, task.txRepository.Save(tx))
	task.txID = tx.ID
	err = task.updateTransaction(errors.New("anchor error"))
	assert.EqualError(t, errors.GetErrs(err)[0], "anchor error")
	tx, err = task.txRepository.Get(task.tenantID, task.txID)
	assert.NoError(t, err)
	assert.Equal(t, tx.Status, transactions.Failed)
	assert.Len(t, tx.Logs, 1)
}
