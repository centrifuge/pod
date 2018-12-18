package config

import (
	"context"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"
	"github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("config-api")

type grpcHandler struct {
	service Service
}

func GRPCHandler(svc Service) configpb.ConfigServiceServer {
	return &grpcHandler{service: svc}
}

func (h grpcHandler) deriveAllTenantResponse(cfgs []*TenantConfig) (*configpb.GetAllTenantResponse, error) {
	response := new(configpb.GetAllTenantResponse)
	response.Data = make([]*configpb.TenantData, len(cfgs))
	for _, t := range cfgs {
		response.Data = append(response.Data, t.createProtobuf())
	}
	return response, nil
}

func (h grpcHandler) GetConfig(ctx context.Context, _ *empty.Empty) (*configpb.ConfigData, error) {
	return h.service.GetConfig()
}

func (h grpcHandler) GetTenant(ctx context.Context, req *configpb.GetTenantRequest) (*configpb.TenantData, error) {
	return h.service.GetTenant([]byte(req.Identifier))
}

func (h grpcHandler) GetAllTenants(ctx context.Context, _ *empty.Empty) (*configpb.GetAllTenantResponse, error) {
	cfgs, err := h.service.GetAllTenants()
	if err != nil {
		return nil, err
	}
	return h.deriveAllTenantResponse(cfgs)
}

func (h grpcHandler) CreateConfig(ctx context.Context, data *configpb.ConfigData) (*configpb.ConfigData, error) {
	apiLog.Infof("Creating node config: %v", data)
	return h.service.CreateConfig(data)
}

func (h grpcHandler) CreateTenant(ctx context.Context, data *configpb.TenantData) (*configpb.TenantData, error) {
	apiLog.Infof("Creating tenant config: %v", data)
	return h.service.CreateTenant(data)
}

func (h grpcHandler) UpdateConfig(ctx context.Context, data *configpb.ConfigData) (*configpb.ConfigData, error) {
	apiLog.Infof("Updating node config: %v", data)
	return h.service.UpdateConfig(data)
}

func (h grpcHandler) UpdateTenant(ctx context.Context, req *configpb.UpdateTenantRequest) (*configpb.TenantData, error) {
	apiLog.Infof("Updating tenant config: %v", req)
	return h.service.UpdateTenant(req.Data)
}

func (h grpcHandler) DeleteConfig(ctx context.Context, _ *empty.Empty) (*configpb.ConfigData, error) {
	apiLog.Infof("Deleting node config")
	return nil, h.service.DeleteConfig()
}

func (h grpcHandler) DeleteTenant(ctx context.Context, req *configpb.GetTenantRequest) (*configpb.TenantData, error) {
	apiLog.Infof("Deleting tenant config: %v", req.Identifier)
	return nil, h.service.DeleteTenant([]byte(req.Identifier))
}
