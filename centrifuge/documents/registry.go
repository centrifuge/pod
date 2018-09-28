package documents

import (
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/matryer/resync"
)

type ServiceRegistry struct {
	services map[string]ModelDeriver
}

var registryInstance *ServiceRegistry
var registryOnce resync.Once

func GetRegistryInstance() *ServiceRegistry {
	registryOnce.Do(func() {
		registryInstance = &ServiceRegistry{}
		registryInstance.services = make(map[string]ModelDeriver)
	})
	return registryInstance
}

func (s *ServiceRegistry) Register(serviceId string, service ModelDeriver) error {

	if s.services[serviceId] != nil {
		return fmt.Errorf("service with provided id already registered")
	}

	s.services[serviceId] = service

	return nil

}

func (s *ServiceRegistry) Unregister(serviceId string) error {

	if s.services[serviceId] == nil {
		return fmt.Errorf("no service with provided id registered")
	}

	s.services[serviceId] = nil

	return nil

}

func (s *ServiceRegistry) LocateService(coreDocument *coredocumentpb.CoreDocument) (ModelDeriver, error) {

	if s.services[coreDocument.EmbeddedData.TypeUrl] == nil {
		return nil, fmt.Errorf("no service for core document type is registered")
	}

	return s.services[coreDocument.EmbeddedData.TypeUrl], nil

}

func KillRegistry() {
	registryInstance = nil
	registryOnce.Reset()
}
