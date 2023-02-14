//go:build integration || testworld

package centchain

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type TestClient struct {
	api         *gsrpc.SubstrateAPI
	meta        *types.Metadata
	rv          *types.RuntimeVersion
	genesisHash types.Hash
}

func NewTestClient(centChainURL string) (*TestClient, error) {
	api, err := gsrpc.NewSubstrateAPI(centChainURL)

	if err != nil {
		return nil, fmt.Errorf("couldn't get substrate API: %w", err)
	}

	meta, err := api.RPC.State.GetMetadataLatest()

	if err != nil {
		return nil, fmt.Errorf("couldn't get latest metadata: %w", err)
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return nil, fmt.Errorf("couldn't get latest runtime version: %w", err)
	}

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return nil, fmt.Errorf("couldn't get genesis hash: %w", err)
	}

	return &TestClient{
		api,
		meta,
		rv,
		genesisHash,
	}, nil
}

func (f *TestClient) GetEvents(blockHash types.Hash) (*Events, error) {
	key, err := types.CreateStorageKey(f.meta, "System", "Events")

	if err != nil {
		return nil, err
	}

	var eventsRaw types.EventRecordsRaw

	ok, err := f.api.RPC.State.GetStorage(key, &eventsRaw, blockHash)

	if err != nil || !ok {
		return nil, errors.New("no events found in storage")
	}

	var events Events

	if err = eventsRaw.DecodeEventRecords(f.meta, &events); err != nil {
		return nil, err
	}

	return &events, nil
}

func (f *TestClient) Close() {
	f.api.Client.Close()
}

const (
	submitTransferInterval = 1 * time.Second
)

func (f *TestClient) SubmitAndWait(ctx context.Context, senderKrp signature.KeyringPair, fn CallProviderFn) (*types.Hash, error) {
	ticker := time.NewTicker(submitTransferInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context done while submitting transfer: %w", ctx.Err())
		case <-ticker.C:
			blockHash, err := f.submitExtrinsic(ctx, senderKrp, fn)

			if err == nil {
				return blockHash, nil
			}

			log.Errorf("Couldn't submit extrinsic: %s", err)
		}
	}
}

func (f *TestClient) submitExtrinsic(ctx context.Context, senderKrp signature.KeyringPair, fn CallProviderFn) (*types.Hash, error) {
	call, err := fn(f.meta)

	if err != nil {
		return nil, fmt.Errorf("couldn't create call: %w", err)
	}

	accountInfo, err := f.getAccountInfo(senderKrp.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account info: %w", err)
	}

	ext := types.NewExtrinsic(*call)

	signOpts := types.SignatureOptions{
		BlockHash:          f.genesisHash, // using genesis since we're using immortal era
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        f.genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(accountInfo.Nonce)),
		SpecVersion:        f.rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: f.rv.TransactionVersion,
	}

	if err := ext.Sign(senderKrp, signOpts); err != nil {
		return nil, fmt.Errorf("couldn't sign extrinsic: %w", err)
	}

	sub, err := f.api.RPC.Author.SubmitAndWatchExtrinsic(ext)

	if err != nil {
		return nil, fmt.Errorf("couldn't submit and watch extrinsic: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context done while waiting for extrinsic to be in block: %w", ctx.Err())
		case st := <-sub.Chan():
			ms, _ := st.MarshalJSON()

			log.Info("Got extrinsic status - ", string(ms))

			switch {
			case st.IsInBlock:
				return &st.AsInBlock, nil
			case st.IsUsurped:
				return nil, errors.New("extrinsic was usurped")
			}
		}
	}
}

func (f *TestClient) getAccountInfo(accountID []byte) (*types.AccountInfo, error) {
	storageKey, err := types.CreateStorageKey(f.meta, "System", "Account", accountID)

	if err != nil {
		return nil, err
	}

	var accountInfo types.AccountInfo

	ok, err := f.api.RPC.State.GetStorageLatest(storageKey, &accountInfo)

	if err != nil || !ok {
		return nil, errors.New("couldn't retrieve account info")
	}

	return &accountInfo, nil
}
