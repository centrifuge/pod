//go:build integration || testworld

package testingutils

import (
	"os/exec"
)

// IsCentChainRunning checks if POS centchain is running in the background
func IsCentChainRunning() bool {
	cmd := `docker ps -a --filter "name=cc-alice" --filter "status=running" --quiet`
	o, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		panic(err)
	}
	return len(o) != 0
}
