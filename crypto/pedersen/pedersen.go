package pedersen

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/centrifuge/go-centrifuge/errors"
	"hash"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Requires to follow instructions to install pycrypto: https://github.com/stefandeml/zokrates-pycrypto
// And either set CENT_PED_BIN with cli.py location or store the repo in $HOME/zksnarks/pycrypto

// NewPedersenHash returns a pedersen hash function
func NewPedersenHash() hash.Hash {
	// Resolve python binary location
	pedBin := os.Getenv("CENT_PED_BIN")
	if pedBin == "" {
		pedBin = fmt.Sprintf("%s/zksnarks/zokrates-pycrypto/cli.py", os.Getenv("HOME"))
	}

	args := []string{pedBin, "run_hash"}
	cmd := exec.Command("python3", args...)

	in, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	errp, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	return &pedersen{in: in, out: out, err: errp, hasher: cmd}
}

type pedersen struct {
	in io.WriteCloser
	out io.ReadCloser
	err io.ReadCloser
	hasher *exec.Cmd
	hashed      []byte
}

func (pd *pedersen) Reset() {
	pd.hashed = []byte{}
}

func (pd *pedersen) Write(p []byte) (n int, err error) {
	if len(p) != 64 {
		return 0, errors.New("invalid payload length")
	}

	shex := hex.EncodeToString(p)+"\n"
	_, err = pd.in.Write([]byte(shex))
	if err != nil {
		return 0, err
	}

	reader := bufio.NewReader(pd.out)
	scanner := bufio.NewScanner(reader)
	var hTxt string
	for scanner.Scan() {
		hTxt = scanner.Text()
		hTxt = strings.Replace(hTxt, "?", "", -1)
		hTxt = strings.Replace(hTxt, "\n", "", -1)
		if hTxt != "" {
			break
		}
	}

	if hTxt == "" {
		return 0, errors.New("Error reading subprocess output")
	}


	pd.hashed, err = hex.DecodeString(hTxt)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (pd *pedersen) Sum(b []byte) []byte {
	return append(pd.hashed, b...)
}

func (pd *pedersen) Size() int {
	return len(pd.hashed)
}

func (pd *pedersen) BlockSize() int {
	return 0
}
