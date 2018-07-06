package utils

import (
	"net/http"
	"bytes"
	"fmt"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("http-utils")

func SendPOSTRequest(url string, contentType string, payload []byte) (statusCode int, err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", contentType)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("%v", resp.Status)
	}
	statusCode = resp.StatusCode
	return
}