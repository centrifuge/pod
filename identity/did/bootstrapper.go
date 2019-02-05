package did

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/savaki/jq"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// BootstrappedDIDService stores the id of the service
const BootstrappedDIDService string = "BootstrappedDIDService"

var smartContractAddresses *config.SmartContractAddresses

// Bootstrap initializes the factory contract
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	cfg, err := configstore.RetrieveConfig(false, context)
	if err != nil {
		return err
	}

	if _, ok := context[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}
	client := context[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	// TODO line will be removed after migration
	migrateNewIdentityContracts()

	factoryContract, err := bindFactory(getFactoryAddress(), client)
	if err != nil {
		return err
	}

	txManager, ok := context[transactions.BootstrappedService].(transactions.Manager)
	if !ok {
		return errors.New("transactions repository not initialised")
	}

	if _, ok := context[bootstrap.BootstrappedQueueServer]; !ok {
		return errors.New("queue hasn't been initialized")
	}
	queueSrv := context[bootstrap.BootstrappedQueueServer].(*queue.Server)


	service := NewService(cfg, factoryContract, client, txManager,queueSrv)
	context[BootstrappedDIDService] = service
	return nil
}

func bindFactory(factoryAddress common.Address, client ethereum.Client) (*FactoryContract, error) {
	return NewFactoryContract(factoryAddress, client.GetEthClient())
}

func getFactoryAddress() common.Address {
	return common.HexToAddress(smartContractAddresses.IdentityFactoryAddr)

}

func getAnchorAddress() common.Address {
	return common.HexToAddress(smartContractAddresses.AnchorRepositoryAddr)
}

// Note: this block will be removed after the identity migration is done
// currently we are using two versions of centrifuge contracts to not break the compatiblitiy
// ---------------------------------------------------------------------------------------------------------------------
func migrateNewIdentityContracts() {
	//runNewSmartContractMigrations()
	smartContractAddresses = getSmartContractAddresses()

}

// RunNewSmartContractMigrations migrates smart contracts to localgeth
// TODO: func will be removed after migration
func runNewSmartContractMigrations() {

	gp := os.Getenv("GOPATH")
	projDir := path.Join(gp, "src", "github.com", "centrifuge", "go-centrifuge")

	smartContractDir := path.Join(projDir, "vendor", "github.com", "manuelpolzhofer", "centrifuge-ethereum-contracts")
	smartContractDirStandard := path.Join(projDir, "vendor", "github.com", "centrifuge", "centrifuge-ethereum-contracts")

	os.Setenv("CENT_ETHEREUM_CONTRACTS_DIR", smartContractDir)

	migrationScript := path.Join(projDir, "build", "scripts", "migrate.sh")
	_, err := exec.Command(migrationScript, projDir).Output()
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("CENT_ETHEREUM_CONTRACTS_DIR", smartContractDirStandard)

}

// GetSmartContractAddresses finds migrated smart contract addresses for localgeth
// TODO: func will be removed after migration
func getSmartContractAddresses() *config.SmartContractAddresses {
	dat, err := findContractDeployJSON()
	if err != nil {
		panic(err)
	}
	idFactoryAddrOp := getOpForContract(".contracts.IdentityFactory.address")
	anchorRepoAddrOp := getOpForContract(".contracts.AnchorRepository.address")
	payObAddrOp := getOpForContract(".contracts.PaymentObligation.address")
	return &config.SmartContractAddresses{
		IdentityFactoryAddr:   getOpAddr(idFactoryAddrOp, dat),
		AnchorRepositoryAddr:  getOpAddr(anchorRepoAddrOp, dat),
		PaymentObligationAddr: getOpAddr(payObAddrOp, dat),
	}
}

// TODO: func will be removed after migration
func findContractDeployJSON() ([]byte, error) {
	gp := os.Getenv("GOPATH")
	projDir := path.Join(gp, "src", "github.com", "centrifuge", "go-centrifuge")
	deployJSONFile := path.Join(projDir, "vendor", "github.com", "manuelpolzhofer", "centrifuge-ethereum-contracts", "deployments", "localgeth.json")
	dat, err := ioutil.ReadFile(deployJSONFile)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// TODO: func will be removed after migration
func getOpForContract(selector string) jq.Op {
	addrOp, err := jq.Parse(selector)
	if err != nil {
		panic(err)
	}
	return addrOp
}

// TODO: func will be removed after migration
func getOpAddr(addrOp jq.Op, dat []byte) string {
	addr, err := addrOp.Apply(dat)
	if err != nil {
		panic(err)
	}

	// remove extra quotes inside the string
	addrStr := string(addr)
	if len(addrStr) > 0 && addrStr[0] == '"' {
		addrStr = addrStr[1:]
	}
	if len(addrStr) > 0 && addrStr[len(addrStr)-1] == '"' {
		addrStr = addrStr[:len(addrStr)-1]
	}
	return addrStr
}

// ---------------------------------------------------------------------------------------------------------------------
