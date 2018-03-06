package context

import (
	"log"
	"sync"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage/invoicestorage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage/coredocumentstorage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
)

var (
	once sync.Once
	Node CentNodeWrapper
)

type CentNodeWrapper interface {
	BootstrapDependencies()
	GetCoreDocumentStorageService() *coredocumentstorage.StorageService
	GetInvoiceStorageService() *invoicestorage.StorageService
	GetSigningService() *signatures.SigningService
}

type CentNode struct {
	leveldb storage.DataStore

	coreDocumentStorageService *coredocumentstorage.StorageService
	invoiceStorageService *invoicestorage.StorageService

	signingService *signatures.SigningService
}

func bootstrapStorageServices(centNode *CentNode) {
	centNode.leveldb = storage.GetLeveldbStorage()

	centNode.coreDocumentStorageService = &coredocumentstorage.StorageService{}
	centNode.coreDocumentStorageService.SetStorageBackend(centNode.leveldb)
	log.Printf("CoreDocumentStorageService Initialized\n")

	centNode.invoiceStorageService = &invoicestorage.StorageService{}
	centNode.invoiceStorageService.SetStorageBackend(centNode.leveldb)
	log.Printf("InvoiceStorageService Initialized\n")

	centNode.signingService = &signatures.SigningService{}
	// Until signing keys can be fetched from ethereum, we load the default keys from a config file.
	centNode.signingService.LoadPublicKeys()
	centNode.signingService.LoadIdentityKeyFromConfig()
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

func (centNode *CentNode) GetSigningService() *signatures.SigningService{
	return centNode.signingService
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
	signingService *signatures.SigningService
}

func bootstrapMockStorageServices(centNode *MockCentNode) {
	centNode.leveldb = storage.GetLeveldbStorage()

	centNode.coreDocumentStorageService = &coredocumentstorage.StorageService{}
	centNode.coreDocumentStorageService.SetStorageBackend(centNode.leveldb)
	log.Printf("CoreDocumentStorageService Mocked\n")

	centNode.invoiceStorageService = &invoicestorage.StorageService{}
	centNode.invoiceStorageService.SetStorageBackend(centNode.leveldb)
	log.Printf("InvoiceStorageService Mocked\n")

	centNode.signingService = &signatures.SigningService{}
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

func (centNode *MockCentNode) GetSigningService() *signatures.SigningService{
	return centNode.signingService
}

func (centNode *MockCentNode) GetInvoiceStorageService() *invoicestorage.StorageService {
	return centNode.invoiceStorageService
}
////////////////////////////////////////////////////
////////////////////////////////////////////////////
////////////////////////////////////////////////////
