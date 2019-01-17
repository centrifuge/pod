package configstore

import (
	"context"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
)

// ErrDerivingAccount used as generic account deriver type
const ErrDerivingAccount = errors.Error("error deriving account")

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

// deriveAllTenantResponse derives all valid accounts, will not return accounts that fail at load time
func (h grpcHandler) deriveAllTenantResponse(cfgs []config.TenantConfiguration) (*accountpb.GetAllAccountResponse, error) {
	response := new(accountpb.GetAllAccountResponse)
	for _, t := range cfgs {
		tpb, err := t.CreateProtobuf()
		if err != nil {
			bID, err := t.GetIdentityID()
			if err != nil {
				apiLog.Errorf("%v", errors.NewTypedError(ErrDerivingAccount, errors.New("error getting ID: %v", err)))
			}
			apiLog.Errorf("%v", errors.NewTypedError(ErrDerivingAccount, errors.New("account [%s]: %v", hexutil.Encode(bID), err)))
			continue
		}
		response.Data = append(response.Data, tpb)
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
	tenantConfig, err := h.service.GetTenant(id)
	if err != nil {
		return nil, err
	}
	return tenantConfig.CreateProtobuf()
}

func (h grpcHandler) GetAllAccounts(ctx context.Context, req *empty.Empty) (*accountpb.GetAllAccountResponse, error) {
	cfgs, err := h.service.GetAllTenants()
	if err != nil {
		return nil, err
	}
	return h.deriveAllTenantResponse(cfgs)
}

func (h grpcHandler) CreateAccount(ctx context.Context, data *accountpb.AccountData) (*accountpb.AccountData, error) {
	apiLog.Infof("Creating account: %v", data)
	tenantConfig := new(TenantConfig)
	err := tenantConfig.loadFromProtobuf(data)
	if err != nil {
		return nil, err
	}
	tc, err := h.service.CreateTenant(tenantConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf()
}

func (h grpcHandler) GenerateAccount(ctx context.Context, req *empty.Empty) (*accountpb.AccountData, error) {
	apiLog.Infof("Generating account")
	tc, err := h.service.GenerateTenant()
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf()
}

func (h grpcHandler) UpdateAccount(ctx context.Context, req *accountpb.UpdateAccountRequest) (*accountpb.AccountData, error) {
	apiLog.Infof("Updating account: %v", req)
	tenantConfig := new(TenantConfig)
	err := tenantConfig.loadFromProtobuf(req.Data)
	if err != nil {
		return nil, err
	}
	tc, err := h.service.UpdateTenant(tenantConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf()
}
