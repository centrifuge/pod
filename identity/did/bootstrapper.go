package did

import (
	"fmt"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/savaki/jq"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// BootstrappedDIDFactory stores the id of the factory
const BootstrappedDIDFactory string = "BootstrappedDIDFactory"

// BootstrappedDIDService stores the id of the service
const BootstrappedDIDService string = "BootstrappedDIDService"

var smartContractAddresses *config.SmartContractAddresses

// Bootstrap initializes the factory contract
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	// we have to allow loading from file in case this is coming from create config cmd where we don't add configs to db
	cfg, err := configstore.RetrieveConfig(false, context)
	if err != nil {
		return err
	}

	if _, ok := context[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}
	client := context[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	factoryAddress := common.HexToAddress("0x0490f72BB148e511d7b66ac6476e3E7A8A2e2740")
	getFactoryAddress(cfg)



	factoryContract, err := bindFactory(factoryAddress, client)
	if err != nil {
		return err
	}

	txManager, ok := context[transactions.BootstrappedService].(transactions.Manager)
	if !ok {
		return errors.New("transactions repository not initialised")
	}

	queueSrv, ok := context[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue hasn't been initialized")
	}

	factory := NewFactory(factoryContract, client, txManager, queueSrv,factoryAddress)
	context[BootstrappedDIDFactory] = factory

	service := NewService(client, txManager, queueSrv)
	context[BootstrappedDIDService] = service

	return nil
}

func bindFactory(factoryAddress common.Address, client ethereum.Client) (*FactoryContract, error) {
	return NewFactoryContract(factoryAddress, client.GetEthClient())
}

func getFactoryAddress(cfg config.Configuration) common.Address {
	fmt.Println("factory contract")
	fmt.Println(cfg.GetContractAddress(config.IdentityFactory))
	return cfg.GetContractAddress(config.IdentityFactory)

}

// Note: this block will be removed after the identity migration is done
// currently we are using two versions of centrifuge contracts to not break the compatiblitiy
// ---------------------------------------------------------------------------------------------------------------------
func migrateNewIdentityContracts() {
	runNewSmartContractMigrations()
	smartContractAddresses = GetSmartContractAddresses()

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
func GetSmartContractAddresses() *config.SmartContractAddresses {
	dat, err := findContractDeployJSON()
	if err != nil {
		panic(err)
	}
	idFactoryAddrOp := getOpForContract(".contracts.IdentityFactory.address")
	anchorRepoAddrOp := getOpForContract(".contracts.AnchorRepository.address")
	payObAddrOp := getOpForContract(".contracts.PaymentObligation.address")
	return &config.SmartContractAddresses{
		IdentityFactoryAddr:   getOpField(idFactoryAddrOp, dat),
		AnchorRepositoryAddr:  getOpField(anchorRepoAddrOp, dat),
		PaymentObligationAddr: getOpField(payObAddrOp, dat),
	}
}
func getFileFromContractRepo(filePath string) ([]byte, error) {
	gp := os.Getenv("GOPATH")
	projDir := path.Join(gp, "src", "github.com", "centrifuge", "go-centrifuge")
	deployJSONFile := path.Join(projDir, "vendor", "github.com", "centrifuge", "centrifuge-ethereum-contracts", filePath)
	dat, err := ioutil.ReadFile(deployJSONFile)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// TODO: func will be refactored after migration
func getIdentityByteCode() string {
	dat, err := findContractDeployJSON()
	if err != nil {
		panic(err)
	}
	optByte := getOpForContract(".contracts.Identity.bytecode")
	byteCodeHex := getOpField(optByte, dat)
	return byteCodeHex

}

// TODO: func will be removed after migration
func findContractDeployJSON() ([]byte, error) {
	return getFileFromContractRepo(path.Join("deployments", "localgeth.json"))
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
func getOpField(addrOp jq.Op, dat []byte) string {
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
