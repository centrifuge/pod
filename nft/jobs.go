package nft

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	nftJob         = "Mint NFT Job"
	transferNFTJob = "Transfer NFT Job"
	nftOnCCJob     = "Mint NFT on CC"
)

// MintNFTJob mints and NFT async.
// args are as follows
// accountID, documentID, tokenID, MintNFTRequest
type MintNFTJob struct {
	jobs.Base
	accountsSrv config.Service
	docSrv      documents.Service
	dispatcher  jobs.Dispatcher
	ethClient   ethereum.Client
	api         API
	identitySrv identity.Service
}

// New returns a new instance of MintNFTJob
func (m *MintNFTJob) New() gocelery.Runner {
	nm := &MintNFTJob{
		accountsSrv: m.accountsSrv,
		docSrv:      m.docSrv,
		dispatcher:  m.dispatcher,
		ethClient:   m.ethClient,
		api:         m.api,
		identitySrv: m.identitySrv,
	}
	nm.Base = jobs.NewBase(nm.loadTasks())
	return nm
}

func (m *MintNFTJob) convertArgs(
	args []interface{}) (ctx context.Context, did identity.DID, docID []byte, tokenID TokenID, req MintNFTRequest,
	err error) {
	did = args[0].(identity.DID)
	tokenID = args[1].(TokenID)
	req = args[2].(MintNFTRequest)
	acc, err := m.accountsSrv.GetAccount(did[:])
	if err != nil {
		err = fmt.Errorf("failed to get account: %w", err)
		return
	}

	ctx = contextutil.WithAccount(context.Background(), acc)
	return ctx, did, req.DocumentID, tokenID, req, nil
}

