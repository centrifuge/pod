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
	PinJSONToIPFS(ctx context.Context, req *PinJSONToIPFSRequest) (*PinJSONToIPFSResponse, error)
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
) (PinataServiceClient, error) {
	if err := validateAPIURL(apiURL); err != nil {
		return nil, err
	}

	if JWTToken == "" {
		return nil, ErrMissingAPIJWT
	}

	log := logging.Logger("pinata-service-client")

	return &client{
		log:      log,
		apiURL:   apiURL,
		JWTToken: JWTToken,
		c:        http.DefaultClient,
	}, nil
}

func validateAPIURL(u string) error {
	if u == "" {
		return ErrMissingAPIURL
	}

	if _, err := url.ParseRequestURI(u); err != nil {
		return ErrInvalidURL
	}

	return nil
}

const (
	pinJSONToIPFSPath = "/pinning/pinJSONToIPFS"
)

func (c *client) PinJSONToIPFS(ctx context.Context, req *PinJSONToIPFSRequest) (*PinJSONToIPFSResponse, error) {
	if req == nil {
		c.log.Error("Request is missing")

		return nil, ErrMissingRequest
	}

	b, err := json.Marshal(req)

	if err != nil {
		c.log.Errorf("Couldn't marshal request to JSON: %s", err)

		return nil, ErrRequestJSONMarshal
	}

	res, err := c.sendRequest(ctx, http.MethodPost, pinJSONToIPFSPath, bytes.NewReader(b))

	if err != nil {
		c.log.Errorf("Couldn't send PinJSONToIPFS request: %s", err)

		return nil, err
	}

	var r PinJSONToIPFSResponse

	if err := c.handleResponse(res, &r); err != nil {
		c.log.Errorf("Response error: %s", err)

		return nil, err
	}

	return &r, nil
}

const (
	unpinPath = "/pinning/unpin"
)

func (c *client) Unpin(ctx context.Context, ipfsHash string) error {
	if ipfsHash == "" {
		c.log.Error("IPFS hash is missing")

		return ErrMissingIPFSHash
	}

	urlPath := fmt.Sprintf("%s/%s", unpinPath, ipfsHash)

	res, err := c.sendRequest(ctx, http.MethodDelete, urlPath, nil)

	if err != nil {
		c.log.Errorf("Couldn't send Unpin request: %s", err)

		return err
	}

	if err := c.handleResponse(res, nil); err != nil {
		c.log.Errorf("Response error: %s", err)

		return err
	}

	return nil
}

func (c *client) sendRequest(
	ctx context.Context,
	HTTPMethod string,
	URLPath string,
	body io.Reader,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, HTTPMethod, c.apiURL+URLPath, body)

	if err != nil {
		c.log.Errorf("Couldn't create HTTP request: %s", err)

		return nil, ErrHTTPRequestCreation
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.JWTToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.c.Do(req)

	if err != nil {
		c.log.Errorf("Couldn perform HTTP request: %s", err)

		return nil, ErrHTTPRequest
	}

	return res, nil
}

func (c *client) handleResponse(res *http.Response, responseObj any) error {
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
