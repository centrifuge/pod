// +build ethereum

package repository_test

import (
	"math/big"
	"os"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor/repository"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var identityService identity.IdentityService

// Add Key
var testAddress string
var testPrivateKey string

func TestMain(m *testing.M) {

	identityService = &identity.EthereumIdentityService{}
	cc.TestFunctionalEthereumBootstrap()
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func createIdentityWithKeys(t *testing.T) []byte {
	centrifugeId := tools.RandomSlice(identity.CentIdByteLength)
	id, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.GetCentrifugeId(), "Resulting Identity should have the same ID as the input")

	testAddress = "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d"
	testPrivateKey = "0xb5fffc3933d93dc956772c69b42c4bc66123631a24e3465956d80b5b604a2d13"

	confirmations, err = id.AddKeyToIdentity(2, utils.HexToByteArray(testAddress))

	return centrifugeId

}

/*
func TestMessageConcatSign(t *testing.T){
	anchorId := tools.RandomSlice(32)
	documentRoot := tools.RandomSlice(32)
	centrifugeId := tools.RandomSlice(32)

	fmt.Println(utils.ByteArrayToHex(anchorId))
	fmt.Println(utils.ByteArrayToHex(documentRoot))
	fmt.Println(utils.ByteArrayToHex(centrifugeId))


	var message []byte

	message = append(anchorId, documentRoot...)

	message = append(message, centrifugeId...)

	fmt.Println(utils.ByteArrayToHex(message))


}
*/

func TestCommitAnchor_Integration(t *testing.T) {
	anchorId := tools.RandomSlice(32)
	documentRoot := tools.RandomSlice(32)
	documentProof := tools.RandomByte32()
	signature := tools.RandomSlice(65)

	var documentProofs [][32]byte

	documentProofs = append(documentProofs, documentProof)

	centrifugeId := createIdentityWithKeys(t)

	var message []byte

	message = append(anchorId, documentRoot...)

	message = append(message, centrifugeId...)

	signature = secp256k1.SignEthereum(message, utils.HexToByteArray(testPrivateKey))

	var anchorIdBigInt big.Int
	anchorIdBigInt.SetBytes(anchorId)

	var centrifugeIdBigInt big.Int
	centrifugeIdBigInt.SetBytes(centrifugeId)

	var documentRoot32Bytes [32]byte
	copy(documentRoot32Bytes[:], documentRoot[:32])

	confirmations, err := repository.CommitAnchor(&anchorIdBigInt, documentRoot32Bytes, &centrifugeIdBigInt, documentProofs, signature)
	if err != nil {
		t.Fatalf("Error commit Anchor %v", err)
	}

	watchCommittedAnchor := <-confirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error thrown by context")

}
