package main

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/cmd"
)

// Below commands let you generate the coredocument protobuf stuff. It requires
// grpc-gateway protobuf labels to be checked out in a separate folder.
// NB: for now you will have to manually check out the protobuf & grpc-gateway project and update the path here
// To generate the go files, run: `cd centrifuge/server && go generate`
//go:generate protoc -I $PROTOBUF/src/ -I coredocument -I $GOPATH/src -I ../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I ../vendor/github.com/grpc-ecosystem/grpc-gateway --go_out=plugins=grpc:$GOPATH/src/ coredocument/coredocument.proto

//go:generate protoc -I $PROTOBUF/src/ -I. -Iinvoice -I $GOPATH/src -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I../vendor/github.com/grpc-ecosystem/grpc-gateway --go_out=plugins=grpc:$GOPATH/src/ invoice/invoice.proto
//go:generate protoc -I $PROTOBUF/src/ -I. -Iinvoice -I $GOPATH/src -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I../vendor/github.com/grpc-ecosystem/grpc-gateway --grpc-gateway_out=logtostderr=true:$GOPATH/src/ invoice/invoice.proto
//go:generate protoc -I $PROTOBUF/src/ -I. -Iinvoice -I $GOPATH/src -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I../vendor/github.com/grpc-ecosystem/grpc-gateway --swagger_out=logtostderr=true:$GOPATH/src/ invoice/invoice.proto


//go:generate protoc -I $PROTOBUF/src/ -I . -Ip2p -I $GOPATH/src -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I../vendor/github.com/grpc-ecosystem/grpc-gateway --go_out=plugins=grpc:$GOPATH/src/ p2p.proto


func main() {
	cmd.Execute()
}
