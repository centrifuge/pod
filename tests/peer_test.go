// +build p2p_test

package tests

import (
	"testing"
	//"context"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
)

// TODO remember to cleanup config files generated

var contractAddresses *config.SmartContractAddresses

func TestMain(m *testing.M) {
	// TODO start POA geth here
	runSmartContractMigrations()
	contractAddresses = getSmartContractAddresses()
	fmt.Printf("contract addresses %+v\n", contractAddresses)
	result := m.Run()
	os.Exit(result)
}

func TestPeer_Start(t *testing.T) {
	cancCtx, canc := context.WithCancel(context.Background())
	alice := createPeer("Alice", 8084, 38204)
	err := alice.Init()
	if err != nil {
		t.Error(err)
	}

	bob := createPeer("Bob", 8084, 38204)
	err = bob.Init()
	if err != nil {
		t.Error(err)
	}

	alice.Start(cancCtx)
	bob.Start(cancCtx)
	// TODO tests on alice and bob
	// Eg:
	alice.CreateInvoice(invoice.Invoice{}, []string{bob.name})

	time.Sleep(100 * time.Second)
	canc()
}

func createPeer(name string, apiPort, p2pPort int64) *peer {
	return NewPeer(
		name,
		"ws://127.0.0.1:9546",
		"keystore", "", "testing", apiPort, p2pPort, nil, true,
		contractAddresses)
}