func (m *MintNFTJob) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"add_nft_to_document": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, _, docID, tokenID, req, err := m.convertArgs(args)
				if err != nil {
					return nil, err
				}
				doc, err := m.docSrv.GetCurrentVersion(ctx, docID)
				if err != nil {
					return nil, fmt.Errorf("failed to get document: %w", err)
				}

				err = doc.AddNFT(req.GrantNFTReadAccess, req.RegistryAddress, tokenID[:], true)
				if err != nil {
					return nil, fmt.Errorf("failed to add nft to document: %w", err)
				}

				jobID, err := m.docSrv.Commit(ctx, doc)
				if err != nil {
					return nil, fmt.Errorf("failed to commit document: %w", err)
				}
				overrides["document_commit_job"] = jobID
				return nil, nil
			},
			Next: "wait_for_document_commit",
		},
		"wait_for_document_commit": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				did := args[0].(identity.DID)
				jobID := overrides["document_commit_job"].(gocelery.JobID)
				job, err := m.dispatcher.Job(did, jobID)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch job: %w", err)
				}

				if !job.IsSuccessful() {
					return nil, fmt.Errorf("document not committed yet")
				}

				return nil, nil
			},
			Next: "validate_nft_proofs",
		},
		"validate_nft_proofs": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, did, _, tokenID, req, err := m.convertArgs(args)
				if err != nil {
					return nil, err
				}

				requestData, err := prepareMintRequest(
					ctx, m.docSrv, tokenID, did, req.DocumentID, req.ProofFields, req.RegistryAddress,
					req.SubmitTokenProof, req.GrantNFTReadAccess, req.SubmitNFTReadAccessProof, req.DepositAddress)
				if err != nil {
					return nil, fmt.Errorf("failed to prepare mint request: %w", err)
				}

				subProofs := toSubstrateProofs(requestData.Props, requestData.Values, requestData.Salts, requestData.Proofs)
				staticProofs := [2][32]byte{requestData.DataRoot, requestData.SignaturesRoot}
				block, err := m.ethClient.GetEthClient().BlockByNumber(context.Background(), nil)
				if err != nil {
					return nil, fmt.Errorf("failed to get latest block: %w", err)
				}

				overrides["eth_from_block"] = block.Number()
				overrides["mint_request"] = requestData
				err = m.api.ValidateNFT(ctx, requestData.AnchorID, requestData.To, subProofs, staticProofs)
				if err != nil {
					return nil, fmt.Errorf("failed to validate nft Proofs: %w", err)
				}

				log.Infof("Successfully validated Proofs on cent chain for anchorID: %s", requestData.AnchorID.String())
				return nil, nil
			},
			Next: "wait_for_asset_deposit",
		},
		"wait_for_asset_deposit": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, _, _, _, req, err := m.convertArgs(args)
				if err != nil {
					return nil, err
				}

				if utils.IsEmptyAddress(req.AssetManagerAddress) {
					return nil, nil
				}

				from := overrides["eth_from_block"].(*big.Int)
				requestData := overrides["mint_request"].(MintRequest)
				log.Infof("Triggered listener on AssetManager Address %s for %s from %s", req.AssetManagerAddress.Hex(),
					hexutil.Encode(requestData.BundledHash[:]), from.String())
				err = ethereum.EventEmitted(
					ctx,
					m.ethClient.GetEthClient(),
					from,
					[]common.Address{req.AssetManagerAddress},
					AssetStoredEventSignature, requestData.BundledHash)
				if err != nil {
					return nil, err
				}

				log.Infof("Asset[%s] successfully deposited\n", hexutil.Encode(requestData.BundledHash[:]))
				return nil, nil
			},
			Next: "execute_mint_nft",
		},
		"execute_mint_nft": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, _, _, _, req, err := m.convertArgs(args)
				if err != nil {
					return nil, err
				}
				requestData := overrides["mint_request"].(MintRequest)

				// to common.Address, tokenId *big.Int, bytes32, properties [][]byte, values [][]byte, salts [][32]byte
				ethArgs := []interface{}{requestData.To, requestData.TokenID, requestData.SigningRoot,
					requestData.Props, requestData.Values, requestData.Salts}
				tx, err := m.identitySrv.ExecuteAsync(ctx, req.RegistryAddress, GenericMintMethodABI, "mint", ethArgs...)
				if err != nil {
					return nil, fmt.Errorf("failed to submit txn: %w", err)
				}

				log.Infof("Sent off ethTX[%s] to mint [tokenID: %s, To: %s, registry: %s, to NFT contract.",
					tx.Hash().Hex(),
					hexutil.Encode(requestData.TokenID.Bytes()),
					requestData.To.Hex(),
					req.RegistryAddress.String())
				overrides["mint_nft_txn"] = tx.Hash()
				return nil, nil
			},
			Next: "wait_mint_nft",
		},
		"wait_mint_nft": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				tx := overrides["mint_nft_txn"].(common.Hash)
				_, err = ethereum.IsTxnSuccessful(context.Background(), m.ethClient, tx)
				return nil, err
			},
			Next: "check_nft_owner",
		},
		"check_nft_owner": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				log.Infof("Verifying owner of the minted NFT...")
				tokenID := args[1].(TokenID)
				req := args[2].(MintNFTRequest)
				owner, err := ownerOf(m.ethClient, req.RegistryAddress, tokenID[:])
				if err != nil {
					return nil, err
				}

				if owner.Hex() != req.DepositAddress.Hex() {
					return nil, fmt.Errorf("owner for tokenID %s should be %s, instead got %s", tokenID.String(),
						req.DepositAddress.Hex(), owner.Hex())
				}

				log.Infof("Document %s minted successfully within transaction %s", hexutil.Encode(req.DocumentID), overrides["mint_nft_txn"])
				return nil, nil
			},
		},
	}
}

func initiateNFTMint(dispatcher jobs.Dispatcher, did identity.DID, tokenID TokenID,
	req MintNFTRequest) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Mint NFT", nftJob, "add_nft_to_document",
		[]interface{}{did, tokenID, req}, make(map[string]interface{}), time.Time{})
	_, err := dispatcher.Dispatch(did, job)
	if err != nil {
		return nil, fmt.Errorf("failed to dispatch mint NFT job: %w", err)
	}

	return job.ID, nil
}

