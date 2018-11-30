// +build testworld

package testworld

import (
	"testing"
	//"context"
	"context"
	"fmt"
	"os"

	"time"

	"github.com/centrifuge/go-centrifuge/config"
)

// TODO remember to cleanup config files generated

var contractAddresses *config.SmartContractAddresses
var bernard *host

func TestMain(m *testing.M) {
	// TODO start POA geth here
	//runSmartContractMigrations()
	contractAddresses = getSmartContractAddresses()
	cancCtx, canc := context.WithCancel(context.Background())
	bernard = createHost("Bernard", 8081, 38201, nil)
	err := bernard.init()
	if err != nil {
		panic(err)
	}
	go bernard.start(cancCtx)
	time.Sleep(30 * time.Second)
	fmt.Printf("contract addresses %+v\n", contractAddresses)
	result := m.Run()
	canc()
	os.Exit(result)
}

func TestHost_Start(t *testing.T) {
	bootnode, err := bernard.p2pURL()
	if err != nil {
		t.Error(err)
	}
	cancCtx, canc := context.WithCancel(context.Background())
	alice := createHost("Alice", 8084, 38204, []string{bootnode})
	err = alice.init()
	if err != nil {
		t.Error(err)
	}

	bob := createHost("Bob", 8085, 38205, []string{bootnode})
	err = bob.init()
	if err != nil {
		t.Error(err)
	}

	// host initialisation sequence and wait
	go alice.start(cancCtx)
	time.Sleep(10 * time.Second)
	go bob.start(cancCtx)
	time.Sleep(10 * time.Second)

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

	eAlice := alice.createHttpExpectation(t)
	eBob := bob.createHttpExpectation(t)
	res, err := alice.createInvoice(eAlice, map[string]interface{}{
		"data": map[string]interface{}{
			"invoice_number": "12324",
			"due_date":       "2018-09-26T23:12:37.902198664Z",
			"gross_amount":   "40",
			"currency":       "GBP",
			"net_amount":     "40",
		},
		"collaborators": []string{b.String()},
	})
	if err != nil {
		t.Error(err)
	}
	docIdentifier := res.Value("header").Path("$.document_id").String().NotEmpty().Raw()
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	getInvoiceAndCheck(eBob, docIdentifier, "GBP")
	fmt.Println("Host test success")
	canc()
}

func createHost(name string, apiPort, p2pPort int64, bootstraps []string) *host {
	// TODO make configs selectable as settings for different networks, eg Kovan + parity
	return newHost(
		name,
		"ws://127.0.0.1:9546",
		"keystore", "", "testing", apiPort, p2pPort, bootstraps, true,
		contractAddresses)
}
