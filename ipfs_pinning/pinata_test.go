package ipfs_pinning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/validation"

	"github.com/centrifuge/go-centrifuge/errors"

	logging "github.com/ipfs/go-log"

	"github.com/stretchr/testify/assert"
)

type erroneousContent struct{}

func (e erroneousContent) MarshalJSON() ([]byte, error) {
	return nil, errors.New("error")
}

type errReader struct{}

func (e *errReader) Read([]byte) (int, error) {
	return 0, errors.New("error")
}

func TestNewPinataServiceClient(t *testing.T) {
	_, err := NewPinataServiceClient("", "")
	assert.ErrorIs(t, err, validation.ErrMissingURL)

	_, err = NewPinataServiceClient("https://centrifuge.io", "")
	assert.ErrorIs(t, err, ErrMissingAPIJWT)

	_, err = NewPinataServiceClient("https://centrifuge.io", "some_jwt_token")
}

func TestClient_PinData(t *testing.T) {
	ctx := context.Background()

	type testContent struct {
		Field1 string            `json:"field_1"`
		Field2 int               `json:"field_2"`
		Field3 []string          `json:"field_3"`
		Field4 map[string]string `json:"field_4"`
	}

	reqData := &testContent{
		Field1: "some_string",
		Field2: 124,
		Field3: []string{"first", "second"},
		Field4: map[string]string{
			"key1": "value1",
		},
	}

	req := &PinRequest{
		CIDVersion: 1,
		Data:       reqData,
		Metadata: map[string]string{
			"meta_key": "meta_value",
		},
	}

	pinataReq := &PinJSONToIPFSRequest{
		PinataOptions: &PinataOptions{
			CIDVersion: 1,
		},
		PinataMetadata: &PinataMetadata{
			KeyValues: req.Metadata,
		},
		PinataContent: req.Data,
	}

	pinataReqJSONBytes, err := json.Marshal(pinataReq)
	assert.NoError(t, err)

	testRes := &PinJSONToIPFSResponse{
		IpfsHash:  "test-hash",
		PinSize:   634,
		Timestamp: time.Now(),
	}

	testResJSONBytes, err := json.Marshal(testRes)
	assert.NoError(t, err)

	jwtToken := "some_jwt_token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, pinJSONToIPFSPath, r.URL.Path)
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+jwtToken, authHeader)

		assert.NotNil(t, r.Body)

		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		assert.Equal(t, pinataReqJSONBytes, b)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(testResJSONBytes)
		assert.NoError(t, err)
	}))

	defer srv.Close()

	c, err := NewPinataServiceClient(srv.URL, jwtToken)
	assert.NoError(t, err)

	res, err := c.PinData(ctx, req)
	assert.NoError(t, err)
	assert.IsType(t, &PinResponse{}, res)
	assert.Equal(t, testRes.IpfsHash, res.CID)
}

func TestClient_PinJSONToIPFS_RequestMissingError(t *testing.T) {
	ctx := context.Background()

	c, err := NewPinataServiceClient("https://centrifuge.io", "some_jwt_token")
	assert.NoError(t, err)

	res, err := c.PinData(ctx, nil)
	assert.ErrorIs(t, err, ErrInvalidPinningRequest)
	assert.Nil(t, res)
}

func TestClient_PinJSONToIPFS_RequestMarshalError(t *testing.T) {
	ctx := context.Background()

	req := &PinRequest{
		CIDVersion: 1,
		Data:       erroneousContent{},
	}

	c, err := NewPinataServiceClient("https://centrifuge.io", "some_jwt_token")
	assert.NoError(t, err)

	res, err := c.PinData(ctx, req)
	assert.ErrorIs(t, err, ErrRequestJSONMarshal)
	assert.Nil(t, res)
}

