package ideth

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/config/configstore"

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

	factoryAddress := getFactoryAddress(cfg)

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

	factory := NewFactory(factoryContract, client, txManager, queueSrv, factoryAddress)
	context[identity.BootstrappedDIDFactory] = factory

	service := NewService(client, txManager, queueSrv)
	context[identity.BootstrappedDIDService] = service

	return nil
}

func bindFactory(factoryAddress common.Address, client ethereum.Client) (*FactoryContract, error) {
	return NewFactoryContract(factoryAddress, client.GetEthClient())
}

func getFactoryAddress(cfg config.Configuration) common.Address {
	return cfg.GetContractAddress(config.IdentityFactory)
}

// TODO: store identity bytecode in config and remove func
func getIdentityByteCode() string {
	filePath := path.Join("deployments", "localgeth.json")
	gp := os.Getenv("GOPATH")
	projDir := path.Join(gp, "src", "github.com", "centrifuge", "go-centrifuge")
	deployJSONFile := path.Join(projDir, "vendor", "github.com", "centrifuge", "centrifuge-ethereum-contracts", filePath)
	dat, err := ioutil.ReadFile(deployJSONFile)

	if err != nil {
		panic(err)
	}

	optByte, err := jq.Parse(".contracts.Identity.bytecode")
	if err != nil {
		panic(err)
	}
	byteCodeHex := getOpField(optByte, dat)
	return byteCodeHex

}

// TODO: store identity bytecode in config and remove func
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
