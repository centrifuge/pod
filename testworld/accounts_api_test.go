//go:build testworld

package testworld

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/testworld/park/behavior/client"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func TestAccountsAPI_GenerateAccount(t *testing.T) {
	tests := []struct {
		Name               string
		ProxyType          string
		ExpectedStatusCode int
		ExpectedAccount    bool
	}{
		{
			Name:               "Invalid proxy type 1",
			ProxyType:          proxyType.ProxyTypeName[proxyType.PodAuth],
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedAccount:    false,
		},
		{
			Name:               "Invalid proxy type 2",
			ProxyType:          proxyType.ProxyTypeName[proxyType.PodOperation],
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedAccount:    false,
		},
		{
			Name:               "Valid proxy type",
			ProxyType:          auth.PodAdminProxyType,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedAccount:    true,
		},
	}

	testIdentity, err := types.NewAccountID(keyrings.EveKeyRingPair.PublicKey)
	assert.NoError(t, err)

	for hostName, host := range controller.GetHosts() {
		for _, test := range tests {
			testName := fmt.Sprintf("%s for host %s", test.Name, hostName)

			t.Run(testName, func(t *testing.T) {
				authToken, err := host.GetMainAccount().GetJW3Token(test.ProxyType)
				assert.NoError(t, err)

				c := client.New(t, controller.GetWebhookReceiver(), host.GetAPIURL(), authToken)

				webhookURL := fmt.Sprintf("https://centrifuge.io/testworld/%s", testName)

				reqPayload, err := client.GenerateJSONPayload(coreapi.GenerateAccountPayload{
					Account: coreapi.Account{
						Identity:         testIdentity,
						WebhookURL:       webhookURL,
						PrecommitEnabled: true,
					},
				})
				assert.NoError(t, err)

				res := c.GenerateAccount(reqPayload, test.ExpectedStatusCode)

				assert.NotNil(t, res)

				if !test.ExpectedAccount {
					return
				}

				var acc coreapi.Account

				err = json.Unmarshal([]byte(res.Body().Raw()), &acc)
				assert.NoError(t, err)

				assert.True(t, testIdentity.Equal(acc.Identity))
				assert.True(t, host.GetMainAccount().GetPodOperatorAccountID().Equal(acc.PodOperatorAccountID))
				assert.NotEmpty(t, acc.DocumentSigningPublicKey.Bytes())
				assert.Equal(t, host.GetMainAccount().GetP2PPublicKey(), acc.P2PPublicSigningKey.Bytes())
				assert.Equal(t, webhookURL, acc.WebhookURL)
				assert.True(t, acc.PrecommitEnabled)

				err = genericUtils.GetService[config.Service](host.GetServiceCtx()).
					DeleteAccount(testIdentity.ToBytes())
				assert.NoError(t, err)
			})
		}
	}
}

func TestAccountsAPI_SignPayload(t *testing.T) {
	t.Parallel()

	testPayload := utils.RandomSlice(32)

	for hostName, host := range controller.GetHosts() {
		t.Run(string(hostName), func(t *testing.T) {
			hostClient, err := controller.GetClientForHost(t, hostName)
			assert.NoError(t, err)

			reqPayload, err := client.GenerateJSONPayload(coreapi.SignRequest{Payload: testPayload})
			assert.NoError(t, err)

			res := hostClient.SignPayload(host.GetMainAccount().GetAccountID(), reqPayload, http.StatusOK)

			var signResponse coreapi.SignResponse

			err = json.Unmarshal([]byte(res.Body().Raw()), &signResponse)
			assert.NoError(t, err)

			assert.Equal(t, testPayload, signResponse.Payload.Bytes())
			assert.Equal(t, host.GetMainAccount().GetAccount().GetSigningPublicKey(), signResponse.PublicKey.Bytes())
			assert.Equal(t, host.GetMainAccount().GetAccount().GetIdentity().ToBytes(), signResponse.SignerID.Bytes())

			validSignature := crypto.VerifyMessage(
				signResponse.PublicKey.Bytes(),
				testPayload,
				signResponse.Signature.Bytes(),
				crypto.CurveEd25519,
			)
			assert.True(t, validSignature)
		})
	}
}

