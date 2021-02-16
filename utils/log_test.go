// +build unit

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gologging "github.com/whyrusleeping/go-logging"
)

func TestGetCentLogFormat(t *testing.T) {
	logFormat := GetCentLogFormat()

	format := gologging.MustStringFormatter(logFormat)
	assert.NotNil(t, format, "formatter should not be nil")
}
