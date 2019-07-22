package jubjub

import (
	"fmt"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateSigningKeyPair(t *testing.T) {
	pk, sk, err := GenerateSigningKeyPair()
	assert.NoError(t, err)

	assert.Len(t, pk, 32)
	assert.Len(t, sk, 32)
}

func TestSign(t *testing.T) {
	_, sk, err := GenerateSigningKeyPair()
	assert.NoError(t, err)
	msg := []byte("signIt")

	sig, err := Sign(sk, msg)
	assert.NoError(t, err)

	assert.Len(t, sig, 64)
}

func TestVerify(t *testing.T) {
	pk, sk, err := GenerateSigningKeyPair()
	assert.NoError(t, err)
	b32, err := utils.SliceToByte32(sk)
	assert.NoError(t, err)
	bbpk := babyjub.PrivateKey(b32)
	msg := []byte("signIt")

	sig, err := SignPython(bbpk[:], msg)
	assert.NoError(t, err)
	fmt.Println("SignPyCrypto", hexutil.Encode(sig))
	sig, err = Sign(bbpk[:], msg)
	assert.NoError(t, err)
	fmt.Println("LenBabyJub", len(sig))
	fmt.Println("SignBabyJub", hexutil.Encode(sig))
	assert.True(t, Verify(pk, msg, sig))

	//wrongMsg := []byte("wrong")
	//assert.False(t, Verify(pk, wrongMsg, sig))
}
