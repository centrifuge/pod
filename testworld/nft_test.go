// +build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

func TestGenericMint_successful(t *testing.T) {
	defaultNFTMint(t, typeDocuments)
}

func defaultNFTMint(t *testing.T, documentType string) (string, nft.TokenID) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	registry := common.HexToAddress(alice.host.dappAddresses["genericNFT"])
	assetAddress := common.HexToAddress(alice.host.dappAddresses["assetManager"])

	// Alice shares document with Bob
	docPayload := genericCoreAPICreate([]string{bob.id.String()})
	attrs, pfs := getAttributeMapRequest(t, alice.id)
	docPayload["attributes"] = attrs
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusAccepted, docPayload)
	jobID := getJobID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), jobID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, attrs)
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, attrs)

	var response *httpexpect.Object
	var err error

	depositAddress := alice.id.String()

	// mint an NFT
	acr, err := alice.host.configService.GetAccount(alice.id.ToAddress().Bytes())
	assert.NoError(t, err)
	pfs = append(pfs, nft.GetSignatureProofField(t, acr))
	payload := map[string]interface{}{
		"document_id":           docIdentifier,
		"registry_address":      registry.String(),
		"deposit_address":       depositAddress, // Centrifuge address
		"proof_fields":          pfs,
		"asset_manager_address": assetAddress,
	}
	response, err = alice.host.mintNFT(alice.httpExpect, alice.id.String(), http.StatusAccepted, payload)
	assert.NoError(t, err, "mintNFT should be successful")
	jobID = getJobID(t, response)
	ok, err := waitForJobComplete(alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)
	assert.True(t, ok)

	docVal := getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, attrs)
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
	return docIdentifier, tokenID
}

func TestTransferNFT_successful(t *testing.T) {
	_, tokenID := defaultNFTMint(t, typeDocuments)
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	registry := alice.host.dappAddresses["genericNFT"]

	ownerOfPayload := map[string]interface{}{
		"token_id":         tokenID.String(),
		"registry_address": registry,
	}

	transferPayload := map[string]interface{}{
		"token_id":         tokenID.String(),
		"registry_address": registry,
		"to":               bob.id.String(),
	}

	// nft owner should be alice
	resp, err := alice.host.ownerOfNFT(alice.httpExpect, alice.id.String(), http.StatusOK, ownerOfPayload)
	assert.NoError(t, err)
	resp.Path("$.owner").String().Equal(strings.ToLower(alice.id.String()))

	// transfer nft from alice to bob
	response, err := alice.host.transferNFT(alice.httpExpect, alice.id.String(), http.StatusOK, transferPayload)
	assert.NoError(t, err)
	jobID := getJobID(t, response)
	ok, err := waitForJobComplete(alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)
	assert.True(t, ok)

	// nft owner should be bob
	resp, err = alice.host.ownerOfNFT(alice.httpExpect, alice.id.String(), http.StatusOK, ownerOfPayload)
	assert.NoError(t, err)
	resp.Path("$.owner").String().Equal(strings.ToLower(bob.id.String()))
}

func getAttributeMapRequest(t *testing.T, did identity.DID) (coreapi.AttributeMapRequest, []string) {
	attrs, pfs := nft.GetAttributes(t, did)
	amr := make(coreapi.AttributeMapRequest)
	for _, attr := range attrs {
		v, err := attr.Value.String()
		assert.NoError(t, err)
		amr[attr.KeyLabel] = coreapi.AttributeRequest{
			Type:  attr.Value.Type.String(),
			Value: v,
		}
	}
	return amr, pfs
}