func TestClient_PinJSONToIPFS_RequestSendError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	type testContent struct {
		Field1 string            `json:"field_1"`
		Field2 int               `json:"field_2"`
		Field3 []string          `json:"field_3"`
		Field4 map[string]string `json:"field_4"`
	}

	reqContent := &testContent{
		Field1: "some_string",
		Field2: 124,
		Field3: []string{"first", "second"},
		Field4: map[string]string{
			"key1": "value1",
		},
	}

	req := &PinRequest{
		CIDVersion: 1,
		Data:       reqContent,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	defer srv.Close()

	c, err := NewPinataServiceClient(srv.URL, "some_jwt_token")
	assert.NoError(t, err)

	cancel()

	res, err := c.PinData(ctx, req)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestClient_PinJSONToIPFS_UnmarshalResponseError(t *testing.T) {
	ctx := context.Background()

	type testContent struct {
		Field1 string            `json:"field_1"`
		Field2 int               `json:"field_2"`
		Field3 []string          `json:"field_3"`
		Field4 map[string]string `json:"field_4"`
	}

	reqData := &testContent{
		Field1: "some_string",
		Field2: 124,
		Field3: []string{"first", "second"},
		Field4: map[string]string{
			"key1": "value1",
		},
	}

	req := &PinRequest{
		CIDVersion: 1,
		Data:       reqData,
		Metadata: map[string]string{
			"meta_key": "meta_value",
		},
	}

	pinataReq := &PinJSONToIPFSRequest{
		PinataOptions: &PinataOptions{
			CIDVersion: 1,
		},
		PinataMetadata: &PinataMetadata{
			KeyValues: req.Metadata,
		},
		PinataContent: req.Data,
	}

	pinataReqJSONBytes, err := json.Marshal(pinataReq)
	assert.NoError(t, err)

	jwtToken := "some_jwt_token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, pinJSONToIPFSPath, r.URL.Path)
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+jwtToken, authHeader)

		assert.NotNil(t, r.Body)

		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		assert.Equal(t, pinataReqJSONBytes, b)

		w.WriteHeader(http.StatusBadRequest)
	}))

	defer srv.Close()

	c, err := NewPinataServiceClient(srv.URL, jwtToken)
	assert.NoError(t, err)

	res, err := c.PinData(ctx, req)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestClient_Unpin(t *testing.T) {
	ctx := context.Background()
	ipfsHash := "some_ipfs_hash"
	jwtToken := "some_jwt_token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, fmt.Sprintf("%s/%s", unpinPath, ipfsHash), r.URL.Path)
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+jwtToken, authHeader)

		w.WriteHeader(http.StatusOK)
	}))

	defer srv.Close()

	c, err := NewPinataServiceClient(srv.URL, "some_jwt_token")
	assert.NoError(t, err)

	err = c.UnpinData(ctx, ipfsHash)
	assert.Nil(t, err)
}

func TestClient_Unpin_RequestSendError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ipfsHash := "some_ipfs_hash"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	defer srv.Close()

	c, err := NewPinataServiceClient(srv.URL, "some_jwt_token")
	assert.NoError(t, err)

	cancel()

	err = c.UnpinData(ctx, ipfsHash)
	assert.NotNil(t, err)
}

func TestClient_Unpin_ResponseError(t *testing.T) {
	ctx := context.Background()
	ipfsHash := "some_ipfs_hash"
	jwtToken := "some_jwt_token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, fmt.Sprintf("%s/%s", unpinPath, ipfsHash), r.URL.Path)
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+jwtToken, authHeader)

		w.WriteHeader(http.StatusBadRequest)
	}))

	defer srv.Close()

	c, err := NewPinataServiceClient(srv.URL, "some_jwt_token")
	assert.NoError(t, err)

	err = c.UnpinData(ctx, ipfsHash)
	assert.NotNil(t, err)
}

