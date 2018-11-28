// +build unit

package node

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockService struct {
	mustReturnStartErr bool
	receivedCTXDone    bool
	lock               sync.RWMutex
}

func (s *MockService) ReceivedCTXDone() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.receivedCTXDone
}

func (MockService) Name() string {
	return "MockNodeService"
}

func (s *MockService) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()
	if s.mustReturnStartErr {
		startupErr <- errors.New("")
		return
	}
	// mock out server behaviour
	for {
		select {
		case <-ctx.Done():
			s.lock.Lock()
			s.receivedCTXDone = true
			s.lock.Unlock()
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func TestNode_StartHappy(t *testing.T) {
	// create node with two mocked out services
	services := []Server{&MockService{mustReturnStartErr: false}, &MockService{mustReturnStartErr: false}}
	n := New(services)
	errChan := make(chan error)
	ctx, _ := context.WithTimeout(context.TODO(), time.Millisecond)
	go n.Start(ctx, errChan)
	// wait for startup and shutdown
	time.Sleep(2 * time.Second)
	for _, service := range services {
		assert.True(t, service.(*MockService).ReceivedCTXDone(), "context done signal should have been received")
	}
}

func TestNode_StartContextCancel(t *testing.T) {
	// create node with two mocked out services
	services := []Server{&MockService{mustReturnStartErr: false}, &MockService{mustReturnStartErr: false}}
	n := New(services)
	errChan := make(chan error)
	ctx, canc := context.WithCancel(context.TODO())
	go n.Start(ctx, errChan)
	// wait for startup
	time.Sleep(2 * time.Second)
	canc()
	// wait for shutdown
	time.Sleep(5 * time.Second)
	for _, service := range services {
		assert.True(t, service.(*MockService).ReceivedCTXDone(), "context done signal should have been received")
	}
}

func TestNode_StartChildError(t *testing.T) {
	// create node with two mocked out services
	services := []Server{&MockService{mustReturnStartErr: true}, &MockService{mustReturnStartErr: false}}
	n := New(services)
	errChan := make(chan error)
	ctx, _ := context.WithCancel(context.TODO())
	go n.Start(ctx, errChan)
	// wait for startup
	time.Sleep(2 * time.Second)

	// the second child would not receive cancel signal
	assert.False(t, services[0].(*MockService).ReceivedCTXDone(), "context done signal should have been received")
	// the second child should receive cancel signal
	assert.True(t, services[1].(*MockService).ReceivedCTXDone(), "context done signal should have been received")
}
