//go:build unit

package notification

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func sendAndVerify(t *testing.T, message Message) {
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var resp Message
		defer request.Body.Close()
		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(data, &resp)
		assert.NoError(t, err)
		writer.WriteHeader(http.StatusOK)
		assert.Equal(t, message.EventType, resp.EventType)
		if message.EventType == EventTypeJob {
			assert.Equal(t, *message.Job, *resp.Job)
			assert.Nil(t, resp.Document)
		} else {
			assert.Equal(t, *message.Document, *resp.Document)
			assert.Nil(t, resp.Job)
		}
	}))

	defer testServer.Close()

	wb := NewWebhookSender()

	acc := config.NewAccountMock(t)
	acc.On("GetWebhookURL").Return(testServer.URL).Once()

	ctx := contextutil.WithAccount(context.Background(), acc)

	err := wb.Send(ctx, message)
	assert.NoError(t, err)
}

func TestNewWebhookSender(t *testing.T) {
	msgs := []Message{
		{
			EventType:  EventTypeJob,
			RecordedAt: time.Now().UTC(),
			Job: &JobMessage{
				ID:         utils.RandomSlice(32),
				Owner:      utils.RandomSlice(20),
				Desc:       "Sample Job",
				ValidUntil: time.Now().Add(time.Hour).UTC(),
				FinishedAt: time.Now().UTC(),
			},
		},
		{
			EventType:  EventTypeDocument,
			RecordedAt: time.Now().UTC(),
			Document: &DocumentMessage{
				ID:        utils.RandomSlice(32),
				VersionID: utils.RandomSlice(32),
				From:      utils.RandomSlice(20),
				To:        utils.RandomSlice(20),
			},
		},
	}

	for _, msg := range msgs {
		sendAndVerify(t, msg)
	}
}
