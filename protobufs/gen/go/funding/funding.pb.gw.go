// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: funding/funding.proto

/*
Package funpb is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package funpb

import (
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray

func request_FundingService_GetListVersion_0(ctx context.Context, marshaler runtime.Marshaler, client FundingServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetListVersionRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["document_id"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "document_id")
	}

	protoReq.DocumentId, err = runtime.String(val)

	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "document_id", err)
	}

	val, ok = pathParams["version_id"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "version_id")
	}

	protoReq.VersionId, err = runtime.String(val)

	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "version_id", err)
	}

	msg, err := client.GetListVersion(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

// RegisterFundingServiceHandlerFromEndpoint is same as RegisterFundingServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterFundingServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterFundingServiceHandler(ctx, mux, conn)
}

// RegisterFundingServiceHandler registers the http handlers for service FundingService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterFundingServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterFundingServiceHandlerClient(ctx, mux, NewFundingServiceClient(conn))
}

// RegisterFundingServiceHandlerClient registers the http handlers for service FundingService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "FundingServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "FundingServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "FundingServiceClient" to call the correct interceptors.
func RegisterFundingServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client FundingServiceClient) error {

	mux.Handle("GET", pattern_FundingService_GetListVersion_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		if cn, ok := w.(http.CloseNotifier); ok {
			go func(done <-chan struct{}, closed <-chan bool) {
				select {
				case <-done:
				case <-closed:
					cancel()
				}
			}(ctx.Done(), cn.CloseNotify())
		}
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_FundingService_GetListVersion_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FundingService_GetListVersion_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_FundingService_GetListVersion_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2, 2, 3, 1, 0, 4, 1, 5, 4, 2, 5}, []string{"v1", "documents", "document_id", "versions", "version_id", "funding_agreements"}, ""))
)

var (
	forward_FundingService_GetListVersion_0 = runtime.ForwardResponseMessage
)
