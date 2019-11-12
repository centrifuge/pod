package centchain

import (
	"github.com/centrifuge/go-substrate-rpc-client/client"
	"github.com/centrifuge/go-substrate-rpc-client/rpc/author"
	"github.com/centrifuge/go-substrate-rpc-client/signature"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

// API exposes required functions to interact with Centrifuge Chain.
type API struct {
	GetBlockHash            func(uint64) (types.Hash, error)
	GetRuntimeVersionLatest func() (*types.RuntimeVersion, error)
	GetStorageLatest        func(key types.StorageKey, target interface{}) error
	GetClient               func() client.Client
	GetBlockLatest          func() (*types.SignedBlock, error)
	GetMetadataLatest       func() (*types.Metadata, error)
}

// SubmitExtrinsic signs the given call with the provided KeyRingPair and submits an extrinsic.
// Returns transaction hash, latest block number before extrinsic submission, and signature attached with the extrinsic.
func (a *API) SubmitExtrinsic(meta *types.Metadata, c types.Call, krp signature.KeyringPair) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error) {
	ext := types.NewExtrinsic(c)
	era := types.ExtrinsicEra{IsMortalEra: false}

	genesisHash, err := a.GetBlockHash(0)
	if err != nil {
		return txHash, bn, sig, err
	}

	rv, err := a.GetRuntimeVersionLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	key, err := types.CreateStorageKey(meta, "System", "AccountNonce", krp.PublicKey)
	if err != nil {
		return txHash, bn, sig, err
	}

	var nonce uint32
	err = a.GetStorageLatest(key, &nonce)
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

	auth := author.NewAuthor(a.GetClient())
	startBlock, err := a.GetBlockLatest()
	if err != nil {
		return txHash, bn, sig, err
	}

	startBlockNumber := startBlock.Block.Header.Number
	txHash, err = auth.SubmitExtrinsic(ext)
	return txHash, startBlockNumber, ext.Signature.Signature, err
}
