//go:build testworld

package testworld

import (
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/testworld/park/behavior"
	logging "github.com/ipfs/go-log"
)

var (
	head *behavior.Head
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelDebug)

	var err error

	head, err = behavior.NewHead()

	if err != nil {
		panic(fmt.Errorf("couldn't create new behaviour head: %w", err))
	}

	if err := head.Start(); err != nil {
		panic(fmt.Errorf("couldn't start behaviour head: %w", err))
	}

	result := m.Run()

	if err := head.Stop(); err != nil {
		panic(fmt.Errorf("couldn't stop behaviour head: %w", err))
	}

	os.Exit(result)
}
