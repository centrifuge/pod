package oracle

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("oracle")

const (
	updateABI = `[{"constant":false,"inputs":[{"name":"tokenID","type":"uint256"},{"name":"_fingerprint","type":"bytes32"},{"name":"_result","type":"bytes32"}],"name":"update","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`
)

// service handles all interactions related to minting of NFTs for unpaid invoices on Ethereum
type service struct {
	identityService identity.Service
	ethClient       ethereum.Client
	docService      documents.Service
	dispatcher      jobs.Dispatcher
}

// newService creates the NFT Oracle Service given the parameters
func newService(
	docService documents.Service,
	identityService identity.Service,
	ethClient ethereum.Client,
	dispatcher jobs.Dispatcher) Service {
	return &service{
		docService:      docService,
		identityService: identityService,
		ethClient:       ethClient,
		dispatcher:      dispatcher,
	}
}

func (s *service) PushAttributeToOracle(ctx context.Context, docID []byte, req PushAttributeToOracleRequest) (*PushToOracleResponse, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	id := tc.GetIdentity()

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

	jobID, err := initOraclePushJob(
		s.dispatcher, id, req.OracleAddress, req.TokenID, utils.MustSliceToByte32(fp), utils.MustSliceToByte32(result))
	if err != nil {
		return nil, err
	}

	return &PushToOracleResponse{
		JobID:                        jobID.Hex(),
		PushAttributeToOracleRequest: req,
	}, nil
}
