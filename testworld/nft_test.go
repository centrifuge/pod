// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

func TestInvoiceUnpaidMint_invoice_successful(t *testing.T) {
	t.Parallel()
	invoiceUnpaidMint(t, typeInvoice, true, true, true, false)
}

func TestPaymentObligationWrapperMint_invoice_successful(t *testing.T) {
	t.Parallel()
	invoiceUnpaidMint(t, typeInvoice, false, false, false, true)
}

/* TODO: testcase not stable
func TestInvoiceUnpaidMint_po_successful(t *testing.T) {
	t.Parallel()
	invoiceUnpaidMint(t, typePO)
}
*/

func invoiceUnpaidMint(t *testing.T, documentType string, grantNFTAccess, tokenProof, nftReadAccessProof bool, poWrapper bool) nft.TokenID {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	registry := alice.host.config.GetContractAddress(config.InvoiceUnpaidNFT)

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
	getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), documentType, params, false)
	getDocumentAndCheck(t, bob.httpExpect, bob.id.String(), documentType, params, false)

	var response *httpexpect.Object
	var err error
	
	depositAddress := alice.id.String()

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
		payload := map[string]interface{}{
			"identifier":                docIdentifier,
			"registryAddress":           registry.String(),
			"depositAddress":            depositAddress, // Centrifuge address
			"proofFields":               []string{proofPrefix + ".gross_amount", proofPrefix + ".currency", proofPrefix + ".date_due", proofPrefix + ".sender", proofPrefix + ".status", signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"},
			"submitTokenProof":          tokenProof,
			"submitNftOwnerAccessProof": nftReadAccessProof,
			"grantNftAccess":            grantNFTAccess,
		}
		response, err = alice.host.mintNFT(alice.httpExpect, alice.id.String(), http.StatusOK, payload)

	} else {
		// mint a PO NFT
		payload := map[string]interface{}{
			"identifier":     docIdentifier,
			"depositAddress": depositAddress, // Centrifuge address
		}
		response, err = alice.host.mintUnpaidInvoiceNFT(alice.httpExpect, alice.id.String(), http.StatusOK, docIdentifier, payload)
	}

	assert.NoError(t, err, "mintNFT should be successful")
	txID = getTransactionID(t, response)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docVal := getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), documentType, params, false)
	assert.True(t, len(docVal.Path("$.header.nfts[0].token_id").String().Raw()) > 0, "successful tokenId should have length 77")
	assert.True(t, len(docVal.Path("$.header.nfts[0].token_index").String().Raw()) > 0, "successful tokenIndex should have a value")

	tokenID, err := nft.TokenIDFromString(docVal.Path("$.header.nfts[0].token_id").String().Raw())
	assert.NoError(t, err, "token ID should be correct")
	respOwner := docVal.Path("$.header.nfts[0].owner").String().Raw()
	assert.NoError(t, err, "token ID should be correct")
	owner, err := alice.host.tokenRegistry.OwnerOf(registry, tokenID.BigInt().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(depositAddress), strings.ToLower(owner.Hex()))
	assert.Equal(t, strings.ToLower(respOwner), strings.ToLower(owner.Hex()))
	return tokenID
}

func TestInvoiceUnpaidMint_errors(t *testing.T) {
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

func TestTransferNFT_successful(t *testing.T) {
	t.Parallel()
	tokenID := invoiceUnpaidMint(t, typeInvoice, false, false, false, true)
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	registry := alice.host.config.GetContractAddress(config.InvoiceUnpaidNFT)

	transferPayload := map[string]interface{}{
		"tokenId": tokenID.String(),
		"registryAddress":  registry.String(),
		"to": bob.id.String(),
	}

	response, err := alice.host.transferNFT(alice.httpExpect, alice.id.String(), http.StatusOK, transferPayload)
	assert.NoError(t, err)
	txID := getTransactionID(t, response)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}
	
}
