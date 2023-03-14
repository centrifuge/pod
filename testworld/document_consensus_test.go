//go:build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/pod/testworld/park/behavior/client"
	"github.com/centrifuge/pod/testworld/park/host"
	"github.com/stretchr/testify/assert"
)

func TestDocumentsAPI_AddCollaborator_WithinHost(t *testing.T) {
	webhookReceiver := controller.GetWebhookReceiver()

	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	hostAccount1, err := controller.CreateRandomAccountOnHost(host.Bob)
	assert.NoError(t, err)
	hostAccount2, err := controller.CreateRandomAccountOnHost(host.Bob)
	assert.NoError(t, err)
	hostAccount3, err := controller.CreateRandomAccountOnHost(host.Bob)
	assert.NoError(t, err)

	hostAccount1JWT, err := hostAccount1.GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)
	hostAccount2JWT, err := hostAccount2.GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)
	hostAccount3JWT, err := hostAccount3.GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)

	hostAccount1Client := client.New(t, webhookReceiver, bob.GetAPIURL(), hostAccount1JWT)
	hostAccount2Client := client.New(t, webhookReceiver, bob.GetAPIURL(), hostAccount2JWT)
	hostAccount3Client := client.New(t, webhookReceiver, bob.GetAPIURL(), hostAccount3JWT)

	// Account 1 shares document with Account 2 first
	payload := genericCoreAPICreate([]string{hostAccount2.GetAccountID().ToHexString()})

	docID, err := hostAccount1Client.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	hostAccount1Client.GetDocumentAndVerify(docID, nil, createAttributes())
	hostAccount2Client.GetDocumentAndVerify(docID, nil, createAttributes())

	msg, err := webhookReceiver.GetReceivedDocumentMsg(hostAccount2.GetAccountID().ToHexString(), docID)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(hostAccount1.GetAccountID().ToHexString()), strings.ToLower(msg.Document.From.String()))

	hostAccount3Client.NonExistingDocumentCheck(docID)

	// Account 2 updates invoice and shares with Account 3 as well
	payload = genericCoreAPIUpdate(
		[]string{
			hostAccount1.GetAccountID().ToHexString(),
			hostAccount3.GetAccountID().ToHexString(),
		},
	)
	payload["document_id"] = docID

	docID, err = hostAccount2Client.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	hostAccount1Client.GetDocumentAndVerify(docID, nil, allAttributes())
	hostAccount2Client.GetDocumentAndVerify(docID, nil, allAttributes())
	hostAccount3Client.GetDocumentAndVerify(docID, nil, allAttributes())

	msg, err = webhookReceiver.GetReceivedDocumentMsg(hostAccount3.GetAccountID().ToHexString(), docID)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(hostAccount2.GetAccountID().ToHexString()), strings.ToLower(msg.Document.From.String()))
}

func TestDocumentsAPI_AddCollaborator_MultiHost(t *testing.T) {
	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)
	charlie, err := controller.GetHost(host.Charlie)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)
	charlieClient, err := controller.GetClientForHost(t, host.Charlie)
	assert.NoError(t, err)

	// Alice creates and shares a document with Bob.
	payload := genericCoreAPICreate([]string{
		bob.GetMainAccount().GetAccountID().ToHexString(),
	})

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	aliceClient.GetDocumentAndVerify(docID, nil, createAttributes())
	bobClient.GetDocumentAndVerify(docID, nil, createAttributes())
	charlieClient.NonExistingDocumentCheck(docID)

	// Bob updates the document and adds Charlie as collaborator.
	payload = genericCoreAPIUpdate([]string{
		alice.GetMainAccount().GetAccountID().ToHexString(),
		charlie.GetMainAccount().GetAccountID().ToHexString(),
	})

	payload["document_id"] = docID

	docID, err = bobClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// All accounts have the document now.
	aliceClient.GetDocumentAndVerify(docID, nil, allAttributes())
	bobClient.GetDocumentAndVerify(docID, nil, allAttributes())
	charlieClient.GetDocumentAndVerify(docID, nil, allAttributes())

}

