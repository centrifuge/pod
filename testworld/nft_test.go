// +build testworld

package testworld

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

func TestGenericMint_successful(t *testing.T) {
	defaultNFTMint(t)
}

func defaultNFTMint(t *testing.T) (string, nft.TokenID) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	registry := common.HexToAddress(alice.host.dappAddresses["genericNFT"])
	assetAddress := common.HexToAddress(alice.host.dappAddresses["assetManager"])

	// Alice shares document with Bob
	docPayload := genericCoreAPICreate([]string{alice.id.String(), bob.id.String()})
	attrs, pfs := getAttributeMapRequest(t, alice.id)
	docPayload["attributes"] = attrs
	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.String(), docPayload)
	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, attrs)
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, attrs)

	var response *httpexpect.Object
	var err error

	depositAddress := alice.id.String()

	// mint an NFT
	acr, err := alice.host.configService.GetAccount(alice.id.ToAddress().Bytes())
	assert.NoError(t, err)
	pfs = append(pfs, nft.GetSignatureProofField(t, acr))
	payload := map[string]interface{}{
		"document_id":           docID,
		"registry_address":      registry.String(),
		"deposit_address":       depositAddress, // Centrifuge address
		"proof_fields":          pfs,
		"asset_manager_address": assetAddress,
	}
	response, err = alice.host.mintNFT(alice.httpExpect, alice.id.String(), http.StatusAccepted, payload)
	assert.NoError(t, err, "mintNFT should be successful")
	jobID := getJobID(t, response)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)

	docVal := getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, attrs)
	assert.True(t, len(docVal.Path("$.header.nfts[0].token_id").String().Raw()) > 0, "successful tokenId should have length 77")

	tokenID, err := nft.TokenIDFromString(docVal.Path("$.header.nfts[0].token_id").String().Raw())
	assert.NoError(t, err, "token ID should be correct")
	respOwner := docVal.Path("$.header.nfts[0].owner").String().Raw()
	assert.NoError(t, err, "token ID should be correct")
	owner, err := alice.host.tokenRegistry.OwnerOf(registry, tokenID.BigInt().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(depositAddress), strings.ToLower(owner.Hex()))
	assert.Equal(t, strings.ToLower(respOwner), strings.ToLower(owner.Hex()))
	return docID, tokenID
}

func TestTransferNFT_successful(t *testing.T) {
	_, tokenID := defaultNFTMint(t)
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
	response, err := alice.host.transferNFT(alice.httpExpect, alice.id.String(), http.StatusAccepted, transferPayload)
	assert.NoError(t, err)
	jobID := getJobID(t, response)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)

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

func TestNFTOnCC(t *testing.T) {
	t.Parallel()

	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	fmt.Println("Creating a new registry with alice as owner...")
	acc, err := alice.host.configService.GetAccount(alice.id[:])
	assert.NoError(t, err)
	ctx := contextutil.WithAccount(context.Background(), acc)
	proofFields := [][]byte{
		// originator
		hexutil.MustDecode("0x010000000000001ce24e7917d4fcaf79095539ac23af9f6d5c80ea8b0d95c9cd860152bff8fdab1700000005"),
		// asset value
		hexutil.MustDecode("0x010000000000001ccd35852d8705a28d4f83ba46f02ebdf46daf03638b40da74b9371d715976e6dd00000005"),
		// asset identifier
		hexutil.MustDecode("0x010000000000001cbbaa573c53fa357a3b53624eb6deab5f4c758f299cffc2b0b6162400e3ec13ee00000005"),
		// MaturityDate
		hexutil.MustDecode("0x010000000000001ce5588a8a267ed4c32962568afe216d4ba70ae60576a611e3ca557b84f1724e2900000005"),
	}
	info := nft.RegistryInfo{
		OwnerCanBurn: true,
		Fields:       proofFields,
	}
	registry, err := alice.host.nftAPI.CreateRegistry(ctx, info)
	assert.NoError(t, err)
	fmt.Println("Registry:", registry.Hex())

	// Alice shares document with Bob
	docPayload := genericCoreAPICreate([]string{alice.id.String(), bob.id.String()})
	attrs, pfs := getAttributeMapRequest(t, alice.id)
	docPayload["attributes"] = attrs
	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.String(), docPayload)
	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, attrs)
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, attrs)

	payload := map[string]interface{}{
		"document_id":      docID,
		"registry_address": registry.Hex(),
		"deposit_address":  alice.host.centChainID,
		"proof_fields":     pfs,
	}

	response := mintNFTOnCC(alice.httpExpect, alice.id.String(), http.StatusAccepted, payload)
	jobID := getJobID(t, response)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)

	docVal := getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, attrs)
	assert.True(t, len(docVal.Path("$.header.nfts[0].token_id").String().Raw()) > 0, "successful tokenId should have length 77")
	tokenID, err := nft.TokenIDFromString(docVal.Path("$.header.nfts[0].token_id").String().Raw())
	assert.NoError(t, err, "token ID should be correct")
	owner := docVal.Path("$.header.nfts[0].owner").String().Raw()
	assert.NoError(t, err, "token ID should be correct")
	assert.Equal(t, strings.ToLower(alice.host.centChainID), strings.ToLower(owner))

	// verify owner
	ownerResp := ownerOfNFTOnCC(alice.httpExpect, alice.id.String(), http.StatusOK, map[string]interface{}{
		"registry_address": registry.Hex(),
		"token_id":         tokenID.String(),
	})

	owner = ownerResp.Path("$.owner").String().Raw()
	assert.Equal(t, strings.ToLower(alice.host.centChainID), strings.ToLower(owner))
	fmt.Println("Token minted and owner verified")

	// transfer nft
	response = transferNFTOnCC(alice.httpExpect, alice.id.String(), http.StatusAccepted, map[string]interface{}{
		"registry_address": registry.Hex(),
		"token_id":         tokenID.String(),
		"to":               bob.host.centChainID,
	})
	jobID = getJobID(t, response)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)

	ownerResp = ownerOfNFTOnCC(alice.httpExpect, alice.id.String(), http.StatusOK, map[string]interface{}{
		"registry_address": registry.Hex(),
		"token_id":         tokenID.String(),
	})

	owner = ownerResp.Path("$.owner").String().Raw()
	assert.Equal(t, strings.ToLower(bob.host.centChainID), strings.ToLower(owner))
	fmt.Println("Token transferred successfully")
}
