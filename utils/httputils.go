package utils

import (
	"fmt"
	logging "github.com/ipfs/go-log"
	"gopkg.in/resty.v1"
)

var log = logging.Logger("http-utils")

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
		err = fmt.Errorf("%v", resp.Status)
	}
	statusCode = resp.StatusCode()
	return
}