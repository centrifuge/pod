package configstore

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("account-api")

type grpcHandler struct {
	service config.Service
}

// GRPCHandler returns an implementation of configpb.ConfigServiceServer
func GRPCHandler(svc config.Service) configpb.ConfigServiceServer {
	return &grpcHandler{service: svc}
}

// GRPCAccountHandler returns an implementation of accountpb.AccountServiceServer
func GRPCAccountHandler(svc config.Service) accountpb.AccountServiceServer {
	return &grpcHandler{service: svc}
}

func (h grpcHandler) deriveAllAccountResponse(cfgs []config.Account) (*accountpb.GetAllAccountResponse, error) {
	response := new(accountpb.GetAllAccountResponse)
	for _, t := range cfgs {
		response.Data = append(response.Data, t.CreateProtobuf())
	}
	return response, nil
}

func (h grpcHandler) GetConfig(ctx context.Context, _ *empty.Empty) (*configpb.ConfigData, error) {
	nodeConfig, err := h.service.GetConfig()
	if err != nil {
		return nil, err
	}
	return nodeConfig.CreateProtobuf(), nil
}

func (h grpcHandler) GetAccount(ctx context.Context, req *accountpb.GetAccountRequest) (*accountpb.AccountData, error) {
	id, err := hexutil.Decode(req.Identifier)
	if err != nil {
		return nil, err
	}
	accountConfig, err := h.service.GetAccount(id)
	if err != nil {
		return nil, err
	}
	return accountConfig.CreateProtobuf(), nil
}

func (h grpcHandler) GetAllAccounts(ctx context.Context, req *empty.Empty) (*accountpb.GetAllAccountResponse, error) {
	cfgs, err := h.service.GetAllAccounts()
	if err != nil {
		return nil, err
	}
	return h.deriveAllAccountResponse(cfgs)
}

func (h grpcHandler) CreateAccount(ctx context.Context, data *accountpb.AccountData) (*accountpb.AccountData, error) {
	apiLog.Infof("Creating account: %v", data)
	accountConfig := new(Account)
	accountConfig.loadFromProtobuf(data)
	tc, err := h.service.CreateAccount(accountConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf(), nil
}

func (h grpcHandler) GenerateAccount(ctx context.Context, req *empty.Empty) (*accountpb.AccountData, error) {
	apiLog.Infof("Generating account")
	tc, err := h.service.GenerateAccount()
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf(), nil
}

func (h grpcHandler) UpdateAccount(ctx context.Context, req *accountpb.UpdateAccountRequest) (*accountpb.AccountData, error) {
	apiLog.Infof("Updating account: %v", req)
	accountConfig := new(Account)
	accountConfig.loadFromProtobuf(req.Data)
	tc, err := h.service.UpdateAccount(accountConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf(), nil
}
