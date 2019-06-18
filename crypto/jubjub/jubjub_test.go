package jubjub

import (
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
	msg := []byte("signIt")

	sig, err := Sign(sk, msg)
	assert.NoError(t, err)

	assert.True(t, Verify(pk, msg, sig))

	wrongMsg := []byte("wrong")
	assert.False(t, Verify(pk, wrongMsg, sig))
}
