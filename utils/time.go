package utils

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// ToTimestamp converts time.Time to timestamp.TimeStamp.
// Ref: https://github.com/golang/protobuf/blob/master/ptypes/timestamp.go#L113
func ToTimestamp(time time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: time.Unix(),
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
