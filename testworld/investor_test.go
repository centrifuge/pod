//go:build testworld

package testworld

import (
	"testing"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/pod/testworld/park/behavior/client"
	"github.com/centrifuge/pod/testworld/park/host"
	"github.com/stretchr/testify/assert"
)

func TestInvestorAPI_GetAsset(t *testing.T) {
	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)

	payload := genericCoreAPICreate(nil)

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// Use a PodAuth token and client since we are going to create all the pool related info for this identity.
	authToken, err := alice.GetMainAccount().GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)

	testClient := client.New(t, controller.GetWebhookReceiver(), alice.GetAPIURL(), authToken)

	_, _ = docID, testClient
}
