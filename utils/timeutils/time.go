package timeutils

import (
	"time"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/ptypes/timestamp"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("time-util")

// EnsureDelayOperation delays the execution by the opDelay provided, used for return statements
func EnsureDelayOperation(start time.Time, opDelay time.Duration) {
	consumed := time.Since(start)
	if consumed < opDelay {
		t := time.NewTimer(opDelay - consumed)
		<-t.C
		t.Stop()
	}
	apiLog.Infof("Time consumed by operation [%s]", consumed.String())
	apiLog.Infof("Real Response Time of operation [%s]", time.Since(start).String())
}

// ToProtoTimestamps converts time.Time to timestamp.Timestamps
func ToProtoTimestamps(tms ...*time.Time) ([]*timestamp.Timestamp, error) {
	if len(tms) < 1 {
		return nil, nil
	}

	pts := make([]*timestamp.Timestamp, len(tms))
	for i, t := range tms {
		if t == nil {
			pts[i] = nil
			continue
		}

		pt, err := utils.ToTimestamp(*t)
		if err != nil {
			return nil, err
		}

		pts[i] = pt
	}

	return pts, nil
}

// FromProtoTimestamps converts timestamp.Timestamps
func FromProtoTimestamps(pts ...*timestamp.Timestamp) ([]*time.Time, error) {
	if len(pts) < 1 {
		return nil, nil
	}

	tms := make([]*time.Time, len(pts))
	for i, pt := range pts {
		if pt == nil {
			tms[i] = nil
			continue
		}

		tm, err := utils.FromTimestamp(pt)
		if err != nil {
			return nil, err
		}

		tms[i] = &tm
	}

	return tms, nil
}
