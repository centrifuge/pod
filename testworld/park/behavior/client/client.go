//go:build testworld

package client

import (
	"encoding/json"
	"testing"

	"github.com/centrifuge/pod/testworld/park/behavior/expect"
	"github.com/centrifuge/pod/testworld/park/behavior/webhook"
	"github.com/gavv/httpexpect"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("testworld-client")
)

type Client struct {
	t               *testing.T
	webhookReceiver *webhook.Receiver
	expect          *httpexpect.Expect
	apiURL          string
	jwtToken        string
}

func New(
	t *testing.T,
	webhookReceiver *webhook.Receiver,
	apiURL string,
	jwtToken string,
) *Client {
	expect := expect.CreateInsecureClientWithExpect(t, apiURL)

	return &Client{
		t,
		webhookReceiver,
		expect,
		apiURL,
		jwtToken,
	}
}

func addCommonHeaders(req *httpexpect.Request, auth string) *httpexpect.Request {
	return req.
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		WithHeader("authorization", "bearer "+auth)
}

func GenerateJSONPayload(req any) (map[string]any, error) {
	b, err := json.Marshal(req)

	if err != nil {
		return nil, err
	}

	var payload map[string]any

	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}
