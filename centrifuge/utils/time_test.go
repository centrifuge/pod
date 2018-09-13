package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToTimestamp(t *testing.T) {
	now := time.Now().UTC()
	ts := ToTimestamp(now)
	assert.NotNil(t, ts, "must be non nil")
	assert.Equal(t, now.Second(), int(ts.Seconds))
	assert.Equal(t, now.Nanosecond(), int(ts.Nanos))
}