func TestClient_sendRequest(t *testing.T) {
	ctx := context.Background()
	httpMethod := http.MethodPost
	path := "/path"
	jwtToken := "some_token"
	reqBody := []byte("req_body")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)

		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+jwtToken, authHeader)

		assert.Equal(t, httpMethod, r.Method)

		assert.NotNil(t, r.Body)
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, reqBody, b)
	}))

	defer srv.Close()

	c := &client{
		log:      logging.Logger("ipfs-pinning-test"),
		apiURL:   srv.URL,
		JWTToken: jwtToken,
		c:        http.DefaultClient,
	}

	res, err := c.sendRequest(ctx, httpMethod, path, bytes.NewReader(reqBody))
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestClient_sendRequest_invalidReq(t *testing.T) {
	ctx := context.Background()
	httpMethod := "/invalidMethod"
	path := "/path"
	jwtToken := "some_token"

	c := &client{
		log:      logging.Logger("ipfs-pinning-test"),
		apiURL:   "https://centrifuge.io",
		JWTToken: jwtToken,
		c:        http.DefaultClient,
	}

	res, err := c.sendRequest(ctx, httpMethod, path, nil)
	assert.ErrorIs(t, err, ErrHTTPRequestCreation)
	assert.Nil(t, res)
}

func TestClient_sendRequest_requestError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	httpMethod := http.MethodPost
	path := "/path"
	jwtToken := "some_token"
	reqBody := []byte("req_body")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	defer srv.Close()

	c := &client{
		log:      logging.Logger("ipfs-pinning-test"),
		apiURL:   srv.URL,
		JWTToken: jwtToken,
		c:        http.DefaultClient,
	}

	cancel()

	res, err := c.sendRequest(ctx, httpMethod, path, bytes.NewReader(reqBody))
	assert.ErrorIs(t, err, ErrHTTPRequest)
	assert.Nil(t, res)
}

func TestClient_handleResponse(t *testing.T) {
	c := &client{log: logging.Logger("ipfs-pinning-test")}

	res := new(http.Response)
	res.Body = nil

	// No response body
	err := c.handleResponse(res, nil)
	assert.NoError(t, err)

	resBody := PinJSONToIPFSResponse{
		IpfsHash:  "hash",
		PinSize:   4,
		Timestamp: time.Now(),
	}

	b, err := json.Marshal(resBody)
	assert.NoError(t, err)

	res.StatusCode = 200
	res.Body = io.NopCloser(bytes.NewReader(b))

	// No response object to unmarshal
	err = c.handleResponse(res, nil)
	assert.NoError(t, err)

	var r PinJSONToIPFSResponse

	/// Response object present
	res.Body = io.NopCloser(bytes.NewReader(b))

	err = c.handleResponse(res, &r)
	assert.NoError(t, err)

	assert.Equal(t, resBody.IpfsHash, r.IpfsHash)
	assert.Equal(t, resBody.PinSize, r.PinSize)
	assert.Equal(t, resBody.Timestamp.Format(time.RFC3339), r.Timestamp.Format(time.RFC3339))
}

func TestClient_handleResponse_errors(t *testing.T) {
	c := &client{log: logging.Logger("ipfs-pinning-test")}

	res := new(http.Response)

	// Response body read error
	res.Body = io.NopCloser(&errReader{})

	err := c.handleResponse(res, nil)
	assert.ErrorIs(t, err, ErrHTTPResponseBodyRead)

	// Response status code >= 400
	res.Body = io.NopCloser(bytes.NewReader([]byte("some_bytes")))
	res.StatusCode = http.StatusBadRequest

	err = c.handleResponse(res, nil)
	assert.ErrorIs(t, err, ErrHTTPResponse)

	// Response object unmarshal
	res.Body = io.NopCloser(bytes.NewReader([]byte("some_bytes")))
	res.StatusCode = http.StatusOK

	var r PinJSONToIPFSResponse

	err = c.handleResponse(res, &r)
	assert.ErrorIs(t, err, ErrResponseJSONUnmarshal)
}
