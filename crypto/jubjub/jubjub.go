package jubjub

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"math/big"
	"os"
	"os/exec"
	"strings"
)

// GenerateSigningKeyPair generates a EDDSA BabyJub key pair
func GeneratePythonSigningKeyPair() (publicKey, privateKey []byte, err error) {
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

// GenerateSigningKeyPair generates a EDDSA BabyJub key pair
func GenerateSigningKeyPair() (publicKey, privateKey []byte, err error) {
	rpk := babyjub.NewRandPrivKey()
	pkComp := rpk.Public().Compress()
	publicKey = pkComp[:]
	privateKey = rpk.Scalar().BigInt().Bytes()

	//// Resolve python binary location
	//pedBin := os.Getenv("CENT_PED_BIN")
	//if pedBin == "" {
	//	pedBin = fmt.Sprintf("%s/zksnarks/zokrates-pycrypto/cli.py", os.Getenv("HOME"))
	//}
	//
	//args := []string{pedBin, "keygen"}
	//o, err := exec.Command("python3", args...).CombinedOutput()
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//keys := strings.Split(string(o), " ")
	//if len(keys) < 2 {
	//	return nil, nil, errors.New("Wrong key format generated")
	//}
	//
	//cleanPKStr := strings.Replace(keys[1], "\n", "", -1)
	//
	//privateKey, err = hex.DecodeString(ensureEvenHexLength(keys[0]))
	//if err !=  nil {
	//	return nil,nil, err
	//}
	//publicKey, err = hex.DecodeString(ensureEvenHexLength(cleanPKStr))
	//if err !=  nil {
	//	return nil,nil, err
	//}

	return
}

func SignPython(privateKey, message []byte) (signature []byte, err error) {
	//Resolve python binary location
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

func Sign(privateKey, message []byte) (signature []byte, err error) {
	b32, err := utils.SliceToByte32(privateKey)
	if err != nil {
		return nil, err
	}
	pk := babyjub.PrivateKey(b32)
	mBI := &big.Int{}
	mBI = mBI.SetBytes(message)
	s := pk.SignMimc7(mBI)
	fmt.Println("R8X_BabyJub", hexutil.Encode(s.R8.X.Bytes()))
	fmt.Println("R8Y_BabyJub", hexutil.Encode(s.R8.Y.Bytes()))
	sComp := s.Compress()
	signature = sComp[:]

	// Resolve python binary location
	//pedBin := os.Getenv("CENT_PED_BIN")
	//if pedBin == "" {
	//	pedBin = fmt.Sprintf("%s/zksnarks/zokrates-pycrypto/cli.py", os.Getenv("HOME"))
	//}
	//
	//args := []string{pedBin, "sig-gen", hex.EncodeToString(privateKey), hex.EncodeToString(message)}
	//o, err := exec.Command("python3", args...).CombinedOutput()
	//if err != nil {
	//	return nil, err
	//}
	//
	//sig := strings.Split(string(o), " ")
	//if len(sig) < 2 {
	//	return nil, errors.New("Wrong signature format generated")
	//}
	//
	//cleanSig2 := strings.Replace(sig[1], "\n", "", -1)
	//sig1, err := hex.DecodeString(ensureEvenHexLength(sig[0]))
	//if err != nil {
	//	return nil, err
	//}
	//
	//sig2, err := hex.DecodeString(ensureEvenHexLength(cleanSig2))
	//if err != nil {
	//	return nil, err
	//}
	//
	//signature = append(sig1, sig2...)
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
