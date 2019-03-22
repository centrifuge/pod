// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestPaymentObligationMint_invoice_successful(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                                     string
		grantAccess, tokenProof, readAccessProof bool
	}{
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
			paymentObligationMint(t, typeInvoice, c.grantAccess, c.tokenProof, c.readAccessProof, false)
		})
	}
}

func TestPaymentObligationWrapperMint_invoice_successful(t *testing.T) {
	t.Parallel()
	paymentObligationMint(t, typeInvoice, false, false, false, true)
}

/* TODO: testcase not stable
func TestPaymentObligationMint_po_successful(t *testing.T) {
	t.Parallel()
	paymentObligationMint(t, typePO)
}
*/

func paymentObligationMint(t *testing.T, documentType string, grantNFTAccess, tokenProof, nftReadAccessProof bool, poWrapper bool) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	registry := alice.host.config.GetContractAddress(config.PaymentObligation)

	// Alice shares document with Bob
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusOK, defaultNFTPayload(documentType, []string{bob.id.String()}, alice.id.String()))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

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

	var response *httpexpect.Object
	var err error

	if !poWrapper {
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
				"registryAddress":           registry.String(),
				"depositAddress":            "0x44a0579754d6c94e7bb2c26bfa7394311cc50ccb", // Centrifuge address
				"proofFields":               []string{proofPrefix + ".gross_amount", proofPrefix + ".currency", proofPrefix + ".date_due", proofPrefix + ".sender", proofPrefix + ".status", signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"},
				"submitTokenProof":          tokenProof,
				"submitNftOwnerAccessProof": nftReadAccessProof,
				"grantNftAccess":            grantNFTAccess,
			},
		}
		response, err = alice.host.mintNFT(alice.httpExpect, alice.id.String(), test.httpStatus, test.payload)
	} else {
		// mint a PO NFT
		test := struct {
			httpStatus int
			documentID string
			payload    map[string]interface{}
		}{
			http.StatusOK,
			docIdentifier,
			map[string]interface{}{

				"identifier":     docIdentifier,
				"depositAddress": "0x44a0579754d6c94e7bb2c26bfa7394311cc50ccb", // Centrifuge address
			},
		}
		response, err = alice.host.mintPONFT(alice.httpExpect, alice.id.String(), test.httpStatus, test.documentID, test.payload)
	}

	assert.NoError(t, err, "mintNFT should be successful")
	txID = getTransactionID(t, response)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}
	assert.True(t, len(response.Value("token_id").String().Raw()) > 0, "successful tokenId should have length 77")

	tokenID, err := nft.TokenIDFromString(response.Value("token_id").String().Raw())
	assert.NoError(t, err, "token ID should be correct")
	owner, err := alice.host.tokenRegistry.OwnerOf(registry, tokenID.BigInt().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower("0x44a0579754d6c94e7bb2c26bfa7394311cc50ccb"), strings.ToLower(owner.Hex()))
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