func TestDocumentsAPI_AddCollaborator_MultiHostMultiAccount(t *testing.T) {
	webhookReceiver := controller.GetWebhookReceiver()

	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)
	charlie, err := controller.GetHost(host.Charlie)
	assert.NoError(t, err)

	aliceJWT, err := alice.GetMainAccount().GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)

	// hostAccount1 is created in Bob's host
	hostAccount1, err := controller.CreateRandomAccountOnHost(host.Bob)
	assert.NoError(t, err)
	// hostAccount2 is created in Charlie's host
	hostAccount2, err := controller.CreateRandomAccountOnHost(host.Charlie)
	assert.NoError(t, err)

	hostAccount1JWT, err := hostAccount1.GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)
	hostAccount2JWT, err := hostAccount2.GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)

	aliceClient := client.New(t, webhookReceiver, alice.GetAPIURL(), aliceJWT)
	hostAccount1Client := client.New(t, webhookReceiver, bob.GetAPIURL(), hostAccount1JWT)
	hostAccount2Client := client.New(t, webhookReceiver, charlie.GetAPIURL(), hostAccount2JWT)

	// Alice creates and shares document with Account 1.
	payload := genericCoreAPICreate([]string{
		hostAccount1.GetAccountID().ToHexString(),
	})

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	aliceClient.GetDocumentAndVerify(docID, nil, createAttributes())
	hostAccount1Client.GetDocumentAndVerify(docID, nil, createAttributes())

	// Confirm that Account 1 received the document.
	msg, err := webhookReceiver.GetReceivedDocumentMsg(hostAccount1.GetAccountID().ToHexString(), docID)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(alice.GetMainAccount().GetAccountID().ToHexString()), strings.ToLower(msg.Document.From.String()))

	// Confirm that Account 2 did not receive the document.
	hostAccount2Client.NonExistingDocumentCheck(docID)

	// Account 1 updates the document and shares it with Alice and Account 2.
	payload = genericCoreAPIUpdate([]string{
		alice.GetMainAccount().GetAccountID().ToHexString(),
		hostAccount2.GetAccountID().ToHexString(),
	})

	payload["document_id"] = docID

	docID, err = hostAccount1Client.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// All accounts have the document now.
	aliceClient.GetDocumentAndVerify(docID, nil, allAttributes())
	hostAccount1Client.GetDocumentAndVerify(docID, nil, allAttributes())
	hostAccount2Client.GetDocumentAndVerify(docID, nil, allAttributes())
}

func TestDocumentsAPI_CollaboratorTimeout(t *testing.T) {
	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)

	// Alice creates and shares a document with Bob.
	payload := genericCoreAPICreate([]string{
		bob.GetMainAccount().GetAccountID().ToHexString(),
	})

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	aliceClient.GetDocumentAndVerify(docID, nil, createAttributes())
	bobClient.GetDocumentAndVerify(docID, nil, createAttributes())

	// Alice stops.
	err = alice.Stop()
	assert.NoError(t, err)

	// Bob updates the document and tries to send it to Alice.
	// Bob will anchor the document without Alice's signature.
	payload = genericCoreAPIUpdate([]string{
		alice.GetMainAccount().GetAccountID().ToHexString(),
	})
	payload["document_id"] = docID

	docID, err = bobClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	bobClient.GetDocumentAndVerify(docID, nil, allAttributes())

	// Restart Alice.
	err = alice.Start()
	assert.NoError(t, err)

	// Alice should not have the latest doc.
	aliceClient.GetDocumentAndVerify(docID, nil, createAttributes())

	// Bob does another update.
	payload = genericCoreAPIUpdate([]string{
		alice.GetMainAccount().GetAccountID().ToHexString(),
	})

	payload["document_id"] = docID

	docID, err = bobClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	aliceClient.GetDocumentAndVerify(docID, nil, allAttributes())
	bobClient.GetDocumentAndVerify(docID, nil, allAttributes())
}

func TestDocumentsAPI_InvalidAttributes(t *testing.T) {
	t.Parallel()

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)

	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	payload := wrongGenericDocumentPayload(
		[]string{
			bob.GetMainAccount().GetAccountID().ToHexString(),
		},
	)
	response := aliceClient.CreateDocument("documents", http.StatusBadRequest, payload)
	errMsg := response.Raw()["message"].(string)
	assert.NotEmpty(t, errMsg)
}
