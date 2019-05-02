package funding

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("funding-api")

// grpcHandler handles all the entity document related actions
// anchoring, sending, finding stored entity document
type grpcHandler struct {
	service Service
	config  config.Service
}

// GRPCHandler returns an implementation of entity.DocumentServiceServer
func GRPCHandler(config config.Service, srv Service) clientfundingpb.FundingServiceServer {
	return &grpcHandler{
		service: srv,
		config:  config,
	}
}

// Create handles a new funding document extension and adds it to an existing document
func (h *grpcHandler) Create(ctx context.Context, req *clientfundingpb.FundingCreatePayload) (*clientfundingpb.FundingResponse, error) {
	apiLog.Debugf("create funding request %v", req)
	ctxHeader, err := contextutil.Context(ctx, h.config)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	identifier, err := hexutil.Decode(req.Identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "identifier is an invalid hex string")
	}

	// returns model with added funding custom fields
	model, err := h.service.DeriveFromPayload(ctxHeader, req, identifier)
	if err != nil {
		return nil, err
	}

	model, jobID, _, err := h.service.Update(ctx, model)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not update document")
	}

	resp, err := h.service.DeriveFundingResponse(model, req)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "could not derive response")
	}

	resp.Header.JobId = jobID.String()
	return resp, nil

}
