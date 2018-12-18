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

func (h grpcHandler) GetConfig(ctx context.Context, _ *empty.Empty) (*configpb.ConfigData, error) {
	return h.service.GetConfig()
}

func (h grpcHandler) GetTenant(ctx context.Context, req *configpb.GetTenantRequest) (*configpb.TenantData, error) {
	return h.service.GetTenant(req)
}

func (h grpcHandler) GetAllTenants(ctx context.Context, _ *empty.Empty) (*configpb.GetAllTenantResponse, error) {
	return h.service.GetAllTenants()
}

func (h grpcHandler) CreateConfig(ctx context.Context, data *configpb.ConfigData) (*configpb.ConfigData, error) {
	return h.service.CreateConfig(data)
}

func (h grpcHandler) CreateTenant(ctx context.Context, data *configpb.TenantData) (*configpb.TenantData, error) {
	return h.service.CreateTenant(data)
}

func (h grpcHandler) UpdateConfig(ctx context.Context, data *configpb.ConfigData) (*configpb.ConfigData, error) {
	return h.service.UpdateConfig(data)
}

func (h grpcHandler) UpdateTenant(ctx context.Context, req *configpb.UpdateTenantRequest) (*configpb.TenantData, error) {
	return h.service.UpdateTenant(req)
}

func (h grpcHandler) DeleteConfig(ctx context.Context, _ *empty.Empty) (*configpb.ConfigData, error) {
	return nil, h.service.DeleteConfig()
}

func (h grpcHandler) DeleteTenant(ctx context.Context, req *configpb.GetTenantRequest) (*configpb.TenantData, error) {
	return nil, h.service.DeleteTenant(req)
}
