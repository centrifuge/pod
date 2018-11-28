// +build p2p_test

package tests

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

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
