// +build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"
	"time"

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
	docID := createAndCommitDocument(t, alice.httpExpect, alice.id.String(),
		genericCoreAPICreate([]string{bob.id.String(), charlie.id.String()}))

	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, createAttributes())
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, createAttributes())
	getDocumentAndVerify(t, charlie.httpExpect, charlie.id.String(), docID, nil, createAttributes())

	// bobs node sends a webhook for received anchored doc
	msg, err := doctorFord.maeve.getReceivedMsg(bob.id.String(), int(notification.ReceivedPayload), docID)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(alice.id.String()), strings.ToLower(msg.FromID))
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
