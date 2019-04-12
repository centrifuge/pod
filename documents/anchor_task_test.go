// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
			err:  "missing job ID",
		},

		{
			kwargs: map[string]interface{}{
				jobs.JobIDParam: "some string",
			},
			err: "invalid job ID",
		},

		// missing model ID
		{
			kwargs: map[string]interface{}{
				jobs.JobIDParam: jobs.NewJobID().String(),
			},
			err: "missing model ID",
		},

		// missing accountID
		{
			kwargs: map[string]interface{}{
				jobs.JobIDParam: jobs.NewJobID().String(),
				DocumentIDParam: hexutil.Encode(utils.RandomSlice(32)),
			},

			err: "missing account ID",
		},

		// all good
		{
			name: "success",
			kwargs: map[string]interface{}{
				jobs.JobIDParam: jobs.NewJobID().String(),
				DocumentIDParam: hexutil.Encode(utils.RandomSlice(32)),
				AccountIDParam:  testingidentity.GenerateRandomDID().String(),
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
				assert.Equal(t, task.JobID.String(), c.kwargs[jobs.JobIDParam])
				assert.Equal(t, hexutil.Encode(task.id), c.kwargs[DocumentIDParam])
				assert.Equal(t, task.accountID.String(), c.kwargs[AccountIDParam])
				return
			}

			assert.EqualError(t, err, c.err)
		})
	}
}
