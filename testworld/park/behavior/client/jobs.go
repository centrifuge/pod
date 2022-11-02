//go:build testworld

package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gavv/httpexpect"
)

const (
	jobCompletionTimeout = 15 * time.Minute
)

func (c *Client) WaitForJobCompletion(jobID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), jobCompletionTimeout)
	defer cancel()

	log.Infof("Waiting for job notification - %s", jobID)

	if err := c.webhookReceiver.WaitForJobCompletion(ctx, jobID); err != nil {

		if jobErrorMsg := c.getJobErrorMessage(jobID); jobErrorMsg != "" {
			log.Errorf("Job error message: %s", jobErrorMsg)
		}

		return fmt.Errorf("couldn't wait for job %s completion: %w", jobID, err)
	}

	if jobErrorMsg := c.getJobErrorMessage(jobID); jobErrorMsg != "" {
		return fmt.Errorf("job error: %s", jobErrorMsg)
	}

	return nil
}

func (c *Client) getJobErrorMessage(jobID string) string {
	resp := addCommonHeaders(c.expect.GET("/v2/jobs/"+jobID), c.jwtToken).Expect().Status(200).JSON().Object()
	task := resp.Value("tasks").Array().Last().Object()
	message := task.Value("error").String().Raw()

	return message
}

func GetJobID(resp *httpexpect.Object) (string, error) {
	jobID := resp.Value("header").Path("$.job_id").String().Raw()

	if jobID != "" {
		return jobID, nil
	}

	return "", errors.New("job ID not found")
}
