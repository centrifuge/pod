package jobsv1

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/jobs"
	logging "github.com/ipfs/go-log"
)

// ErrInvalidJobID error for Invalid transaction ID.
const ErrInvalidJobID = errors.Error("Invalid Job ID")

// ErrInvalidAccountID error for Invalid account ID.
const ErrInvalidAccountID = errors.Error("Invalid Tenant ID")

var apiLog = logging.Logger("jobs-api")

// GRPCHandler returns an implementation of the TransactionServiceServer
func GRPCHandler(srv jobs.Manager, configService config.Service) jobspb.JobServiceServer {
	return grpcHandler{srv: srv, configService: configService}
}

// grpcHandler implements transactionspb.TransactionServiceServer
type grpcHandler struct {
	srv           jobs.Manager
	configService config.Service
}

// GetJobStatus returns transaction status of the given transaction id.
func (h grpcHandler) GetJobStatus(ctx context.Context, req *jobspb.JobStatusRequest) (*jobspb.JobStatusResponse, error) {
	ctxHeader, err := contextutil.Context(ctx, h.configService)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	id, err := jobs.FromString(req.JobId)
	if err != nil {
		return nil, errors.NewTypedError(ErrInvalidJobID, err)
	}

	tc, err := contextutil.Account(ctxHeader)
	if err != nil {
		return nil, ErrInvalidAccountID
	}

	accID, err := tc.GetIdentityID()
	if err != nil {
		return nil, ErrInvalidAccountID
	}

	did, err := identity.NewDIDFromBytes(accID)
	if err != nil {
		return nil, err
	}
	return h.srv.GetJobStatus(did, id)
}
