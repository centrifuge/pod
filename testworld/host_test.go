// +build testworld

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

func TestHost_Start(t *testing.T) {
	cancCtx, canc := context.WithCancel(context.Background())
	alice := createHost("Alice", 8084, 38204)
	err := alice.Init()
	if err != nil {
		t.Error(err)
	}

	bob := createHost("Bob", 8085, 38205)
	err = bob.Init()
	if err != nil {
		t.Error(err)
	}

	go alice.Start(cancCtx)
	go bob.Start(cancCtx)

	a, err := alice.id()
	if err != nil {
		t.Error(err)
	}
	b, err := bob.id()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("CentID for Alice %s \n", a)
	fmt.Printf("CentID for Bob %s \n", b)
	// TODO tests on alice and bob
	// Eg:
	alice.CreateInvoice(invoice.Invoice{}, []string{bob.name})

	time.Sleep(100 * time.Second)
	canc()
}

func createHost(name string, apiPort, p2pPort int64) *host {
	return newHost(
		name,
		"ws://127.0.0.1:9546",
		"keystore", "", "testing", apiPort, p2pPort, nil, true,
		contractAddresses)
}