// TransferNFTJob is a job runner for transferring NFT ownership
// args are as follows
// did(from), to, registry, tokenID
type TransferNFTJob struct {
	jobs.Base
	identitySrv identity.Service
	accountSrv  config.Service
	ethClient   ethereum.Client
}

// New returns a new instance of TransferNFTJob
func (t *TransferNFTJob) New() gocelery.Runner {
	nt := &TransferNFTJob{
		identitySrv: t.identitySrv,
		accountSrv:  t.accountSrv,
		ethClient:   t.ethClient,
	}

	nt.Base = jobs.NewBase(nt.loadTasks())
	return nt
}

func (t *TransferNFTJob) convertArgs(
	args []interface{}) (ctx context.Context, from, to, registry common.Address, tokenID TokenID, err error) {
	to, registry, tokenID = args[1].(common.Address), args[2].(common.Address), args[3].(TokenID)
	did := args[0].(identity.DID)
	acc, err := t.accountSrv.GetAccount(did[:])
	if err != nil {
		return ctx, from, to, registry, tokenID, err
	}

	return contextutil.WithAccount(context.Background(), acc), did.ToAddress(), to, registry, tokenID, nil
}

func (t *TransferNFTJob) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"transfer_ownership": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, from, to, registry, tokenID, err := t.convertArgs(args)
				if err != nil {
					return nil, err
				}

				tx, err := t.identitySrv.ExecuteAsync(ctx, registry, ABI, "transferFrom", from, to,
					utils.ByteSliceToBigInt(tokenID[:]))
				if err != nil {
					return nil, fmt.Errorf("failed to transfer nft ownership: %w", err)
				}

				overrides["transfer_owner_txn"] = tx.Hash()
				return nil, nil
			},
			Next: "wait_for_txn",
		},
		"wait_for_txn": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				tx := overrides["transfer_owner_txn"].(common.Hash)
				_, err = ethereum.IsTxnSuccessful(context.Background(), t.ethClient, tx)
				if err != nil {
					return nil, fmt.Errorf("txn not complete yet: %w", err)
				}

				_, _, to, registry, tokenID, err := t.convertArgs(args)
				if err != nil {
					return nil, err
				}
				owner, err := ownerOf(t.ethClient, registry, tokenID[:])
				if err != nil {
					return nil, fmt.Errorf("failed to get new owner of NFT: %w", err)
				}

				if !bytes.Equal(owner.Bytes(), to.Bytes()) {
					return nil, fmt.Errorf("new nft owner[%s] doesn't match expected one[%s]", owner, to)
				}

				return nil, nil
			},
		},
	}
}

func initiateTransferNFTJob(dispatcher jobs.Dispatcher, did identity.DID, to, registry common.Address,
	tokenID TokenID) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Transfer NFT owner", transferNFTJob, "transfer_ownership",
		[]interface{}{did, to, registry, tokenID}, make(map[string]interface{}), time.Time{})
	_, err := dispatcher.Dispatch(did, job)
	return job.ID, err
}

// MintNFTOnCCJob mints and NFT async.
// args are as follows
// accountID, tokenID, MintNFTOnCCRequest
type MintNFTOnCCJob struct {
	jobs.Base

	accountsSrv config.Service
	docSrv      documents.Service
	dispatcher  jobs.Dispatcher
	api         API
}

// New returns a new instance of MintNFTOnCCJob
func (m *MintNFTOnCCJob) New() gocelery.Runner {
	nm := &MintNFTOnCCJob{
		accountsSrv: m.accountsSrv,
		docSrv:      m.docSrv,
		dispatcher:  m.dispatcher,
		api:         m.api,
	}
	nm.Base = jobs.NewBase(nm.loadTasks())
	return nm
}

