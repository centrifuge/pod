//go:build unit

package http

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/pod/http/auth/access"

	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/errors"
	httpV2 "github.com/centrifuge/pod/http/v2"
	httpV3 "github.com/centrifuge/pod/http/v3"
	"github.com/stretchr/testify/assert"
)

func TestServer_Start(t *testing.T) {
	validationWrapperMock := access.NewValidationWrapperMock(t)
	validationWrapperFactoryMock := access.NewValidationWrapperFactoryMock(t)

	validationWrapperFactoryMock.On("GetValidationWrappers").
		Return(access.ValidationWrappers{validationWrapperMock}, nil).
		Once()

	configMock := config.NewConfigurationMock(t)
	configServiceMock := config.NewServiceMock(t)

	cctx := map[string]interface{}{
		bootstrap.BootstrappedConfig:         configMock,
		config.BootstrappedConfigStorage:     configServiceMock,
		httpV2.BootstrappedService:           &httpV2.Service{},
		httpV3.BootstrappedService:           &httpV3.Service{},
		BootstrappedValidationWrapperFactory: validationWrapperFactoryMock,
	}

	ctx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, cctx)
	ctx, canc := context.WithCancel(ctx)

	service := apiServer{config: configMock}

	configMock.On("GetServerAddress").
		Return("0.0.0.0:8082").
		Times(2)

	configMock.On("IsPProfEnabled").
		Return(true).
		Once()

	configMock.On("GetNetworkString").
		Return("network").
		Once()

	var wg sync.WaitGroup
	errChan := make(chan error)

	wg.Add(1)

	go service.Start(ctx, &wg, errChan)

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
	}

	canc()
	wg.Wait()
}

func TestServer_Start_RouterError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)

	// Node obj registry is not present in context, which should cause a failure.
	ctx, canc := context.WithCancel(context.Background())

	service := apiServer{config: configMock}

	configMock.On("GetServerAddress").
		Return("0.0.0.0:8082").
		Once()

	var wg sync.WaitGroup
	errChan := make(chan error)

	wg.Add(1)

	go service.Start(ctx, &wg, errChan)

	select {
	case err := <-errChan:
		assert.True(t, errors.IsOfType(ErrRouterCreation, err))
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Expected an error during startup")
	}

	canc()
	wg.Wait()
}

func TestServer_Start_HTTPServerError(t *testing.T) {
	validationWrapperMock := access.NewValidationWrapperMock(t)
	validationWrapperFactoryMock := access.NewValidationWrapperFactoryMock(t)

	validationWrapperFactoryMock.On("GetValidationWrappers").
		Return(access.ValidationWrappers{validationWrapperMock}, nil).
		Once()

	configMock := config.NewConfigurationMock(t)
	configServiceMock := config.NewServiceMock(t)

	cctx := map[string]interface{}{
		bootstrap.BootstrappedConfig:         configMock,
		config.BootstrappedConfigStorage:     configServiceMock,
		httpV2.BootstrappedService:           &httpV2.Service{},
		httpV3.BootstrappedService:           &httpV3.Service{},
		BootstrappedValidationWrapperFactory: validationWrapperFactoryMock,
	}

	ctx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, cctx)
	ctx, canc := context.WithCancel(ctx)

	service := apiServer{config: configMock}

	// The invalid port should trigger an error.
	configMock.On("GetServerAddress").
		Return("0.0.0.0:9999999").
		Times(2)

	configMock.On("IsPProfEnabled").
		Return(true).
		Once()

	configMock.On("GetNetworkString").
		Return("network").
		Once()

	var wg sync.WaitGroup
	errChan := make(chan error)

	wg.Add(1)

	go service.Start(ctx, &wg, errChan)

	select {
	case err := <-errChan:
		assert.NotNil(t, err)
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Expected an error during startup")
	}

	canc()
	wg.Wait()
}

func TestServer_Start_CanceledContext(t *testing.T) {
	validationWrapperMock := access.NewValidationWrapperMock(t)
	validationWrapperFactoryMock := access.NewValidationWrapperFactoryMock(t)

	validationWrapperFactoryMock.On("GetValidationWrappers").
		Return(access.ValidationWrappers{validationWrapperMock}, nil).
		Once()

	configMock := config.NewConfigurationMock(t)
	configServiceMock := config.NewServiceMock(t)

	cctx := map[string]interface{}{
		bootstrap.BootstrappedConfig:         configMock,
		config.BootstrappedConfigStorage:     configServiceMock,
		httpV2.BootstrappedService:           &httpV2.Service{},
		httpV3.BootstrappedService:           &httpV3.Service{},
		BootstrappedValidationWrapperFactory: validationWrapperFactoryMock,
	}

	ctx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, cctx)
	ctx, canc := context.WithCancel(ctx)
	canc()

	service := apiServer{config: configMock}

	configMock.On("GetServerAddress").
		Return("0.0.0.0:8082").
		Times(2)

	configMock.On("IsPProfEnabled").
		Return(true).
		Once()

	configMock.On("GetNetworkString").
		Return("network").
		Once()

	var wg sync.WaitGroup
	errChan := make(chan error)

	wg.Add(1)

	go service.Start(ctx, &wg, errChan)

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
	}

	wg.Wait()
}
