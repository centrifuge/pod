package context

import (
	"log"
	"sync"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage/invoicestorage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage/coredocumentstorage"
)

var (
	once sync.Once
	Node CentNodeWrapper
)

type CentNodeWrapper interface {
	BootstrapDependencies()
	GetCoreDocumentStorageService() *coredocumentstorage.StorageService
	GetInvoiceStorageService() *invoicestorage.StorageService
}

type CentNode struct {
	leveldb storage.DataStore

	coreDocumentStorageService *coredocumentstorage.StorageService
	invoiceStorageService *invoicestorage.StorageService
}


func bootstrapStorageServices(centNode *CentNode) {
	centNode.leveldb = storage.GetLeveldbStorage()

	centNode.coreDocumentStorageService = &coredocumentstorage.StorageService{}
	centNode.coreDocumentStorageService.SetStorageBackend(centNode.leveldb)
	log.Printf("CoreDocumentStorageService Initialized\n")

	centNode.invoiceStorageService = &invoicestorage.StorageService{}
	centNode.invoiceStorageService.SetStorageBackend(centNode.leveldb)
	log.Printf("InvoiceStorageService Initialized\n")
}

func (centNode *CentNode) BootstrapDependencies() {
	once.Do(func() {
		bootstrapStorageServices(centNode)
		Node = centNode
	})
}

func (centNode *CentNode) GetCoreDocumentStorageService() *coredocumentstorage.StorageService {
	return centNode.coreDocumentStorageService
}

func (centNode *CentNode) GetInvoiceStorageService() *invoicestorage.StorageService {
	return centNode.invoiceStorageService
}

////////////////////////////////////////////////////
///////// DEFAULT MOCKING CENT NODE FOR TESTS /////////
////////////////////////////////////////////////////
type MockCentNode struct {
	leveldb storage.DataStore

	coreDocumentStorageService *coredocumentstorage.StorageService
	invoiceStorageService *invoicestorage.StorageService
}

func bootstrapMockStorageServices(centNode *MockCentNode) {
	centNode.leveldb = storage.GetLeveldbStorage()

	centNode.coreDocumentStorageService = &coredocumentstorage.StorageService{}
	centNode.coreDocumentStorageService.SetStorageBackend(centNode.leveldb)
	log.Printf("CoreDocumentStorageService Mocked\n")

	centNode.invoiceStorageService = &invoicestorage.StorageService{}
	centNode.invoiceStorageService.SetStorageBackend(centNode.leveldb)
	log.Printf("InvoiceStorageService Mocked\n")
}

func (centNode *MockCentNode) BootstrapDependencies() {
	once.Do(func() {
		bootstrapMockStorageServices(centNode)
		Node = centNode
	})
}

func (centNode *MockCentNode) GetCoreDocumentStorageService() *coredocumentstorage.StorageService {
	return centNode.coreDocumentStorageService
}

func (centNode *MockCentNode) GetInvoiceStorageService() *invoicestorage.StorageService {
	return centNode.invoiceStorageService
}
////////////////////////////////////////////////////
////////////////////////////////////////////////////
////////////////////////////////////////////////////
