// +build testworld

package testworld

import (
	"testing"

	"fmt"
	"os"
)

// TODO remember to cleanup config files generated

var doctorFord *hostManager
var configFile string = "configs/local.json"

func TestMain(m *testing.M) {
	c, err := loadConfig(configFile)
	if err != nil {
		panic(err)
	}
	// TODO start POA geth here
	//runSmartContractMigrations()
	contractAddresses := getSmartContractAddresses()
	doctorFord = newHostManager(c.EthNodeUrl, c.AccountKeyPath, c.AccountPassword, c.Network, c.TxPoolAccess, contractAddresses)
	err = doctorFord.init()
	if err != nil {
		panic(err)
	}
	fmt.Printf("contract addresses %+v\n", contractAddresses)
	result := m.Run()
	doctorFord.stop()
	os.Exit(result)
}

func TestHost_Happy(t *testing.T) {
	alice := doctorFord.getHost("Alice")
	bob := doctorFord.getHost("Bob")
	charlie := doctorFord.getHost("Charlie")
	eAlice := alice.createHttpExpectation(t)
	eBob := bob.createHttpExpectation(t)
	eCharlie := charlie.createHttpExpectation(t)

	b, err := bob.id()
	if err != nil {
		t.Error(err)
	}

	c, err := charlie.id()
	if err != nil {
		t.Error(err)
	}
	res, err := alice.createInvoice(eAlice, map[string]interface{}{
		"data": map[string]interface{}{
			"invoice_number": "12324",
			"due_date":       "2018-09-26T23:12:37.902198664Z",
			"gross_amount":   "40",
			"currency":       "GBP",
			"net_amount":     "40",
		},
		"collaborators": []string{b.String(), c.String()},
	})
	if err != nil {
		t.Error(err)
	}
	docIdentifier := res.Value("header").Path("$.document_id").String().NotEmpty().Raw()
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	getInvoiceAndCheck(eBob, docIdentifier, "GBP")
	getInvoiceAndCheck(eCharlie, docIdentifier, "GBP")
	fmt.Println("Host test success")
}
