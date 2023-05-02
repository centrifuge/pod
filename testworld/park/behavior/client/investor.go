//go:build testworld

package client

import (
	"net/http"

	"github.com/centrifuge/pod/http/coreapi"
	"github.com/gavv/httpexpect"
)

func (c *Client) GetAsset(poolID, loanID, assetID string) *httpexpect.Value {
	req := c.expect.GET("/v3/investor/assets").
		WithQuery(coreapi.PoolIDNameParam, poolID).
		WithQuery(coreapi.LoanIDNameParam, loanID).
		WithQuery(coreapi.AssetIDNameParam, assetID)

	objGet := addCommonHeaders(req, c.jwtToken).
		Expect().Status(http.StatusOK).JSON().NotNull()

	return objGet
}
