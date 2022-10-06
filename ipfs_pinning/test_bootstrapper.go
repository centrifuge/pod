//go:build integration || testworld

package ipfs_pinning

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	//nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/path"
	mh "github.com/multiformats/go-multihash"
)

var testServer *httptest.Server

func (b *Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {

	testServer = httptest.NewServer(http.HandlerFunc(handlePinRequest))

	pinningService, err := NewPinataServiceClient(
		testServer.URL,
		"test-auth",
	)

	if err != nil {
		return fmt.Errorf("couldn't create pinning service client: %w", err)
	}

	ctx[BootstrappedIPFSPinningService] = pinningService

	return nil
}

func (*Bootstrapper) TestTearDown() error {
	testServer.Close()

	return nil
}

func handlePinRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(fmt.Errorf("couldn't read pin request body: %w", err))
	}

	var pinReq PinJSONToIPFSRequest

	if err := json.Unmarshal(reqBody, &pinReq); err != nil {
		panic(fmt.Errorf("couldn't unmarshal pin request: %w", err))
	}

	contentBytes, err := json.Marshal(pinReq.PinataContent)

	if err != nil {
		panic(fmt.Errorf("couldn't read pin request body: %w", err))
	}

	var nftMeta NFTMetadata

	if err = json.Unmarshal(contentBytes, &nftMeta); err != nil {
		panic(fmt.Errorf("couldn't unmarshal NFT metadata: %w", err))
	}

	nftMetaBytes, err := json.Marshal(nftMeta)

	if err = json.Unmarshal(contentBytes, &nftMeta); err != nil {
		panic(fmt.Errorf("couldn't marshal NFT metadata: %w", err))
	}

	v1CidPrefix := cid.Prefix{
		Codec:    cid.Raw,
		MhLength: -1,
		MhType:   mh.SHA2_256,
		Version:  1,
	}

	metadataCID, err := v1CidPrefix.Sum(nftMetaBytes)

	if err != nil {
		panic(fmt.Errorf("couldn't create metadata CID: %w", err))
	}

	metaPath := path.New(metadataCID.String())

	pinRes := PinJSONToIPFSResponse{
		IpfsHash:  metaPath.String(),
		PinSize:   len(nftMetaBytes),
		Timestamp: time.Now(),
	}

	response, err := json.Marshal(pinRes)

	if err != nil {
		panic(fmt.Errorf("couldn't marshal request: %1111w", err))
	}

	w.Write(response)
	w.WriteHeader(http.StatusOK)
}
