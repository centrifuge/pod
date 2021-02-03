package nft

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	nftJob = "Mint NFT Job"
)

type task struct {
	runnerFunc gocelery.RunnerFunc
	next       string
}

// MintNFTJob mints and NFT async.
// args are as follows
// accountID, documentID, tokenID, MintNFTRequest
type MintNFTJob struct {
	accountsSrv config.Service
	docSrv      documents.Service
	dispatcher  jobsv2.Dispatcher
	ethClient   ethereum.Client
	api         API
	identitySrv identity.Service

	tasks map[string]task
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
	nm.loadTasks()
	return nm
}

// RunnerFunc returns runner func associated with the task
func (m *MintNFTJob) RunnerFunc(task string) gocelery.RunnerFunc {
	return m.tasks[task].runnerFunc
}

// Next returns the next task, if there is any, after the given task
func (m *MintNFTJob) Next(task string) (next string, ok bool) {
	next = m.tasks[task].next
	return next, next != ""
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

func (m *MintNFTJob) loadTasks() {
	m.tasks = map[string]task{
		"add_nft_to_document": {
			runnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, _, docID, tokenID, req, err := m.convertArgs(args)
				if err != nil {
					return nil, err
				}
				doc, err := m.docSrv.GetCurrentVersion(ctx, docID)
				if err != nil {
					return nil, fmt.Errorf("failed to get document: %w", err)
				}

				err = doc.AddNFT(req.GrantNFTReadAccess, req.RegistryAddress, tokenID[:])
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
			next: "wait_for_document_commit",
		},
		"wait_for_document_commit": {
			runnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
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
			next: "validate_nft_proofs",
		},
		"validate_nft_proofs": {
			runnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, did, _, tokenID, req, err := m.convertArgs(args)
				if err != nil {
					return nil, err
				}

				requestData, err := prepareMintRequest(ctx, m.docSrv, tokenID, did, req)
				if err != nil {
					return nil, fmt.Errorf("failed to prepare mint request: %w", err)
				}

				subProofs := toSubstrateProofs(requestData.Props, requestData.Values, requestData.Salts, requestData.Proofs)
				staticProofs := [3][32]byte{requestData.LeftDataRoot, requestData.RightDataRoot, requestData.SignaturesRoot}
				block, err := m.ethClient.GetEthClient().BlockByNumber(context.Background(), nil)
				if err != nil {
					return nil, fmt.Errorf("failed to get latest block: %v", err)
				}

				overrides["eth_from_block"] = block.Number()
				overrides["mint_request"] = requestData
				err = m.api.ValidateNFT(ctx, requestData.AnchorID, requestData.To, subProofs, staticProofs)
				if err != nil {
					return nil, fmt.Errorf("failed to validate nft proofs: %w", err)
				}

				log.Infof("Successfully validated Proofs on cent chain for anchorID: %s", requestData.AnchorID.String())
				return nil, nil
			},
			next: "wait_for_asset_deposit",
		},
		"wait_for_asset_deposit": {
			runnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, _, _, _, req, err := m.convertArgs(args)
				if err != nil {
					return nil, err
				}

				if utils.IsEmptyAddress(req.AssetManagerAddress) {
					return nil, nil
				}

				from := overrides["eth_from_block"].(*big.Int)
				requestData := overrides["mint_request"].(MintRequest)
				log.Infof("Triggered listener on AssetManager Address %s", req.AssetManagerAddress.Hex())
				err = ethereum.EventEmitted(
					ctx,
					m.ethClient.GetEthClient(),
					from,
					[]common.Address{req.AssetManagerAddress},
					AssetStoredEventSignature, requestData.BundledHash)
				if err != nil {
					return nil, err
				}

				log.Infof("Asset[%s] successfully deposited: %v\n", hexutil.Encode(requestData.BundledHash[:]))
				return nil, nil
			},
			next: "execute_mint_nft",
		},
		"execute_mint_nft": {
			runnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
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

				log.Infof("Sent off ethTX[%s] to mint [tokenID: %s, anchorID: %s, registry: %s, to NFT contract.",
					tx.Hash().Hex(),
					hexutil.Encode(requestData.TokenID.Bytes()),
					hexutil.Encode(requestData.AnchorID[:]),
					req.RegistryAddress.String())
				overrides["mint_nft_txn"] = tx.Hash()
				return nil, nil
			},
			next: "wait_mint_nft",
		},
		"wait_mint_nft": {
			runnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				tx := overrides["mint_nft_txn"].(common.Hash)
				_, err = ethereum.IsTxnSuccessful(context.Background(), m.ethClient, tx)
				return nil, err
			},
			next: "check_nft_owner",
		},
		"check_nft_owner": {
			runnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
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

func initiateNFTMint(dispatcher jobsv2.Dispatcher, did identity.DID, tokenID TokenID,
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
