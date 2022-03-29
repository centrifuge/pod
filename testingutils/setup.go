//go:build integration || unit || cmd || testworld
// +build integration unit cmd testworld

package testingutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	logging "github.com/ipfs/go-log"
	"github.com/savaki/jq"
)

var log = logging.Logger("test-setup")

var migrationsRan = os.Getenv("MIGRATION_RAN") == "true"

// StartPOAGeth runs the proof of authority geth for tests
func StartPOAGeth() {
	// don't run if its already running
	if IsPOAGethRunning() {
		return
	}
	projDir := GetProjectDir()
	gethRunScript := path.Join(projDir, "build", "scripts", "docker", "run.sh")
	o, err := exec.Command(gethRunScript, "dev").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", string(o))
	time.Sleep(10 * time.Second)
}

// StartCentChain runs centchain for tests
func StartCentChain() {
	// don't run if its already running
	if IsCentChainRunning() {
		return
	}
	projDir := GetProjectDir()
	runScript := path.Join(projDir, "build", "scripts", "docker", "run.sh")
	o, err := exec.Command(runScript, "ccdev").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", string(o))
	time.Sleep(10 * time.Second)
}

// StartBridge deploys contracts and run bridge
// if bridge is already running, this is a noop.
func StartBridge() {
	// don't run if its already running
	if IsBridgeRunning() {
		return
	}
	// run the bridge
	projDir := GetProjectDir()
	runScript := path.Join(projDir, "build", "scripts", "docker", "run.sh")
	o, err := exec.Command(runScript, "bridge").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", string(o))
	time.Sleep(10 * time.Second)
}

// RunSmartContractMigrations migrates smart contracts to localgeth
func RunSmartContractMigrations() {
	if migrationsRan {
		log.Infof("Not running migrations")
		return
	}

	var err error
	var out []byte
	projDir := GetProjectDir()
	migrationScript := path.Join(projDir, "build", "scripts", "migrate.sh")
	for i := 0; i < 3; i++ {
		fmt.Printf("Trying to migrate contracts for the %d th time\n", i)
		out, err = exec.Command(migrationScript, projDir).CombinedOutput()
		fmt.Println(string(out))
		if err == nil {
			err := os.Setenv("MIGRATIONS_RAN", "true")
			if err != nil {
				fmt.Println("Error setting MIGRATION_RAN flag on env, setting manually")
				migrationsRan = true
			}
			return
		}
	}

	// trying 3 times to migrate didnt work
	log.Fatal(err, string(out))
}

func DeployOracleContract(fingerprint, ward string) (string, error) {
	var err error
	var out []byte
	projDir := GetProjectDir()
	migrationScript := path.Join(projDir, "build", "scripts", "deploy_oracle.sh")
	fmt.Println("Trying to migrate oracle contracts")
	cmd := exec.Command(migrationScript, projDir)
	path := os.Getenv("PATH")
	cmd.Env = append(cmd.Env, fmt.Sprintf("FINGERPRINT=%s", fingerprint))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", path))
	cmd.Env = append(cmd.Env, fmt.Sprintf("OWNER=%s", ward))
	out, err = cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		return "", err
	}

	addrs := GetDAppSmartContractAddresses()
	return addrs["oracle"], nil
}

func GetDAppSmartContractAddresses() map[string]string {
	projDir := GetProjectDir()
	addresses := map[string]string{}
	b, err := ioutil.ReadFile(path.Join(projDir, "localAddresses"))
	if err != nil {
		return addresses
	}
	f := strings.TrimSpace(string(b))
	elems := strings.Split(f, "\n")
	for i := 0; i < len(elems); i++ {
		addrEntry := strings.Split(elems[i], " ")
		if len(addrEntry) < 2 {
			return addresses
		}
		addresses[addrEntry[0]] = addrEntry[1]
	}
	return addresses
}

// GetSmartContractAddresses finds migrated smart contract addresses for localgeth
func GetSmartContractAddresses() *config.SmartContractAddresses {
	iddat, err := findContractDeployJSON("IdentityFactory.json")
	if err != nil {
		panic(err)
	}

	addrOp := getOpForContract(".networks.1337.address")
	return &config.SmartContractAddresses{
		IdentityFactoryAddr: getOpAddr(addrOp, iddat),
	}
}

func findContractDeployJSON(file string) ([]byte, error) {
	projDir := GetProjectDir()
	deployJSONFile := path.Join(projDir, "build", "centrifuge-ethereum-contracts", "build", "contracts", file)
	dat, err := ioutil.ReadFile(deployJSONFile)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

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

func getOpForContract(selector string) jq.Op {
	addrOp, err := jq.Parse(selector)
	if err != nil {
		panic(err)
	}
	return addrOp
}

func GetProjectDir() string {
	gp := os.Getenv("BASE_PATH")
	projDir := path.Join(gp, "centrifuge", "go-centrifuge")
	return projDir
}

// IsPOAGethRunning checks if POA geth is running in the background
func IsPOAGethRunning() bool {
	cmd := "docker ps -a --filter \"name=geth-node\" --filter \"status=running\" --quiet"
	o, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		panic(err)
	}
	return len(o) != 0
}

// IsCentChainRunning checks if POS centchain is running in the background
func IsCentChainRunning() bool {
	cmd := "docker ps -a --filter \"name=cc-alice\" --filter \"status=running\" --quiet"
	o, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		panic(err)
	}
	return len(o) != 0
}

// IsBridgeRunning checks if bridge is running in the background
func IsBridgeRunning() bool {
	cmd := "docker ps -a --filter \"name=bridge\" --filter \"status=running\" --quiet"
	o, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		panic(err)
	}
	return len(o) != 0
}

// LoadTestConfig loads configuration for integration tests
func LoadTestConfig() config.Configuration {
	// To get the config location, we need to traverse the path to find the `go-centrifuge` folder
	projDir := GetProjectDir()
	c := config.LoadConfiguration(fmt.Sprintf("%s/build/configs/testing_config.yaml", projDir))
	return c
}

// SetupSmartContractAddresses sets up smart contract addresses on provided config
func SetupSmartContractAddresses(cfg config.Configuration, sca *config.SmartContractAddresses) {
	network := cfg.Get("centrifugeNetwork").(string)
	cfg.SetupSmartContractAddresses(network, sca)
	fmt.Printf("contract addresses %+v\n", sca)
}

// BuildIntegrationTestingContext sets up configuration for integration tests
func BuildIntegrationTestingContext() map[string]interface{} {
	projDir := GetProjectDir()
	StartPOAGeth()
	StartCentChain()
	RunSmartContractMigrations() // Running migrations so bridge addresses are generated before running bridge
	StartBridge()
	addresses := GetSmartContractAddresses()
	cfg := LoadTestConfig()
	cfg.Set("keys.p2p.publicKey", fmt.Sprintf("%s/build/resources/p2pKey.pub.pem", projDir))
	cfg.Set("keys.p2p.privateKey", fmt.Sprintf("%s/build/resources/p2pKey.key.pem", projDir))
	cfg.Set("keys.signing.publicKey", fmt.Sprintf("%s/build/resources/signingKey.pub.pem", projDir))
	cfg.Set("keys.signing.privateKey", fmt.Sprintf("%s/build/resources/signingKey.key.pem", projDir))
	SetupSmartContractAddresses(cfg, addresses)
	cm := make(map[string]interface{})
	cm[bootstrap.BootstrappedConfig] = cfg
	return cm
}
