package ipfs_pinning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	logging "github.com/ipfs/go-log"
)

type PinataServiceClient interface {
	PinJSONToIPFS(ctx context.Context, data any, options *PinataOptions, metadata *PinataMetadata) (*PinJSONToIPFSResponse, error)
	Unpin(ctx context.Context, ipfsHash string) error
}

type client struct {
	log *logging.ZapEventLogger

	apiURL string

	JWTToken string

	c *http.Client
}

func NewPinataServiceClient(
	apiURL string,
	JWTToken string,
) PinataServiceClient {
	log := logging.Logger("pinata-service-client")

	return &client{
		log:      log,
		apiURL:   apiURL,
		JWTToken: JWTToken,
		c:        http.DefaultClient,
	}
}

const (
	pinJSONToIPFSPath = "/pinning/pinJSONToIPFS"
)

func (c *client) PinJSONToIPFS(ctx context.Context, data any, options *PinataOptions, metadata *PinataMetadata) (*PinJSONToIPFSResponse, error) {
	req := PinJSONToIPFSRequestBody{
		PinataOptions:  options,
		PinataMetadata: metadata,
		PinataContent:  data,
	}

	b, err := json.Marshal(req)

	if err != nil {
		c.log.Errorf("Couldn't marshal request to JSON: %s", err)

		return nil, ErrRequestJSONMarshal
	}

	res, err := c.sendRequest(ctx, http.MethodPost, pinJSONToIPFSPath, nil, bytes.NewReader(b))

	if err != nil {
		c.log.Errorf("Couldn't send PinJSONToIPFS request: %s", err)

		return nil, err
	}

	var r PinJSONToIPFSResponse

	if err := c.unmarshalResponse(res, &r); err != nil {
		c.log.Errorf("Response error: %s", err)

		return nil, err
	}

	return &r, nil
}

const (
	unpinPath = "/pinning/unpin"
)

func (c *client) Unpin(ctx context.Context, ipfsHash string) error {
	urlPath := fmt.Sprintf("%s/%s", unpinPath, ipfsHash)

	res, err := c.sendRequest(ctx, http.MethodDelete, urlPath, nil, nil)

	if err != nil {
		c.log.Errorf("Couldn't send Unpin request: %s", err)

		return err
	}

	if err := c.unmarshalResponse(res, nil); err != nil {
		c.log.Errorf("Response error: %s", err)

		return err
	}

	return nil
}

func (c *client) sendRequest(
	ctx context.Context,
	HTTPMethod string,
	URLPath string,
	queryParams url.Values,
	body io.Reader,
) (*http.Response, error) {
	u, err := url.Parse(c.apiURL + URLPath)

	if err != nil {
		c.log.Errorf("Couldn't parse URL: %s", err)

		return nil, ErrInvalidURL
	}

	if queryParams != nil {
		u.RawQuery = queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, HTTPMethod, u.String(), body)

	if err != nil {
		c.log.Errorf("Couldn't create HTTP request: %s", err)

		return nil, ErrHTTPRequestCreation
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.JWTToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.c.Do(req)

	if err != nil {
		c.log.Errorf("Couldn perform HTTP request: %s", err)

		return nil, ErrHTTPRequest
	}

	return res, nil
}

func (c *client) unmarshalResponse(res *http.Response, responseObj any) error {
	if res.Body == nil {
		return nil
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		c.log.Errorf("Couldn't read request body: %s", err)

		return ErrHTTPResponseBodyRead
	}

	if res.StatusCode >= http.StatusBadRequest {
		c.log.Errorf("Error response with status %d, response body:\n%s", res.StatusCode, string(b))

		return ErrHTTPResponse
	}

	if responseObj == nil {
		return nil
	}

	if err := json.Unmarshal(b, &responseObj); err != nil {
		c.log.Errorf("Couldn't unmarshal response: %s", err)

		return ErrResponseJSONUnmarshal
	}

	return nil
}
