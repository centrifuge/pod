// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const tokenIdLength = 77

func TestPaymentObligationMint_successful(t *testing.T) {
	t.Parallel()

	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice shares invoice document with Bob
	res, err := alice.host.createInvoice(alice.httpExpect, http.StatusOK, defaultNFTPayload([]string{bob.id.String()}))
	if err != nil {
		t.Error(err)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getInvoiceAndCheck(alice.httpExpect, params)
	getInvoiceAndCheck(bob.httpExpect, params)

	// mint an NFT
	test := struct {
		httpStatus int
		payload    map[string]interface{}
	}{
		http.StatusOK,
		map[string]interface{}{

			"identifier":      docIdentifier,
			"registryAddress": doctorFord.contractAddresses.PaymentObligationAddr,
			"depositAddress":  "0xf72855759a39fb75fc7341139f5d7a3974d4da08", // dummy address
			"proofFields":     []string{"invoice.gross_amount", "invoice.currency", "invoice.due_date", "collaborators[0]"},
		},
	}

	response, err := alice.host.mintNFT(alice.httpExpect, test.httpStatus, test.payload)
	assert.Nil(t, err, "mintNFT should be successful")
	assert.True(t, len(response.Value("token_id").String().Raw()) >= tokenIdLength, "successful tokenId should have length 77")

}

func TestPaymentObligationMint_errors(t *testing.T) {
	t.Parallel()

	alice := doctorFord.getHostTestSuite(t, "Alice")

	tests := []struct {
		errorMsg   string
		httpStatus int
		payload    map[string]interface{}
	}{
		{
			"RegistryAddress is not a valid Ethereum address",
			http.StatusInternalServerError,
			map[string]interface{}{

				"registryAddress": "0x123",
			},
		},
		{
			"DepositAddress is not a valid Ethereum address",
			http.StatusInternalServerError,
			map[string]interface{}{

				"registryAddress": "0xf72855759a39fb75fc7341139f5d7a3974d4da08", //dummy address
				"depositAddress":  "abc",
			},
		},
		{
			"no service exists for provided documentID",
			http.StatusInternalServerError,
			map[string]interface{}{

				"identifier":      "0x12121212",
				"registryAddress": "0xf72855759a39fb75fc7341139f5d7a3974d4da08", //dummy address
				"depositAddress":  "0xf72855759a39fb75fc7341139f5d7a3974d4da08", //dummy address
			},
		},
	}

	for _, test := range tests {
		response, err := alice.host.mintNFT(alice.httpExpect, test.httpStatus, test.payload)
		assert.Nil(t, err, "it should be possible to call the API endpoint")
		response.Value("message").String().Contains(test.errorMsg)

	}

}
