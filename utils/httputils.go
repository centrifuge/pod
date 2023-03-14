package utils

import (
	"bytes"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/centrifuge/pod/errors"
)

// SendPOSTRequest sends post with data to given URL.
func SendPOSTRequest(url string, contentType string, payload []byte) (statusCode int, err error) {
	c := http.Client{}
	cfg := &tls.Config{InsecureSkipVerify: true} // Temporary until we have defined a cert truststore
	c.Transport = &http.Transport{TLSClientConfig: cfg}
	resp, err := c.Post(url, contentType, bytes.NewReader(payload))
	if err != nil {
		return statusCode, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

// GetFreeAddrPort returns a loopback address and port that can be listened from.
// Note: port is included in the address.
func GetFreeAddrPort() (string, int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", 0, errors.New("failed to get a random free port")
	}

	defer l.Close()
	return l.Addr().String(), l.Addr().(*net.TCPAddr).Port, nil
}