func TestAccountsAPI_GetSelf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name               string
		ProxyType          string
		ExpectedStatusCode int
		ExpectedAccount    bool
	}{
		{
			Name:               "Valid proxy type 1",
			ProxyType:          proxyType.ProxyTypeName[proxyType.PodAuth],
			ExpectedStatusCode: http.StatusOK,
			ExpectedAccount:    true,
		},
		{
			Name:               "Valid proxy type 2",
			ProxyType:          proxyType.ProxyTypeName[proxyType.PodOperation],
			ExpectedStatusCode: http.StatusOK,
			ExpectedAccount:    true,
		},
		{
			Name:               "Invalid proxy type",
			ProxyType:          auth.PodAdminProxyType,
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedAccount:    false,
		},
	}

	for hostName, host := range controller.GetHosts() {
		for _, test := range tests {
			testName := fmt.Sprintf("%s for host %s", test.Name, hostName)

			t.Run(testName, func(t *testing.T) {
				authToken, err := host.GetMainAccount().GetJW3Token(test.ProxyType)
				assert.NoError(t, err)

				c := client.New(t, controller.GetWebhookReceiver(), host.GetAPIURL(), authToken)

				res := c.GetSelf(test.ExpectedStatusCode)

				assert.NotNil(t, res)

				if !test.ExpectedAccount {
					return
				}

				var acc coreapi.Account

				err = json.Unmarshal([]byte(res.Body().Raw()), &acc)
				assert.NoError(t, err)

				assert.True(t, host.GetMainAccount().GetAccountID().Equal(acc.Identity))
				assert.True(t, host.GetMainAccount().GetPodOperatorAccountID().Equal(acc.PodOperatorAccountID))
				assert.NotEmpty(t, acc.DocumentSigningPublicKey.Bytes())
				assert.Equal(t, host.GetMainAccount().GetP2PPublicKey(), acc.P2PPublicSigningKey.Bytes())
			})
		}
	}
}

func TestAccountsAPI_GetAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name               string
		ProxyType          string
		ExpectedStatusCode int
		ExpectedAccount    bool
	}{
		{
			Name:               "Invalid proxy type 1",
			ProxyType:          proxyType.ProxyTypeName[proxyType.PodAuth],
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedAccount:    false,
		},
		{
			Name:               "Invalid proxy type 2",
			ProxyType:          proxyType.ProxyTypeName[proxyType.PodOperation],
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedAccount:    false,
		},
		{
			Name:               "Valid proxy type",
			ProxyType:          auth.PodAdminProxyType,
			ExpectedStatusCode: http.StatusOK,
			ExpectedAccount:    true,
		},
	}

	for hostName, host := range controller.GetHosts() {
		for _, test := range tests {
			testName := fmt.Sprintf("%s for host %s", test.Name, hostName)

			t.Run(testName, func(t *testing.T) {
				authToken, err := host.GetMainAccount().GetJW3Token(test.ProxyType)
				assert.NoError(t, err)

				c := client.New(t, controller.GetWebhookReceiver(), host.GetAPIURL(), authToken)

				res := c.GetAccount(test.ExpectedStatusCode, host.GetMainAccount().GetAccountID().ToHexString())

				assert.NotNil(t, res)

				if !test.ExpectedAccount {
					return
				}

				var acc coreapi.Account

				b, err := json.Marshal(res.Raw())
				assert.NoError(t, err)

				err = json.Unmarshal(b, &acc)
				assert.NoError(t, err)

				assert.True(t, host.GetMainAccount().GetAccountID().Equal(acc.Identity))
				assert.True(t, host.GetMainAccount().GetPodOperatorAccountID().Equal(acc.PodOperatorAccountID))
				assert.NotEmpty(t, acc.DocumentSigningPublicKey.Bytes())
				assert.Equal(t, host.GetMainAccount().GetP2PPublicKey(), acc.P2PPublicSigningKey.Bytes())
			})
		}
	}
}

func TestAccountsAPI_GetAccounts(t *testing.T) {
	tests := []struct {
		Name               string
		ProxyType          string
		ExpectedStatusCode int
		ExpectedAccounts   bool
	}{
		{
			Name:               "Invalid proxy type 1",
			ProxyType:          proxyType.ProxyTypeName[proxyType.PodAuth],
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedAccounts:   false,
		},
		{
			Name:               "Invalid proxy type 2",
			ProxyType:          proxyType.ProxyTypeName[proxyType.PodOperation],
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedAccounts:   false,
		},
		{
			Name:               "Valid proxy type",
			ProxyType:          auth.PodAdminProxyType,
			ExpectedStatusCode: http.StatusOK,
			ExpectedAccounts:   true,
		},
	}

	for hostName, host := range controller.GetHosts() {
		for _, test := range tests {
			testName := fmt.Sprintf("%s for host %s", test.Name, hostName)

			t.Run(testName, func(t *testing.T) {
				authToken, err := host.GetMainAccount().GetJW3Token(test.ProxyType)
				assert.NoError(t, err)

				c := client.New(t, controller.GetWebhookReceiver(), host.GetAPIURL(), authToken)

				res := c.GetAllAccounts(test.ExpectedStatusCode)

				assert.NotNil(t, res)

				if !test.ExpectedAccounts {
					return
				}

				var accounts coreapi.Accounts

				b, err := json.Marshal(res.Raw())
				assert.NoError(t, err)

				err = json.Unmarshal(b, &accounts)
				assert.NoError(t, err)

				assert.NotEmpty(t, accounts.Data)
			})
		}
	}
}
