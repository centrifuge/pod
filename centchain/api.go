package centchain

import (
	"sync"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/client"
	"github.com/centrifuge/go-substrate-rpc-client/rpc/author"
	"github.com/centrifuge/go-substrate-rpc-client/signature"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

// API exposes required functions to interact with Centrifuge Chain.
type API interface {
	// GetMetadataLatest returns latest metadata from the centrifuge chain.
	GetMetadataLatest() (*types.Metadata, error)

	// SubmitExtrinsic signs the given call with the provided KeyRingPair and submits an extrinsic.
	// Returns transaction hash, latest block number before extrinsic submission, and signature attached with the extrinsic.
	SubmitExtrinsic(meta *types.Metadata, c types.Call, krp signature.KeyringPair) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error)
}

type api struct {
	getBlockHash            func(uint64) (types.Hash, error)
	getRuntimeVersionLatest func() (*types.RuntimeVersion, error)
	getStorageLatest        func(key types.StorageKey, target interface{}) error
	getClient               func() client.Client
	getBlockLatest          func() (*types.SignedBlock, error)
	getMetadataLatest       func() (*types.Metadata, error)
	mu                      sync.Mutex
}

// NewAPI returns a new centrifuge chain api.
func NewAPI(sapi *gsrpc.SubstrateAPI) API {
	return api{
		getBlockHash:            sapi.RPC.Chain.GetBlockHash,
		getRuntimeVersionLatest: sapi.RPC.State.GetRuntimeVersionLatest,
		getStorageLatest:        sapi.RPC.State.GetStorageLatest,
		getClient:               func() client.Client { return sapi.Client },
		getBlockLatest:          sapi.RPC.Chain.GetBlockLatest,
		getMetadataLatest:       sapi.RPC.State.GetMetadataLatest,
	}
}

func (a api) GetMetadataLatest() (*types.Metadata, error) {
	return a.getMetadataLatest()
}

func (a api) SubmitExtrinsic(meta *types.Metadata, c types.Call, krp signature.KeyringPair) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	ext := types.NewExtrinsic(c)
	era := types.ExtrinsicEra{IsMortalEra: false}

	genesisHash, err := a.getBlockHash(0)
	if err != nil {
		return txHash, bn, sig, err
	}

	rv, err := a.getRuntimeVersionLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	key, err := types.CreateStorageKey(meta, "System", "AccountNonce", krp.PublicKey)
	if err != nil {
		return txHash, bn, sig, err
	}

	var nonce uint32
	err = a.getStorageLatest(key, &nonce)
	if err != nil {
		return txHash, bn, sig, err
	}

	o := types.SignatureOptions{
		BlockHash:   genesisHash,
		Era:         era,
		GenesisHash: genesisHash,
		Nonce:       types.UCompact(nonce),
		SpecVersion: rv.SpecVersion,
		Tip:         0,
	}

	err = ext.Sign(krp, o)
	if err != nil {
		return txHash, bn, sig, err
	}

	auth := author.NewAuthor(a.getClient())
	startBlock, err := a.getBlockLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	startBlockNumber := startBlock.Block.Header.Number
	txHash, err = auth.SubmitExtrinsic(ext)
	return txHash, startBlockNumber, ext.Signature.Signature, err
}
