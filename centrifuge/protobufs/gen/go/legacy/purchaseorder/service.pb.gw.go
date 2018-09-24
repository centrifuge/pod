// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: legacy/purchaseorder/service.proto

/*
Package purchaseorderpb is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package purchaseorderpb

import (
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
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

func request_PurchaseOrderDocumentService_CreatePurchaseOrderProof_0(ctx context.Context, marshaler runtime.Marshaler, client PurchaseOrderDocumentServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq CreatePurchaseOrderProofEnvelope
	var metadata runtime.ServerMetadata

	if err := marshaler.NewDecoder(req.Body).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.CreatePurchaseOrderProof(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func request_PurchaseOrderDocumentService_AnchorPurchaseOrderDocument_0(ctx context.Context, marshaler runtime.Marshaler, client PurchaseOrderDocumentServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq AnchorPurchaseOrderEnvelope
	var metadata runtime.ServerMetadata

	if err := marshaler.NewDecoder(req.Body).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.AnchorPurchaseOrderDocument(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func request_PurchaseOrderDocumentService_SendPurchaseOrderDocument_0(ctx context.Context, marshaler runtime.Marshaler, client PurchaseOrderDocumentServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq SendPurchaseOrderEnvelope
	var metadata runtime.ServerMetadata

	if err := marshaler.NewDecoder(req.Body).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.SendPurchaseOrderDocument(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func request_PurchaseOrderDocumentService_GetPurchaseOrderDocument_0(ctx context.Context, marshaler runtime.Marshaler, client PurchaseOrderDocumentServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetPurchaseOrderDocumentEnvelope
	var metadata runtime.ServerMetadata

	if err := marshaler.NewDecoder(req.Body).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetPurchaseOrderDocument(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func request_PurchaseOrderDocumentService_GetReceivedPurchaseOrderDocuments_0(ctx context.Context, marshaler runtime.Marshaler, client PurchaseOrderDocumentServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq empty.Empty
	var metadata runtime.ServerMetadata

	msg, err := client.GetReceivedPurchaseOrderDocuments(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

// RegisterPurchaseOrderDocumentServiceHandlerFromEndpoint is same as RegisterPurchaseOrderDocumentServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterPurchaseOrderDocumentServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
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

	return RegisterPurchaseOrderDocumentServiceHandler(ctx, mux, conn)
}

// RegisterPurchaseOrderDocumentServiceHandler registers the http handlers for service PurchaseOrderDocumentService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterPurchaseOrderDocumentServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterPurchaseOrderDocumentServiceHandlerClient(ctx, mux, NewPurchaseOrderDocumentServiceClient(conn))
}

// RegisterPurchaseOrderDocumentServiceHandlerClient registers the http handlers for service PurchaseOrderDocumentService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "PurchaseOrderDocumentServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "PurchaseOrderDocumentServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "PurchaseOrderDocumentServiceClient" to call the correct interceptors.
func RegisterPurchaseOrderDocumentServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client PurchaseOrderDocumentServiceClient) error {

	mux.Handle("POST", pattern_PurchaseOrderDocumentService_CreatePurchaseOrderProof_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
		resp, md, err := request_PurchaseOrderDocumentService_CreatePurchaseOrderProof_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_PurchaseOrderDocumentService_CreatePurchaseOrderProof_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_PurchaseOrderDocumentService_AnchorPurchaseOrderDocument_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
		resp, md, err := request_PurchaseOrderDocumentService_AnchorPurchaseOrderDocument_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_PurchaseOrderDocumentService_AnchorPurchaseOrderDocument_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_PurchaseOrderDocumentService_SendPurchaseOrderDocument_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
		resp, md, err := request_PurchaseOrderDocumentService_SendPurchaseOrderDocument_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_PurchaseOrderDocumentService_SendPurchaseOrderDocument_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_PurchaseOrderDocumentService_GetPurchaseOrderDocument_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
		resp, md, err := request_PurchaseOrderDocumentService_GetPurchaseOrderDocument_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_PurchaseOrderDocumentService_GetPurchaseOrderDocument_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_PurchaseOrderDocumentService_GetReceivedPurchaseOrderDocuments_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
		resp, md, err := request_PurchaseOrderDocumentService_GetReceivedPurchaseOrderDocuments_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_PurchaseOrderDocumentService_GetReceivedPurchaseOrderDocuments_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_PurchaseOrderDocumentService_CreatePurchaseOrderProof_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"legacy", "purchaseorder", "proof"}, ""))

	pattern_PurchaseOrderDocumentService_AnchorPurchaseOrderDocument_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"legacy", "purchaseorder", "anchor"}, ""))

	pattern_PurchaseOrderDocumentService_SendPurchaseOrderDocument_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"legacy", "purchaseorder", "send"}, ""))

	pattern_PurchaseOrderDocumentService_GetPurchaseOrderDocument_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"legacy", "purchaseorder", "get"}, ""))

	pattern_PurchaseOrderDocumentService_GetReceivedPurchaseOrderDocuments_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"legacy", "purchaseorder", "getReceived"}, ""))
)

var (
	forward_PurchaseOrderDocumentService_CreatePurchaseOrderProof_0 = runtime.ForwardResponseMessage

	forward_PurchaseOrderDocumentService_AnchorPurchaseOrderDocument_0 = runtime.ForwardResponseMessage

	forward_PurchaseOrderDocumentService_SendPurchaseOrderDocument_0 = runtime.ForwardResponseMessage

	forward_PurchaseOrderDocumentService_GetPurchaseOrderDocument_0 = runtime.ForwardResponseMessage

	forward_PurchaseOrderDocumentService_GetReceivedPurchaseOrderDocuments_0 = runtime.ForwardResponseMessage
)
