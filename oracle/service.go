package oracle

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("oracle")

const (
	// NFTValueUpdatedSignature used for finding events
	NFTValueUpdatedSignature = "NFTValueUpdated(uint indexed)"

	updateABI = `[{"constant":false,"inputs":[{"name":"tokenID","type":"uint256"},{"name":"_fingerprint","type":"bytes32"},{"name":"_result","type":"bytes32"}],"name":"update","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`
)

// service handles all interactions related to minting of NFTs for unpaid invoices on Ethereum
type service struct {
	identityService identity.Service
	ethClient       ethereum.Client
	queue           queue.TaskQueuer
	jobsManager     jobs.Manager
	docService      documents.Service
}

// newService creates the NFT Oracle Service given the parameters
func newService(
	docService documents.Service,
	identityService identity.Service,
	ethClient ethereum.Client,
	queue queue.TaskQueuer,
	jobsMan jobs.Manager) Service {
	return &service{
		docService:      docService,
		identityService: identityService,
		ethClient:       ethClient,
		queue:           queue,
		jobsManager:     jobsMan,
	}
}

func (s *service) PushAttributeToOracle(ctx context.Context, docID []byte, req PushAttributeToOracleRequest) (*PushToOracleResponse, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	didBytes := tc.GetIdentityID()

	// Mint NFT within transaction
	// We use context.Background() for now so that the transaction is only limited by ethereum timeouts
	did, err := identity.NewDIDFromBytes(didBytes)
	if err != nil {
		return nil, err
	}

	doc, err := s.docService.GetCurrentVersion(ctx, docID)
	if err != nil {
		return nil, err
	}

	value, err := doc.GetAttribute(req.AttributeKey)
	if err != nil {
		return nil, err
	}

	result, err := value.Value.ToBytes()
	if err != nil {
		return nil, err
	}

	fp, err := doc.CalculateTransitionRulesFingerprint()
	if err != nil {
		return nil, err
	}

	jobID, _, err := s.jobsManager.ExecuteWithinJob(contextutil.Copy(ctx), did, jobs.NilJobID(), "Updating NFT Oracle",
		s.updateOracleJob(ctx,
			req.OracleAddress,
			req.TokenID,
			utils.MustSliceToByte32(fp), utils.MustSliceToByte32(result)))
	if err != nil {
		return nil, err
	}

	return &PushToOracleResponse{
		JobID:                        jobID.String(),
		PushAttributeToOracleRequest: req,
	}, nil
}

func (s *service) updateOracleJob(ctx context.Context, oracleAddress common.Address, tokenID nft.TokenID, fingerprint, result [32]byte) func(accountID identity.DID, txID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
	return func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, errOut chan<- error) {
		// to common.Address, tokenId *big.Int, bytes32, properties [][]byte, values [][]byte, salts [][32]byte
		args := []interface{}{tokenID.BigInt(), fingerprint, result}

		block, err := s.ethClient.GetEthClient().BlockByNumber(context.Background(), nil)
		if err != nil {
			errOut <- errors.New("failed to get latest block: %v", err)
			return
		}

		txID, done, err := s.identityService.Execute(ctx, oracleAddress, updateABI, "update", args...)
		if err != nil {
			errOut <- err
			return
		}

		log.Infof("Sent off ethTX to update NFT oracle[Oracle Address: %s tokenID: %s] to NFT Oracle contract.",
			oracleAddress.String(), tokenID.String())

		err = <-done
		if err != nil {
			// some problem occurred in a child task
			errOut <- errors.New("update nft oracle contract failed for tokenID %s and transaction %s with error %s",
				tokenID.String(), txID, err.Error())
			return
		}

		txHash, done, err := ethereum.CreateWaitForEventJob(
			ctx, txMan, s.queue, accountID, jobID,
			NFTValueUpdatedSignature, block.Number(), oracleAddress, common.Hash(tokenID))
		if err != nil {
			errOut <- err
			return
		}

		err = <-done
		if err != nil {
			log.Errorf("failed to listen for NFT Value Update : %v\n", err)
			errOut <- err
			return
		}

		log.Infof("Asset successfully deposited with TX hash: %v\n", txHash.String())
		errOut <- nil
	}
}
