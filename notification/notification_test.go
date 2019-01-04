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
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, nil)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

type mockConfig struct {
	url string
}

func (m mockConfig) GetReceiveEventNotificationEndpoint() string {
	return m.url
}

func TestWebhookSender_Send(t *testing.T) {
	docID := utils.RandomSlice(32)
	cid := utils.RandomSlice(32)
	ts, err := ptypes.TimestampProto(time.Now().UTC())
	assert.Nil(t, err, "Should not error out")
	var wg sync.WaitGroup
	wg.Add(1)
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(writer http.ResponseWriter, request *http.Request) {
		var resp struct {
			EventType    uint32               `json:"event_type,omitempty"`
			CentrifugeId string               `json:"centrifuge_id,omitempty"`
			Recorded     *timestamp.Timestamp `json:"recorded,omitempty"`
			DocumentType string               `json:"document_type,omitempty"`
			DocumentId   string               `json:"document_id,omitempty"`
		}
		defer request.Body.Close()
		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(data, &resp)
		assert.NoError(t, err)
		writer.Write([]byte("success"))
		assert.Equal(t, hexutil.Encode(docID), resp.DocumentId)
		assert.Equal(t, hexutil.Encode(cid), resp.CentrifugeId)
		wg.Done()
	})

	server := &http.Server{Addr: ":8090", Handler: mux}
	go server.ListenAndServe()
	defer server.Close()

	wb := NewWebhookSender(mockConfig{url: "http://localhost:8090/webhook"})
	notif := &notificationpb.NotificationMessage{
		DocumentId:   hexutil.Encode(docID),
		DocumentType: documenttypes.InvoiceDataTypeUrl,
		CentrifugeId: hexutil.Encode(cid),
		EventType:    uint32(ReceivedPayload),
		Recorded:     ts,
	}

	status, err := wb.Send(notif)
	assert.NoError(t, err)
	assert.Equal(t, status, Success)
	wg.Wait()
}
