package utils

import (
	"gopkg.in/resty.v1"
)

// SendPOSTRequest sends post with data to given URL.
func SendPOSTRequest(url string, contentType string, payload []byte) (statusCode int, err error) {
	resp, err := resty.R().
		SetHeader("Content-Type", contentType).
		SetBody(payload).
		Post(url)

	if err != nil {
		return statusCode, err
	}

	return resp.StatusCode(), nil
}
