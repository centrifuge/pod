// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/notification"
	"github.com/stretchr/testify/assert"
)

func TestHost_BasicDocumentShare(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// alice shares a document with bob and charlie
	res := createDocument(alice.httpExpect, alice.id.String(), typeDocuments, http.StatusAccepted, genericCoreAPICreate([]string{bob.id.String(), charlie.id.String()}))
	txID := getJobID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), docIdentifier, nil, createAttributes())
	// alices job completes with a webhook
	msg, err := doctorFord.maeve.getReceivedMsg(alice.id.String(), int(notification.JobCompleted), txID)
	assert.NoError(t, err)
	assert.Equal(t, string(jobs.Success), msg.Status)

	// bobs node sends a webhook for received anchored doc
	msg, err = doctorFord.maeve.getReceivedMsg(bob.id.String(), int(notification.ReceivedPayload), docIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(alice.id.String()), strings.ToLower(msg.FromID))
	fmt.Println("Host test success")
}

func TestHost_RestartWithAccounts(t *testing.T) {
	// Name can be randomly generated
	tempHostName := "Sleepy"
	bootnode, err := doctorFord.bernard.p2pURL()
	assert.NoError(t, err)
	sleepyHost := doctorFord.createTempHost(tempHostName, defaultP2PTimeout, 8090, 38210, true, true, []string{bootnode})
	doctorFord.addNiceHost(tempHostName, sleepyHost)
	err = doctorFord.startTempHost(tempHostName)
	assert.NoError(t, err)
	up, err := sleepyHost.isLive(10 * time.Second)
	assert.NoError(t, err)
	assert.True(t, up)
	sleepyTS := doctorFord.getHostTestSuite(t, tempHostName)

	// Create accounts for new host
	err = sleepyHost.createAccounts(sleepyTS.httpExpect)
	assert.NoError(t, err)
	err = sleepyHost.loadAccounts(sleepyTS.httpExpect)
	assert.NoError(t, err)

	// Verify accounts are created
	acc1 := sleepyHost.accounts[0]
	res := getAccount(sleepyTS.httpExpect, sleepyTS.id.String(), http.StatusOK, acc1)
	acc1Res := res.Value("identity_id").String().NotEmpty()
	acc1Res.Equal(strings.ToLower(acc1))

	// Stop host
	sleepyHost.kill()

	// Start host
	doctorFord.reLive(t, tempHostName)
	up, err = sleepyHost.isLive(10 * time.Second)
	assert.NoError(t, err)
	assert.True(t, up)

	// Verify accounts are available after restart
	res = getAccount(sleepyTS.httpExpect, sleepyTS.id.String(), http.StatusOK, acc1)
	acc1Res = res.Value("identity_id").String().NotEmpty()
	acc1Res.Equal(strings.ToLower(acc1))
}
