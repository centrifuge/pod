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
	assert.True(t, time.Now().Sub(start) >= opDelay)

	// Longer than threshold
	start = time.Now()
	tm := time.NewTimer(500 * time.Millisecond)
	<-tm.C
	EnsureDelayOperation(start, opDelay)
	assert.True(t, time.Now().Sub(start) >= opDelay)
}

func TestProtoTimestamps(t *testing.T) {
	// zero length
	pts, err := ToProtoTimestamps()
	assert.NoError(t, err)
	assert.Nil(t, pts)
	ntms, err := FromProtoTimestamps()
	assert.NoError(t, err)
	assert.Nil(t, ntms)

	// with values
	tm := time.Now().UTC()
	tms := []*time.Time{nil, &tm}
	pts, err = ToProtoTimestamps(tms...)
	assert.NoError(t, err)
	ntms, err = FromProtoTimestamps(pts...)
	assert.NoError(t, err)
	assert.Equal(t, tms, ntms)

	// error
	pts[1].Seconds = 253402300800
	_, err = FromProtoTimestamps(pts...)
	assert.Error(t, err)
}