func (m *MintNFTOnCCJob) convertArgs(
	args []interface{}) (ctx context.Context, did identity.DID, tokenID TokenID, req MintNFTOnCCRequest,
	err error) {
	did = args[0].(identity.DID)
	tokenID = args[1].(TokenID)
	req = args[2].(MintNFTOnCCRequest)
	acc, err := m.accountsSrv.GetAccount(did[:])
	if err != nil {
		err = fmt.Errorf("failed to get account: %w", err)
		return
	}

	ctx = contextutil.WithAccount(context.Background(), acc)
	return ctx, did, tokenID, req, nil
}

func (m *MintNFTOnCCJob) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"add_nft_to_document": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				log.Infof("Adding NFT to document...")
				ctx, _, tokenID, req, err := m.convertArgs(args)
				if err != nil {
					return nil, err
				}
				doc, err := m.docSrv.GetCurrentVersion(ctx, req.DocumentID)
				if err != nil {
					return nil, fmt.Errorf("failed to get document: %w", err)
				}

				err = doc.AddNFT(req.GrantNFTReadAccess, req.RegistryAddress, tokenID[:], false)
				if err != nil {
					return nil, fmt.Errorf("failed to add nft to document: %w", err)
				}

				jobID, err := m.docSrv.Commit(ctx, doc)
				if err != nil {
					return nil, fmt.Errorf("failed to commit document: %w", err)
				}
				overrides["document_commit_job"] = jobID
				return nil, nil
			},
			Next: "wait_for_document_commit",
		},
		"wait_for_document_commit": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				log.Infof("Waiting for document to be anchored...")
				did := args[0].(identity.DID)
				jobID := overrides["document_commit_job"].(gocelery.JobID)
				job, err := m.dispatcher.Job(did, jobID)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch job: %w", err)
				}

				if !job.IsSuccessful() {
					return nil, fmt.Errorf("document not committed yet")
				}

				return nil, nil
			},
			Next: "mint_nft",
		},
		"mint_nft": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				log.Infof("Minting NFT on Centrifuge chain...")
				ctx, did, tokenID, req, err := m.convertArgs(args)
				if err != nil {
					return nil, err
				}

				requestData, err := prepareMintRequest(ctx, m.docSrv, tokenID, did, req.DocumentID, req.ProofFields,
					req.RegistryAddress, true, req.GrantNFTReadAccess, false, req.RegistryAddress)
				if err != nil {
					return nil, fmt.Errorf("failed to prepare mint request: %w", err)
				}

				proofs := toNFTOnCCProofs(requestData.Props, requestData.Values, requestData.Salts, requestData.Proofs)
				staticProofs := [2][32]byte{requestData.DataRoot, requestData.SignaturesRoot}
				mi := MintInfo{
					AnchorID:     requestData.AnchorID,
					StaticHashes: staticProofs,
					Proofs:       proofs,
				}

				extInfo, err := m.api.MintNFT(ctx,
					req.DepositAddress, types.H160(req.RegistryAddress), types.NewU256(*tokenID.BigInt()),
					AssetInfo{Metadata: []byte{}}, mi)
				if err != nil {
					return nil, err
				}

				overrides["ext_info"] = extInfo
				log.Infof("Successfully minted NFT[%s] on cent chain for anchorID[%s] with extrinsinc hash[%s]",
					tokenID.String(), requestData.AnchorID.String(), extInfo.Hash.Hex())
				return nil, nil
			},
		},
	}
}

func initiateNFTMintOnCC(dispatcher jobs.Dispatcher, did identity.DID, tokenID TokenID,
	req MintNFTOnCCRequest) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		"Mint NFT on Centrifuge Chain", nftOnCCJob, "add_nft_to_document",
		[]interface{}{did, tokenID, req}, make(map[string]interface{}), time.Time{})
	_, err := dispatcher.Dispatch(did, job)
	if err != nil {
		return nil, fmt.Errorf("failed to dispatch mint NFT on CC job: %w", err)
	}
	return job.ID, nil
}
