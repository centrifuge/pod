package pedersen

import (
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"os/exec"
)

// Requires to follow instructions to install pycrypto: https://github.com/stefandeml/zokrates-pycrypto
// And either set CENT_PED_BIN with cli.py location or store the repo in $HOME/zksnarks/pycrypto
func NewPedersenHash() hash.Hash {
	// Resolve python binary location
	pedBin := os.Getenv("CENT_PED_BIN")
	if pedBin == "" {
		pedBin = fmt.Sprintf("%s/zksnarks/pycrypto/cli.py",os.Getenv("HOME"))
	}
	return &pedersen{binLocation: pedBin}
}

type pedersen struct {
	binLocation  string
	hashed       []byte
}

func (pd *pedersen) Reset() {
	pd.hashed = []byte{}
}

func (pd *pedersen) Write(p []byte) (n int, err error) {
	args := []string{pd.binLocation, "hash", hex.EncodeToString(p)}
	o, err := exec.Command("python3", args...).Output()
	if err != nil {
		return 0, err
	}
	pd.hashed, err = hex.DecodeString(string(o[:len(o)-1])) // removing newline
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (pd *pedersen) Sum(b []byte) []byte {
	return pd.hashed
}

func (pd *pedersen) Size() int {
	return len(pd.hashed)
}

func (pd *pedersen) BlockSize() int {
	return 0
}
