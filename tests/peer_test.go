// +build integration

package tests

import (
	"testing"
	//"context"
	"time"
)


// TODO remember to cleanup config files generated

func TestPeer_Start(t *testing.T) {
	//cancCtx, canc := context.WithCancel(context.Background())
	err := NewPeer("Alice", "ws://127.0.0.1:9546", "keystore", "", "russianhill", 8084, 38204, nil, true).Init()
	t.Error(err)

	//go NewPeer("Bob", "peerconfigs/bob.yaml").Start(cancCtx)


	time.Sleep(9 * time.Second)
	//canc()

}