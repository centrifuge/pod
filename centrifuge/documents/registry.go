package documents

import (
	"fmt"
	"sync"
)

//ServiceRegistry matches for a provided coreDocument the corresponding service
type ServiceRegistry struct {
	services map[string]Service
	mutex    sync.RWMutex
}

var registryInstance *ServiceRegistry
var registryOnce sync.Once

// GetRegistryInstance returns current registry instance
func GetRegistryInstance() *ServiceRegistry {
	registryOnce.Do(func() {
		registryInstance = &ServiceRegistry{}
		registryInstance.services = make(map[string]Service)
		registryInstance.mutex = sync.RWMutex{}
	})
	return registryInstance
}

// Register can register a service which implements the ModelDeriver interface
func (s *ServiceRegistry) Register(serviceID string, service Service) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.services[serviceID]; ok {
		return fmt.Errorf("service with provided id already registered")
	}

	s.services[serviceID] = service
	return nil
}

// LocateService will return the registered service for the embedded document type
func (s *ServiceRegistry) LocateService(serviceID string) (Service, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.services[serviceID] == nil {
		return nil, fmt.Errorf("no service for core document type is registered %s\n", serviceID)
	}

	return s.services[serviceID], nil
}
