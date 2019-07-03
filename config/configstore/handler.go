package configstore

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("account-api")

type grpcHandler struct {
	service config.Service
}

// GRPCAccountHandler returns an implementation of accountpb.AccountServiceServer
func GRPCAccountHandler(svc config.Service) accountpb.AccountServiceServer {
	return &grpcHandler{service: svc}
}

func (h grpcHandler) CreateAccount(ctx context.Context, data *accountpb.AccountData) (*accountpb.AccountData, error) {
	apiLog.Infof("Creating account: %v", data)
	accountConfig := new(Account)
	err := accountConfig.loadFromProtobuf(data)
	if err != nil {
		return nil, err
	}
	tc, err := h.service.CreateAccount(accountConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf()
}

func (h grpcHandler) UpdateAccount(ctx context.Context, req *accountpb.UpdateAccountRequest) (*accountpb.AccountData, error) {
	apiLog.Infof("Updating account: %v", req)
	accountConfig := new(Account)
	err := accountConfig.loadFromProtobuf(req.Data)
	if err != nil {
		return nil, err
	}
	tc, err := h.service.UpdateAccount(accountConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf()
}
