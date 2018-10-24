package utils

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

func ToTimestamp(time time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: int64(time.Second()),
		Nanos:   int32(time.Nanosecond()),
	}
}
