// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: invoice/invoice.proto

/*
Package invoice is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package invoice

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

func request_InvoiceDocumentService_SendInvoiceDocument_0(ctx context.Context, marshaler runtime.Marshaler, client InvoiceDocumentServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq SendInvoiceEnvelope
	var metadata runtime.ServerMetadata

	if req.ContentLength > 0 {
		if err := marshaler.NewDecoder(req.Body).Decode(&protoReq); err != nil {
			return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
		}
	}

	msg, err := client.SendInvoiceDocument(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func request_InvoiceDocumentService_GetInvoiceDocument_0(ctx context.Context, marshaler runtime.Marshaler, client InvoiceDocumentServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetInvoiceDocumentEnvelope
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["document_identifier"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "document_identifier")
	}

	protoReq.DocumentIdentifier, err = runtime.Bytes(val)

	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "document_identifier", err)
	}

	msg, err := client.GetInvoiceDocument(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func request_InvoiceDocumentService_GetReceivedInvoiceDocuments_0(ctx context.Context, marshaler runtime.Marshaler, client InvoiceDocumentServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq empty.Empty
	var metadata runtime.ServerMetadata

	msg, err := client.GetReceivedInvoiceDocuments(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

// RegisterInvoiceDocumentServiceHandlerFromEndpoint is same as RegisterInvoiceDocumentServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterInvoiceDocumentServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterInvoiceDocumentServiceHandler(ctx, mux, conn)
}

// RegisterInvoiceDocumentServiceHandler registers the http handlers for service InvoiceDocumentService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterInvoiceDocumentServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterInvoiceDocumentServiceHandlerClient(ctx, mux, NewInvoiceDocumentServiceClient(conn))
}

// RegisterInvoiceDocumentServiceHandler registers the http handlers for service InvoiceDocumentService to "mux".
// The handlers forward requests to the grpc endpoint over the given implementation of "InvoiceDocumentServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "InvoiceDocumentServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "InvoiceDocumentServiceClient" to call the correct interceptors.
func RegisterInvoiceDocumentServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client InvoiceDocumentServiceClient) error {

	mux.Handle("POST", pattern_InvoiceDocumentService_SendInvoiceDocument_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
		resp, md, err := request_InvoiceDocumentService_SendInvoiceDocument_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_InvoiceDocumentService_SendInvoiceDocument_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_InvoiceDocumentService_GetInvoiceDocument_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
		resp, md, err := request_InvoiceDocumentService_GetInvoiceDocument_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_InvoiceDocumentService_GetInvoiceDocument_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_InvoiceDocumentService_GetReceivedInvoiceDocuments_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
		resp, md, err := request_InvoiceDocumentService_GetReceivedInvoiceDocuments_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_InvoiceDocumentService_GetReceivedInvoiceDocuments_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_InvoiceDocumentService_GetReceivedInvoiceDocuments_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
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
		resp, md, err := request_InvoiceDocumentService_GetReceivedInvoiceDocuments_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_InvoiceDocumentService_GetReceivedInvoiceDocuments_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_InvoiceDocumentService_SendInvoiceDocument_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"invoice", "send"}, ""))

	pattern_InvoiceDocumentService_GetInvoiceDocument_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"invoice", "get", "document_identifier"}, ""))

	pattern_InvoiceDocumentService_GetReceivedInvoiceDocuments_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"invoice", "getReceived"}, ""))
)

var (
	forward_InvoiceDocumentService_SendInvoiceDocument_0 = runtime.ForwardResponseMessage

	forward_InvoiceDocumentService_GetInvoiceDocument_0 = runtime.ForwardResponseMessage

	forward_InvoiceDocumentService_GetReceivedInvoiceDocuments_0 = runtime.ForwardResponseMessage
)
