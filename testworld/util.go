// +build testworld

package testworld

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"encoding/json"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/savaki/jq"
)

// runSmartContractMigrations migrates smart contracts to localgeth
func runSmartContractMigrations() {
	projDir := getProjectDir()
	migrationScript := path.Join(projDir, "build", "scripts", "migrate.sh")
	_, err := exec.Command(migrationScript, projDir).Output()
	if err != nil {
		log.Fatal(err)
	}
}

// getSmartContractAddresses finds migrated smart contract addresses for localgeth
func getSmartContractAddresses() *config.SmartContractAddresses {
	projDir := getProjectDir()
	deployJSONFile := path.Join(projDir, "vendor", "github.com", "centrifuge", "centrifuge-ethereum-contracts", "deployments", "localgeth.json")
	dat, err := ioutil.ReadFile(deployJSONFile)
	if err != nil {
		panic(err)
	}
	idFactoryAddrOp := getOpForContract(".contracts.IdentityFactory.address")
	idRegistryAddrOp := getOpForContract(".contracts.IdentityRegistry.address")
	anchorRepoAddrOp := getOpForContract(".contracts.AnchorRepository.address")
	payObAddrOp := getOpForContract(".contracts.PaymentObligation.address")
	return &config.SmartContractAddresses{
		IdentityFactoryAddr:   getOpAddr(idFactoryAddrOp, dat),
		IdentityRegistryAddr:  getOpAddr(idRegistryAddrOp, dat),
		AnchorRepositoryAddr:  getOpAddr(anchorRepoAddrOp, dat),
		PaymentObligationAddr: getOpAddr(payObAddrOp, dat),
	}
}

func getOpAddr(addrOp jq.Op, dat []byte) string {
	addr, err := addrOp.Apply(dat)
	if err != nil {
		panic(err)
	}

	// remove annoying quotes
	addrStr := string(addr)
	if len(addrStr) > 0 && addrStr[0] == '"' {
		addrStr = addrStr[1:]
	}
	if len(addrStr) > 0 && addrStr[len(addrStr)-1] == '"' {
		addrStr = addrStr[:len(addrStr)-1]
	}
	return addrStr
}

func getOpForContract(selector string) jq.Op {
	addrOp, err := jq.Parse(selector)
	if err != nil {
		panic(err)
	}
	return addrOp
}

func getProjectDir() string {
	gp := os.Getenv("GOPATH")
	projDir := path.Join(gp, "src", "github.com", "centrifuge", "go-centrifuge")
	return projDir
}

// ToMap converts a struct to a map using the struct's tags.
//
// ToMap uses tags on struct fields to decide which fields to add to the
// returned map.
func toMap(in interface{}) (map[string]interface{}, error) {
	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(in)
	json.Unmarshal(inrec, &inInterface)
	return map[string]interface{}{
		"data": map[string]interface{}{
			"invoice_number": "12324",
			"due_date":       "2018-09-26T23:12:37.902198664Z",
			"gross_amount":   "40",
			"currency":       "GBP",
			"net_amount":     "40",
		},
		"collaborators": []string{"0x24fe6555beb9"}}, nil
}
