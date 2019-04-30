package timeutils

import (
	"time"

	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("time-util")

// EnsureDelayOperation delays the execution by the opDelay provided, used for return statements
func EnsureDelayOperation(start time.Time, opDelay time.Duration) {
	consumed := time.Now().Sub(start)
	if consumed < opDelay {
		t := time.NewTimer(opDelay - consumed)
		<-t.C
	}
	apiLog.Infof("Time consumed by operation [%s]", consumed.String())
	apiLog.Infof("Real Response Time of operation [%s]", time.Now().Sub(start).String())
}
