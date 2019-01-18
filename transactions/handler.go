package transactions

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

// ErrInvalidTransactionID error for Invalid transaction ID.
const ErrInvalidTransactionID = errors.Error("Invalid Transaction ID")

// ErrInvalidAccountID error for Invalid account ID.
const ErrInvalidAccountID = errors.Error("Invalid Tenant ID")

var apiLog = logging.Logger("transaction-api")

// GRPCHandler returns an implementation of the TransactionServiceServer
func GRPCHandler(srv Service, configService config.Service) transactionspb.TransactionServiceServer {
	return grpcHandler{srv: srv, configService: configService}
}

// grpcHandler implements transactionspb.TransactionServiceServer
type grpcHandler struct {
	srv           Service
	configService config.Service
}

// GetTransactionStatus returns transaction status of the given transaction id.
func (h grpcHandler) GetTransactionStatus(ctx context.Context, req *transactionspb.TransactionStatusRequest) (*transactionspb.TransactionStatusResponse, error) {
	ctxHeader, err := contextutil.Context(ctx, h.configService)
	if err != nil {
		apiLog.Error(err)
		return nil, err
	}

	id := uuid.FromStringOrNil(req.TransactionId)
	if id == uuid.Nil {
		return nil, ErrInvalidTransactionID
	}

	tc, err := contextutil.Account(ctxHeader)
	if err != nil {
		return nil, ErrInvalidAccountID
	}

	accID, err := tc.GetIdentityID()
	if err != nil {
		return nil, ErrInvalidAccountID
	}
	cid, err := identity.ToCentID(accID)
	if err != nil {
		return nil, ErrInvalidAccountID
	}

	return h.srv.GetTransactionStatus(cid, id)
}
