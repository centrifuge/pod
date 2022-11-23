package crypto

import (
	"crypto/sha256"

	sr25519 "github.com/ChainSafe/go-schnorrkel"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/blake2b"
)

// GenerateSigningKeyPair generates based on the curveType and writes keys to file paths given.
func GenerateSigningKeyPair(publicFileName, privateFileName string, curveType CurveType) (err error) {
	var publicKey, privateKey []byte
	switch curveType {
	case CurveEd25519:
		publicKey, privateKey, err = ed25519.GenerateSigningKeyPair()
	case CurveSr25519:
		secretKey, pubKey, err := sr25519.GenerateKeypair()

		if err != nil {
			return err
		}

		encodedPubKey := pubKey.Encode()
		publicKey = encodedPubKey[:]

		encodedSecretKey := secretKey.Encode()
		privateKey = encodedSecretKey[:]
	default:
		publicKey, privateKey, err = ed25519.GenerateSigningKeyPair()
	}

	if err != nil {
		return err
	}

	err = utils.WriteKeyToPemFile(privateFileName, utils.PrivateKey, privateKey)
	if err != nil {
		return err
	}

	err = utils.WriteKeyToPemFile(publicFileName, utils.PublicKey, publicKey)
	if err != nil {
		return err
	}
	return nil
}

func GenerateSR25519SecretSeed() (string, error) {
	ms, err := sr25519.GenerateMiniSecretKey()
	if err != nil {
		return "", err
	}

	seed := ms.Encode()

	return hexutil.Encode(seed[:]), nil
}

func GenerateSR25519SeedsIfEmpty(seeds ...*string) error {
	for _, seed := range seeds {
		if *seed == "" {
			var err error

			*seed, err = GenerateSR25519SecretSeed()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GenerateHashPair generates a preimage and hash pair. This is useful in a commit reveal scheme such as what we use for anchor pre-commit > commit flow.
func GenerateHashPair(preimageSize int) (preimage, hash []byte, err error) {
	preimage = utils.RandomSlice(preimageSize)
	hash, err = Blake2bHash(preimage)
	return preimage, hash, err
}

// Blake2bHash wraps inconvenient blake2b_256 hashing ops
func Blake2bHash(value []byte) (hash []byte, err error) {
	h, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(value)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// Sha256Hash wraps inconvenient sha256 hashing ops
func Sha256Hash(value []byte) (hash []byte, err error) {
	h := sha256.New()
	_, err = h.Write(value)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
