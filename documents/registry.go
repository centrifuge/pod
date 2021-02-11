package documents

import (
	"sync"

	"github.com/centrifuge/go-centrifuge/errors"
)

// ServiceRegistry matches for a provided coreDocument the corresponding service
type ServiceRegistry struct {
	services map[string]Service
	mutex    sync.RWMutex
}

// NewServiceRegistry returns a new instance of service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]Service),
	}
}

// Register can register a service which implements the ModelDeriver interface
func (s *ServiceRegistry) Register(serviceID string, service Service) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.services[serviceID]; ok {
		return errors.New("service with provided id already registered")
	}

	s.services[serviceID] = service
	return nil
}

// LocateService will return the registered service for the embedded document type
func (s *ServiceRegistry) LocateService(serviceID string) (Service, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.services[serviceID] == nil {
		return nil, errors.New("service [%s] not registered", serviceID)
	}
	return s.services[serviceID], nil
}
