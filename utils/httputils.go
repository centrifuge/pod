package utils

import (
	"github.com/go-errors/errors"
	logging "github.com/ipfs/go-log"
	"gopkg.in/resty.v1"
)

var log = logging.Logger("http-utils")

// SendPOSTRequest sends post with data to given URL.
func SendPOSTRequest(url string, contentType string, payload []byte) (statusCode int, err error) {
	resp, err := resty.R().
		SetHeader("Content-Type", contentType).
		SetBody(payload).
		Post(url)

	if err != nil {
		log.Error(err)
		return
	}
	if resp.StatusCode() != 200 {
		err = errors.Errorf("%s", resp.Status())
	}
	statusCode = resp.StatusCode()
	return
}
