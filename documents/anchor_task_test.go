// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/transactions"
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
				transactions.TxIDParam: "some string",
			},
			err: "invalid transaction ID",
		},

		// missing model ID
		{
			kwargs: map[string]interface{}{
				transactions.TxIDParam: uuid.Must(uuid.NewV4()).String(),
			},
			err: "missing model ID",
		},

		// missing accountID
		{
			kwargs: map[string]interface{}{
				transactions.TxIDParam: uuid.Must(uuid.NewV4()).String(),
				modelIDParam:           hexutil.Encode(utils.RandomSlice(32)),
			},

			err: "missing account ID",
		},

		// all good
		{
			name: "success",
			kwargs: map[string]interface{}{
				transactions.TxIDParam: uuid.Must(uuid.NewV4()).String(),
				modelIDParam:           hexutil.Encode(utils.RandomSlice(32)),
				accountIDParam:         identity.RandomCentID().String(),
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
				assert.Equal(t, task.TxID.String(), c.kwargs[transactions.TxIDParam])
				assert.Equal(t, hexutil.Encode(task.id), c.kwargs[modelIDParam])
				assert.Equal(t, task.accountID.String(), c.kwargs[accountIDParam])
				return
			}

			assert.EqualError(t, err, c.err)
		})
	}
}
