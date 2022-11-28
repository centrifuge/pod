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
	controller *behavior.Controller
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelDebug)

	var err error

	controller, err = behavior.NewController()

	if err != nil {
		panic(fmt.Errorf("couldn't create new behaviour controller: %w", err))
	}

	if err := controller.Start(); err != nil {
		panic(fmt.Errorf("couldn't start behaviour controller: %w", err))
	}

	result := m.Run()

	if err := controller.Stop(); err != nil {
		panic(fmt.Errorf("couldn't stop behaviour controller: %w", err))
	}

	os.Exit(result)
}
