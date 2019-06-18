package jubjub

import (
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"strings"
)

// GenerateSigningKeyPair generates a EDDSA JubJub key pair
func GenerateSigningKeyPair() (publicKey, privateKey []byte, err error) {
	// Resolve python binary location
	pedBin := os.Getenv("CENT_PED_BIN")
	if pedBin == "" {
		pedBin = fmt.Sprintf("%s/zksnarks/zokrates-pycrypto/cli.py", os.Getenv("HOME"))
	}

	args := []string{pedBin, "keygen"}
	o, err := exec.Command("python3", args...).CombinedOutput()
	if err != nil {
		return nil, nil, err
	}

	keys := strings.Split(string(o), " ")
	if len(keys) < 2 {
		return nil, nil, errors.New("Wrong key format generated")
	}

	cleanPKStr := strings.Replace(keys[1], "\n", "", -1)

	privateKey, err = hex.DecodeString(ensureEvenHexLength(keys[0]))
	if err !=  nil {
		return nil,nil, err
	}
	publicKey, err = hex.DecodeString(ensureEvenHexLength(cleanPKStr))
	if err !=  nil {
		return nil,nil, err
	}

	return
}

func Sign(privateKey, message []byte) (signature []byte, err error) {
	// Resolve python binary location
	pedBin := os.Getenv("CENT_PED_BIN")
	if pedBin == "" {
		pedBin = fmt.Sprintf("%s/zksnarks/zokrates-pycrypto/cli.py", os.Getenv("HOME"))
	}

	args := []string{pedBin, "sig-gen", hex.EncodeToString(privateKey), hex.EncodeToString(message)}
	o, err := exec.Command("python3", args...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	sig := strings.Split(string(o), " ")
	if len(sig) < 2 {
		return nil, errors.New("Wrong signature format generated")
	}

	cleanSig2 := strings.Replace(sig[1], "\n", "", -1)
	sig1, err := hex.DecodeString(ensureEvenHexLength(sig[0]))
	if err != nil {
		return nil, err
	}

	sig2, err := hex.DecodeString(ensureEvenHexLength(cleanSig2))
	if err != nil {
		return nil, err
	}

	signature = append(sig1, sig2...)
	return

}

func Verify(publicKey, message, signature []byte) bool {
	if len(signature) != 64 {
		return false
	}

	sig1 := signature[:32]
	sig2 := signature[32:]

	// Resolve python binary location
	pedBin := os.Getenv("CENT_PED_BIN")
	if pedBin == "" {
		pedBin = fmt.Sprintf("%s/zksnarks/zokrates-pycrypto/cli.py", os.Getenv("HOME"))
	}

	args := []string{pedBin, "sig-verify", hex.EncodeToString(publicKey), hex.EncodeToString(message), hex.EncodeToString(sig1), hex.EncodeToString(sig2)}
	err := exec.Command("python3", args...).Run()
	if err != nil {
		return false
	}

	return true
}


func ensureEvenHexLength(input string) (output string) {
	output = input
	if len(input) % 2 != 0 {
		output = "0"+input
	}
	return
}
