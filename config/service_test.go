// +build unit

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_GetConfig_NoConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	cfg, err := svc.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}
