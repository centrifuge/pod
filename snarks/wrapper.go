package snarks

import (
	"encoding/json"
	"fmt"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
)

type BuyerProof struct {
	RootHash string
	Right []bool
	Hashes []string
	Value string
}

//type UniPoint struct {
//
//}
//
//type BiPoint struct {
//
//}

type Proof struct {
	A []string `json:"a"`
	B [][]string `json:"b"`
	C []string `json:"c"`
}

type ZKProof struct {
	Proof Proof `json:"proof"`
	Inputs []string `json:"inputs"`
}

func CallNFTCrypto(payload []byte, outDir string) error {
	// Copy zkPayload as out/proof_data.json
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		err = os.Mkdir(outDir, 0755)
		if err != nil {
			return err
		}
	}
	err := ioutil.WriteFile(outDir+"/proof_data.json", payload, 0644)
	if err != nil {
		return err
	}
	// Call nft.py
	nftBin := fmt.Sprintf("%s/zksnarks/zokrates-pycrypto/nft.py", os.Getenv("HOME"))
	err = exec.Command("python3", nftBin).Run()
	if err != nil {
		return err
	}
	// Copy witness to Zokrates demo dir
	zkDemoDir := fmt.Sprintf("%s/zksnarks/ZoKrates/demo/", os.Getenv("HOME"))

	return exec.Command("cp", outDir+"/nft_witness.txt", zkDemoDir).Run()
}

func GetZokratesProofs() (ret [8]*big.Int, err error) {
	zkBase := fmt.Sprintf("%s/zksnarks/ZoKrates", os.Getenv("HOME"))
	err = os.Setenv("ZOKRATES_HOME",zkBase+"/zokrates_stdlib/stdlib")
	if err != nil {
		return ret, err
	}
	cmd := exec.Command("/bin/sh", "run.sh")
	cmd.Dir = zkBase+"/demo"
	err = cmd.Run()
	if err != nil {
		return ret, err
	}

	var zk *ZKProof
	data, err := ioutil.ReadFile(zkBase+"/demo/proof.json")
	err = json.Unmarshal(data, &zk)
	if err != nil {
		return ret, err
	}

	p0b, err := hexutil.Decode(zk.Proof.A[0])
	ret[0] = utils.ByteSliceToBigInt(p0b)
	p0b, err = hexutil.Decode(zk.Proof.A[1])
	ret[1] = utils.ByteSliceToBigInt(p0b)
	p0b, err = hexutil.Decode(zk.Proof.B[0][0])
	ret[2] = utils.ByteSliceToBigInt(p0b)
	p0b, err = hexutil.Decode(zk.Proof.B[0][1])
	ret[3] = utils.ByteSliceToBigInt(p0b)
	p0b, err = hexutil.Decode(zk.Proof.B[1][0])
	ret[4] = utils.ByteSliceToBigInt(p0b)
	p0b, err = hexutil.Decode(zk.Proof.B[1][1])
	ret[5] = utils.ByteSliceToBigInt(p0b)
	p0b, err = hexutil.Decode(zk.Proof.C[0])
	ret[6] = utils.ByteSliceToBigInt(p0b)
	p0b, err = hexutil.Decode(zk.Proof.C[1])
	ret[7] = utils.ByteSliceToBigInt(p0b)

	return ret, nil
}

func GenerateDefaultBuyerRatingProof(buyerID, buyerPublicKey string) (*BuyerProof, error) {
	creditBase := fmt.Sprintf("%s/creditrating", os.Getenv("HOME"))
	type buyer struct {
		Buyer1 string `json:"BUYER1"`
		Buyer1Pub string `json:"BUYER1_PUB"`
		Buyer1Rating string `json:"BUYER1_RATING"`
		Buyer2 string `json:"BUYER2"`
		Buyer2Pub string `json:"BUYER2_PUB"`
		Buyer2Rating string `json:"BUYER2_RATING"`
		Buyer3 string `json:"BUYER3"`
		Buyer3Pub string `json:"BUYER3_PUB"`
		Buyer3Rating string `json:"BUYER3_RATING"`
		Buyer4 string `json:"BUYER4"`
		Buyer4Pub string `json:"BUYER4_PUB"`
		Buyer4Rating string `json:"BUYER4_RATING"`
	}
	by := buyer{
		Buyer1: buyerID,
		Buyer1Pub: buyerPublicKey,
		Buyer1Rating: "64",
		Buyer2: "0000000000000000000000000000000000000002",
		Buyer2Pub: "ac11bb576922e4929b20d8475d952d8322f857b7d6fbd5356a3797ad02963d10",
		Buyer2Rating: "0a",
		Buyer3: "0000000000000000000000000000000000000003",
		Buyer3Pub: "1ac6bc29acacc0b36ffbc69dd96a0faee3275f95f73a8b06088792fc24c1eb8c",
		Buyer3Rating: "0a",
		Buyer4: "0000000000000000000000000000000000000004",
		Buyer4Pub: "9ca2f5f6f64e3bf5d1c05ef325adf3666f6bcd8a6e44434ce5074b56462241bf",
		Buyer4Rating: "60",
	}
	overrideJSON, err := json.MarshalIndent(by, "", "  ")
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile("ratings.json", overrideJSON, 0644)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("/bin/sh", creditBase+"/credit_rating.sh")
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	type buyerMap struct {
		CreditRatingRootHash string `json:"credit_rating_roothash"`
		Buyers map[string]BuyerProof `json:"buyers"`
	}
	var bm *buyerMap
	data, err := ioutil.ReadFile("out/credit_rating.json")
	err = json.Unmarshal(data, &bm)
	if err != nil {
		return nil, err
	}
	buyerProof := &BuyerProof{
		RootHash: bm.CreditRatingRootHash,
		Right: bm.Buyers[buyerID].Right,
		Hashes: bm.Buyers[buyerID].Hashes,
		Value: bm.Buyers[buyerID].Value,
	}
	return buyerProof, nil
}
