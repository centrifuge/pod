//go:build testworld

package client

import (
	"github.com/centrifuge/pod/http/coreapi"
	"github.com/gavv/httpexpect"
)

type AssetRequest struct {
	PoolID  string
	LoanID  string
	AssetID string
}

func (c *Client) GetAsset(assetRequest *AssetRequest, expectedStatus int) *httpexpect.Object {
	req := c.expect.GET("/v3/investor/assets").
		WithQuery(coreapi.PoolIDQueryParam, assetRequest.PoolID).
		WithQuery(coreapi.LoanIDQueryParam, assetRequest.LoanID).
		WithQuery(coreapi.AssetIDQueryParam, assetRequest.AssetID)

	objGet := addCommonHeaders(req, c.authToken).
		Expect().Status(expectedStatus).JSON().NotNull().Object()

	return objGet
}
