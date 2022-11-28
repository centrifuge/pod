package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnsureDelayOperation(t *testing.T) {
	// Shorter than threshold
	start := time.Now()
	opDelay := 10 * time.Millisecond
	EnsureDelayOperation(start, opDelay)
	assert.True(t, time.Since(start) >= opDelay)

	// Longer than threshold
	start = time.Now()
	tm := time.NewTimer(500 * time.Millisecond)
	<-tm.C
	EnsureDelayOperation(start, opDelay)
	assert.True(t, time.Since(start) >= opDelay)
}
