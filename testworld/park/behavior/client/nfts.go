//go:build testworld

package client

import (
	"fmt"

	"github.com/gavv/httpexpect"
)

func (c *Client) CommitAndMintNFT(httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	path := fmt.Sprintf("/v3/nfts/collections/%d/commit_and_mint", payload["collection_id"])
	resp := addCommonHeaders(c.expect.POST(path), c.jwtToken).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func (c *Client) MintNFT(httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	path := fmt.Sprintf("/v3/nfts/collections/%d/mint", payload["collection_id"])
	resp := addCommonHeaders(c.expect.POST(path), c.jwtToken).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func (c *Client) GetOwnerOfNFT(httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	path := fmt.Sprintf(
		"/v3/nfts/collections/%d/items/%s/owner",
		payload["collection_id"],
		payload["item_id"],
	)

	resp := addCommonHeaders(c.expect.GET(path), c.jwtToken).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func (c *Client) GetMetadataOfNFT(httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	path := fmt.Sprintf(
		"/v3/nfts/collections/%d/items/%s/metadata",
		payload["collection_id"],
		payload["item_id"],
	)

	resp := addCommonHeaders(c.expect.GET(path), c.jwtToken).
		Expect().Status(httpStatus)

	return resp.JSON().Object()
}

func (c *Client) GetAttributeOfNFT(httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	path := fmt.Sprintf(
		"/v3/nfts/collections/%d/items/%s/attribute/%s",
		payload["collection_id"],
		payload["item_id"],
		payload["attribute_name"],
	)

	resp := addCommonHeaders(c.expect.GET(path), c.jwtToken).
		Expect().Status(httpStatus)

	return resp.JSON().Object()
}

func (c *Client) CreateNFTCollection(httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(c.expect.POST("/v3/nfts/collections"), c.jwtToken).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}
