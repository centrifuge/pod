package documents

import (
	"fmt"
	"sync"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
)

//ServiceRegistry matches for a provided coreDocument the corresponding service
type ServiceRegistry struct {
	services map[string]ModelDeriver
	mutex    *sync.Mutex
}

var registryInstance *ServiceRegistry
var registryOnce sync.Once

func GetRegistryInstance() *ServiceRegistry {
	registryOnce.Do(func() {
		registryInstance = &ServiceRegistry{}
		registryInstance.services = make(map[string]ModelDeriver)
		registryInstance.mutex = &sync.Mutex{}
	})
	return registryInstance
}

//Register can register a service which implements the ModelDeriver interface
func (s *ServiceRegistry) Register(serviceID string, service ModelDeriver) error {

	s.mutex.Lock()
	if _, ok := s.services[serviceID]; ok {
		s.mutex.Unlock()
		return fmt.Errorf("service with provided id already registered")
	}

	s.services[serviceID] = service
	s.mutex.Unlock()

	return nil

}

//LocateService will return the registered service for the embedded document type
func (s *ServiceRegistry) LocateService(coreDocument *coredocumentpb.CoreDocument) (ModelDeriver, error) {

	if s.services[coreDocument.EmbeddedData.TypeUrl] == nil {
		return nil, fmt.Errorf("no service for core document type is registered")
	}

	return s.services[coreDocument.EmbeddedData.TypeUrl], nil

}
