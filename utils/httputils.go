package utils

import (
	"crypto/tls"
	"net"

	"github.com/centrifuge/go-centrifuge/errors"
	"gopkg.in/resty.v1"
)

// SendPOSTRequest sends post with data to given URL.
func SendPOSTRequest(url string, contentType string, payload []byte) (statusCode int, err error) {
	c := resty.New()
	cfg := &tls.Config{InsecureSkipVerify: true} // Temporary until we have defined a cert truststore
	c.SetTLSClientConfig(cfg)

	resp, err := c.R().
		SetHeader("Content-Type", contentType).
		SetBody(payload).
		Post(url)

	if err != nil {
		return statusCode, err
	}

	return resp.StatusCode(), nil
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
