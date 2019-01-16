package configstore

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"
	"github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("config-api")

type grpcHandler struct {
	service config.Service
}

// GRPCHandler returns an implementation of configpb.ConfigServiceServer
func GRPCHandler(svc config.Service) configpb.ConfigServiceServer {
	return &grpcHandler{service: svc}
}

func (h grpcHandler) deriveAllTenantResponse(cfgs []config.TenantConfiguration) (*configpb.GetAllTenantResponse, error) {
	response := new(configpb.GetAllTenantResponse)
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

func (h grpcHandler) GetTenant(ctx context.Context, req *configpb.GetTenantRequest) (*configpb.TenantData, error) {
	id, err := hexutil.Decode(req.Identifier)
	if err != nil {
		return nil, err
	}
	tenantConfig, err := h.service.GetTenant(id)
	if err != nil {
		return nil, err
	}
	return tenantConfig.CreateProtobuf(), nil
}

func (h grpcHandler) GetAllTenants(ctx context.Context, _ *empty.Empty) (*configpb.GetAllTenantResponse, error) {
	cfgs, err := h.service.GetAllTenants()
	if err != nil {
		return nil, err
	}
	return h.deriveAllTenantResponse(cfgs)
}

func (h grpcHandler) CreateTenant(ctx context.Context, data *configpb.TenantData) (*configpb.TenantData, error) {
	apiLog.Infof("Creating tenant config: %v", data)
	tenantConfig := new(TenantConfig)
	tenantConfig.loadFromProtobuf(data)
	tc, err := h.service.CreateTenant(tenantConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf(), nil
}

func (h grpcHandler) GenerateTenant(context.Context, *empty.Empty) (*configpb.TenantData, error) {
	apiLog.Infof("Generating tenant config")
	tc, err := h.service.GenerateTenant()
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf(), nil
}

func (h grpcHandler) UpdateTenant(ctx context.Context, req *configpb.UpdateTenantRequest) (*configpb.TenantData, error) {
	apiLog.Infof("Updating tenant config: %v", req)
	tenantConfig := new(TenantConfig)
	tenantConfig.loadFromProtobuf(req.Data)
	tc, err := h.service.UpdateTenant(tenantConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf(), nil
}

func (h grpcHandler) DeleteTenant(ctx context.Context, req *configpb.GetTenantRequest) (*empty.Empty, error) {
	apiLog.Infof("Deleting tenant config: %v", req.Identifier)
	id, err := hexutil.Decode(req.Identifier)
	if err != nil {
		return nil, err
	}
	return nil, h.service.DeleteTenant(id)
}
