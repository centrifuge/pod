package utils

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// ToTimestamp converts time.Time to timestamp.TimeStamp.
func ToTimestamp(time time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: int64(time.Second()),
		Nanos:   int32(time.Nanosecond()),
	}
}

// ToTimestampProper converts time.Time to timestamp.TimeStamp.
func ToTimestampProper(time time.Time) (*timestamp.Timestamp, error) {
	return ptypes.TimestampProto(time)
}

// FromTimestamp converts a timestamp protobuf to time
func FromTimestamp(t *timestamp.Timestamp) (time.Time, error) {
	return ptypes.Timestamp(t)
}
