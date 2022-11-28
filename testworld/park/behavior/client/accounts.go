//go:build testworld

package client

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/gavv/httpexpect"
)

func (c *Client) GetAccount(httpStatus int, identifier string) *httpexpect.Object {
	resp := addCommonHeaders(c.expect.GET("/v2/accounts/"+identifier), c.jwtToken).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func (c *Client) GetAllAccounts(httpStatus int) *httpexpect.Object {
	resp := addCommonHeaders(c.expect.GET("/v2/accounts"), c.jwtToken).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func (c *Client) GenerateAccount(
	payload map[string]any,
	statusCode int,
) *httpexpect.Response {
	req := addCommonHeaders(c.expect.POST("/v2/accounts/generate"), c.jwtToken).WithJSON(payload)

	return req.Expect().Status(statusCode)
}

func (c *Client) GetSelf(statusCode int) *httpexpect.Response {
	return addCommonHeaders(c.expect.GET("/v2/accounts/self"), c.jwtToken).
		Expect().
		Status(statusCode)
}

func (c *Client) SignPayload(
	accountID *types.AccountID,
	payload map[string]any,
	httpStatus int,
) *httpexpect.Response {
	path := fmt.Sprintf("/v2/accounts/%s/sign", accountID.ToHexString())

	req := addCommonHeaders(c.expect.POST(path), c.jwtToken).WithJSON(payload)

	return req.Expect().Status(httpStatus)
}

func (c *Client) getAccounts(accounts *httpexpect.Array) map[string]string {
	accIDs := make(map[string]string)

	for i := 0; i < int(accounts.Length().Raw()); i++ {
		val := accounts.Element(i).Path("$.identity_id").String().NotEmpty().Raw()
		accIDs[val] = val
	}

	return accIDs
}
