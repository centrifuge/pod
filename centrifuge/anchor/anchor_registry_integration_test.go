package anchor_test

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor"
	"github.com/stretchr/testify/assert"
	"flag"
	"os"
	"github.com/spf13/viper"
)

var (
	ethereumTest = flag.Bool("ethereum", false, "run Ethereum integration tests")
)

func TestMain(m *testing.M) {
	flag.Parse()

	//for now set up the env vars manually in integration test
	//TODO move to generalized config once it is available
	viper.BindEnv("ethereum.gethipc", "CENT_ETHEREUM_GETHIPC")
	viper.BindEnv("ethereum.gasLimit", "CENT_ETHEREUM_GASLIMIT")
	viper.BindEnv("ethereum.gasPrice", "CENT_ETHEREUM_GASPRICE")
	viper.BindEnv("ethereum.contextWaitTimeout", "CENT_ETHEREUM_CONTEXTWAITTIMEOUT")
	viper.BindEnv("ethereum.accounts.main.password", "CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD")
	viper.BindEnv("ethereum.accounts.main.key", "CENT_ETHEREUM_ACCOUNTS_MAIN_KEY")
	viper.BindEnv("anchor.ethereum.anchorRegistryAddress", "CENT_ANCHOR_ETHEREUM_ANCHORREGISTRYADDRESS")

	result := m.Run()
	os.Exit(result)
}

func TestRegisterAsAnchor_Integration(t *testing.T) {
	if !*ethereumTest{
		return
	}
	confirmations := make(chan *anchor.Anchor,1)
	id := tools.RandomString32()
	rootHash := tools.RandomString32()
	err := anchor.RegisterAsAnchor(id, rootHash, confirmations)
	if err != nil {
		t.Fatalf("Error registering Anchor %v", err)
	}

	registeredAnchor := <-confirmations
	assert.Equal(t, registeredAnchor.AnchorID, id, "Resulting anchor should have the same ID as the input")
	assert.Equal(t, registeredAnchor.RootHash, rootHash, "Resulting anchor should have the same root hash as the input")
}


// TODO make this work in high concurrency
//func TestRegisterAsAnchor_Integration_Concurrency(t *testing.T) {
//	if !*ethereumTest {
//		return
//	}
//
//	var submittedIds [5]string
//	var submittedRhs [5]string
//
//	howMany := cap(submittedIds)
//	confirmations := make(chan *anchor.Anchor, howMany)
//
//	for ix := 0; ix < howMany; ix++ {
//		id := tools.RandomString32()
//		submittedIds[ix] = id
//
//		rootHash := tools.RandomString32()
//		submittedRhs[ix] = rootHash
//
//		err := anchor.RegisterAsAnchor(id, rootHash, confirmations)
//		assert.Nil(t, err, "should not error out upon anchor registration")
//	}
//
//	for ix := 0; ix < howMany; ix++ {
//		singleAnchor := <-confirmations
//		assert.Contains(t, submittedIds, singleAnchor.AnchorID , "Should have the ID that was passed into create function [%v]", singleAnchor.AnchorID)
//		assert.Contains(t, submittedRhs, singleAnchor.RootHash , "Should have the RootHash that was passed into create function [%v]", singleAnchor.RootHash)
//	}
//}
