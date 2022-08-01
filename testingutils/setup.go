//go:build integration || unit || testworld

package testingutils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("test-setup")

// StartCentChain runs centchain for tests
func StartCentChain() {
	// don't run if its already running
	if IsCentChainRunning() {
		return
	}
	projDir := GetProjectDir()
	runScript := path.Join(projDir, "build", "scripts", "docker", "run.sh")
	o, err := exec.Command(runScript, "ccdev").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", string(o))
	time.Sleep(10 * time.Second)
}

func GetProjectDir() string {
	gp := os.Getenv("BASE_PATH")
	projDir := path.Join(gp, "centrifuge", "go-centrifuge")
	return projDir
}

// IsCentChainRunning checks if POS centchain is running in the background
func IsCentChainRunning() bool {
	cmd := `docker ps -a --filter "name=cc-alice" --filter "status=running" --quiet`
	o, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		panic(err)
	}
	return len(o) != 0
}
