// +build unit

package notification

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var cfg config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
	}
	ctx := make(map[string]interface{})
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestWebhookSender_Send(t *testing.T) {
	docID := utils.RandomSlice(32)
	accountID := utils.RandomSlice(identity.DIDLength)
	senderID := utils.RandomSlice(identity.DIDLength)
	statusMsg := "failure"
	message := "some random error"
	var wg sync.WaitGroup
	wg.Add(1)
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(writer http.ResponseWriter, request *http.Request) {
		var resp struct {
			EventType    uint32    `json:"event_type,omitempty"`
			AccountId    string    `json:"account_id,omitempty"`
			FromId       string    `json:"from_id,omitempty"`
			ToId         string    `json:"to_id,omitempty"`
			Recorded     time.Time `json:"recorded,omitempty"`
			DocumentType string    `json:"document_type,omitempty"`
			DocumentId   string    `json:"document_id,omitempty"`
			Status       string    `json:"status,omitempty"`
			Message      string    `json:"message,omitempty"`
		}
		defer request.Body.Close()
		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(data, &resp)
		assert.NoError(t, err)
		writer.Write([]byte("success"))
		assert.Equal(t, hexutil.Encode(docID), resp.DocumentId)
		assert.Equal(t, hexutil.Encode(accountID), resp.AccountId)
		assert.Equal(t, hexutil.Encode(senderID), resp.FromId)
		assert.Equal(t, statusMsg, resp.Status)
		assert.Equal(t, message, resp.Message)
		wg.Done()
	})

	server := &http.Server{Addr: ":8090", Handler: mux}
	go server.ListenAndServe()
	defer server.Close()

	wb := NewWebhookSender()
	notif := Message{
		DocumentID:   hexutil.Encode(docID),
		DocumentType: documenttypes.InvoiceDataTypeUrl,
		AccountID:    hexutil.Encode(accountID),
		FromID:       hexutil.Encode(senderID),
		ToID:         hexutil.Encode(accountID),
		EventType:    ReceivedPayload,
		Recorded:     time.Now().UTC(),
		Status:       statusMsg,
		Message:      message,
	}

	cfg.Set("notifications.endpoint", "http://localhost:8090/webhook")
	status, err := wb.Send(testingconfig.CreateAccountContext(t, cfg), notif)
	assert.NoError(t, err)
	assert.Equal(t, status, Success)
	wg.Wait()
}
