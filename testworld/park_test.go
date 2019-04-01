// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHost_BasicDocumentShare(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// alice shares a document with bob and charlie
	res := createDocument(alice.httpExpect, alice.id.String(), typeInvoice, http.StatusOK, defaultInvoicePayload([]string{bob.id.String(), charlie.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(alice.httpExpect, alice.id.String(), typeInvoice, params)
	getDocumentAndCheck(bob.httpExpect, bob.id.String(), typeInvoice, params)
	getDocumentAndCheck(charlie.httpExpect, charlie.id.String(), typeInvoice, params)
	fmt.Println("Host test success")
}

func TestHost_RestartWithAccounts(t *testing.T) {
	t.Parallel()

	// Name can be randomly generated
	tempHostName := "Sleepy"
	bootnode, err := doctorFord.bernard.p2pURL()
	assert.NoError(t, err)
	sleepyHost := doctorFord.createTempHost(tempHostName, doctorFord.twConfigName, defaultP2PTimeout, 8088, 38208, true, true, []string{bootnode})
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
	acc1Res.Equal(acc1)

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
	acc1Res.Equal(acc1)
}
