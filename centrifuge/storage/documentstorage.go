package storage

import (
	"github.com/golang/protobuf/proto"
)

type DocumentStorageServiceInterface interface {
	GetDocument([]byte) proto.Message
	PutDocument(message proto.Message) error
}