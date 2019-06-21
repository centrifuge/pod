package snarks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type BuyerProof struct {
	RootHash string
	Right []bool
	Hashes []string
	Value string
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

func CallZokrates() error {
	zkBase := fmt.Sprintf("%s/zksnarks/ZoKrates", os.Getenv("HOME"))
	err := os.Setenv("ZOKRATES_HOME",zkBase+"/zokrates_stdlib/stdlib")
	if err != nil {
		return err
	}
	cmd := exec.Command("/bin/sh", "run.sh")
	cmd.Dir = zkBase+"/demo"
	return cmd.Run()
}

func GenerateDefaultBuyerRatingProof(buyerID, buyerPublicKey string) (*BuyerProof, error) {
	creditBase := fmt.Sprintf("%s/creditrating", os.Getenv("HOME"))
	type buyer struct {
		Buyer1 string `json:"BUYER1"`
		Buyer1Pub string `json:"BUYER1_PUB"`
		Buyer1Rating string `json:"BUYER1_RATING"`
	}
	by := buyer{
		Buyer1: buyerID,
		Buyer1Pub: buyerPublicKey,
		Buyer1Rating: "64",
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
