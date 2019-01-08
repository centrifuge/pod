package transactions

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/satori/go.uuid"
)

// ErrInvalidTransactionID error for Invalid transaction ID.
const ErrInvalidTransactionID = errors.Error("Invalid Transaction ID")

// ErrInvalidTenantID error for Invalid tenant ID.
const ErrInvalidTenantID = errors.Error("Invalid Tenant ID")

// GRPCHandler returns an implementation of the TransactionServiceServer
func GRPCHandler(srv Service, configService configstore.Service) transactionspb.TransactionServiceServer {
	return grpcHandler{srv: srv, configService: configService}
}

// grpcHandler implements transactionspb.TransactionServiceServer
type grpcHandler struct {
	srv           Service
	configService configstore.Service
}

// GetTransactionStatus returns transaction status of the given transaction id.
func (h grpcHandler) GetTransactionStatus(ctx context.Context, req *transactionspb.TransactionStatusRequest) (*transactionspb.TransactionStatusResponse, error) {
	// TODO [multi-tenancy] use the tenant ID in the context for this
	tcs, err := h.configService.GetAllTenants()
	if err != nil || len(tcs) == 0 {
		return nil, ErrInvalidTenantID
	}
	id := uuid.FromStringOrNil(req.TransactionId)
	if id == uuid.Nil {
		return nil, ErrInvalidTransactionID
	}

	cid, err := identity.ToCentID(tcs[0].IdentityID)
	if err != nil || len(tcs) == 0 {
		return nil, ErrInvalidTenantID
	}

	return h.srv.GetTransactionStatus(cid, id)
}
