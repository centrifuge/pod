// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/stretchr/testify/assert"
)

func TestPaymentObligationMint_invoice_successful(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                                     string
		grantAccess, tokenProof, readAccessProof bool
	}{
		//{
		//	name:        "grant access",
		//	grantAccess: true,
		//},
		//
		//{
		//	name:       "token proof",
		//	tokenProof: true,
		//},
		//
		//{
		//	name:        "grant access and token proof",
		//	grantAccess: true,
		//	tokenProof:  true,
		//},
		//
		//{
		//	name:            "grant access and read access proof",
		//	grantAccess:     true,
		//	readAccessProof: true,
		//},

		{
			name:            "grant access, token proof and read access proof",
			grantAccess:     true,
			tokenProof:      true,
			readAccessProof: true,
		},
	}

	for _, c := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			paymentObligationMint(t, typeInvoice, c.grantAccess, c.tokenProof, c.readAccessProof)
		})
	}
}

/* TODO: testcase not stable
func TestPaymentObligationMint_po_successful(t *testing.T) {
	t.Parallel()
	paymentObligationMint(t, typePO)
}
*/

func paymentObligationMint(t *testing.T, documentType string, grantNFTAccess, tokenProof, nftReadAccessProof bool) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice shares document with Bob
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusOK, defaultNFTPayload(documentType, []string{bob.id.String()}))
	txID := getTransactionID(t, res)

	waitTillStatus(t, alice.httpExpect, alice.id.String(), txID, "success")

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(alice.httpExpect, alice.id.String(), documentType, params)
	getDocumentAndCheck(bob.httpExpect, bob.id.String(), documentType, params)

	proofPrefix := documentType
	if proofPrefix == typePO {
		proofPrefix = poPrefix
	}
	acc, err := alice.host.configService.GetAccount(alice.id[:])
	if err != nil {
		t.Error(err)
	}
	keys, err := acc.GetKeys()
	if err != nil {
		t.Error(err)
	}
	signerId := hexutil.Encode(append(alice.id[:], keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signingRoot := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signatureSender := fmt.Sprintf("%s.signatures[%s].signature", documents.SignaturesTreePrefix, signerId)

	// mint an NFT
	test := struct {
		httpStatus int
		payload    map[string]interface{}
	}{
		http.StatusOK,
		map[string]interface{}{

			"identifier":                docIdentifier,
			"registryAddress":           doctorFord.getHost("Alice").config.GetContractAddress(config.PaymentObligation).String(),
			"depositAddress":            "0x44a0579754d6c94e7bb2c26bfa7394311cc50ccb", // Centrifuge address
			"proofFields":               []string{proofPrefix + ".gross_amount", proofPrefix + ".currency", proofPrefix + ".due_date", proofPrefix + ".sender", proofPrefix + ".invoice_status", signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"},
			"submitTokenProof":          tokenProof,
			"submitNftOwnerAccessProof": nftReadAccessProof,
			"grantNftAccess":            grantNFTAccess,
		},
	}

	response, err := alice.host.mintNFT(alice.httpExpect, alice.id.String(), test.httpStatus, test.payload)
	txID = getTransactionID(t, response)
	waitTillStatus(t, alice.httpExpect, alice.id.String(), txID, "success")

	assert.Nil(t, err, "mintNFT should be successful")
	assert.True(t, len(response.Value("token_id").String().Raw()) > 0, "successful tokenId should have length 77")

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
	}
	for _, test := range tests {
		t.Run(test.errorMsg, func(t *testing.T) {
			t.Parallel()
			response, err := alice.host.mintNFT(alice.httpExpect, alice.id.String(), test.httpStatus, test.payload)
			assert.Nil(t, err, "it should be possible to call the API endpoint")
			response.Value("error").String().Contains(test.errorMsg)
		})
	}
}
