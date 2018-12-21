package transactions

import (
	"context"

	"github.com/centrifuge/go-centrifuge/common"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/satori/go.uuid"
)

const ErrInvalidTransactionID = errors.Error("Invalid Transaction ID")

// GRPCHandler returns an implementation of the TransactionServiceServer
func GRPCHandler(srv Service) transactionspb.TransactionServiceServer {
	return grpcHandler{srv: srv}
}

// grpcHandler implements transactionspb.TransactionServiceServer
type grpcHandler struct {
	srv Service
}

// GetTransactionStatus returns transaction status of the given transaction id.
func (h grpcHandler) GetTransactionStatus(ctx context.Context, req *transactionspb.TransactionStatusRequest) (*transactionspb.TransactionStatusResponse, error) {
	identity := common.DummyIdentity
	id := uuid.FromStringOrNil(req.TransactionId)
	if id == uuid.Nil {
		return nil, ErrInvalidTransactionID
	}

	return h.srv.GetTransactionStatus(identity, id)
}
