package api

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof" // we need this side effect that loads the pprof endpoints to defaultServerMux
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// ErrNoAuthHeader used for requests when header is not passed.
const ErrNoAuthHeader = errors.Error("'authorization' header missing")

var (
	log = logging.Logger("api-server")

	// noAuthPaths holds the paths that doesn't require header to be passed.
	noAuthPaths = [...]string{"/health.HealthCheckService/Ping"}

	// TenantKey represents the key used to fetch the tenant id from context
	TenantKey struct{}
)

// Config defines methods required for the package api
type Config interface {
	GetServerAddress() string
	GetServerPort() int
	GetNetworkString() string
	IsPProfEnabled() bool
}

// apiServer is an implementation of node.Server interface for serving HTTP based Centrifuge API
type apiServer struct {
	config Config
}

func (apiServer) Name() string {
	return "APIServer"
}

// Serve exposes the client APIs for interacting with a centrifuge node
func (c apiServer) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()
	certPool, err := loadCertPool()
	if err != nil {
		startupErr <- err
	}
	keyPair, err := loadKeyPair()
	if err != nil {
		startupErr <- err
	}

	addr := c.config.GetServerAddress()
	creds := credentials.NewTLS(&tls.Config{
		RootCAs:            certPool,
		ServerName:         addr,
		Certificates:       []tls.Certificate{keyPair},
		InsecureSkipVerify: true,
	})

	// set http error interceptor
	runtime.HTTPError = httpResponseInterceptor

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		// grpcInterceptor(), // enable this once we start requiring the tenant id to passed as header
	}

	grpcServer := grpc.NewServer(opts...)
	dcreds := credentials.NewTLS(&tls.Config{
		ServerName:         addr,
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	mux := http.NewServeMux()
	gwmux := runtime.NewServeMux()

	err = registerServices(ctx, c.config, grpcServer, gwmux, addr, dopts)
	if err != nil {
		startupErr <- err
		return
	}

	if c.config.IsPProfEnabled() {
		log.Info("added pprof endpoints to the server")
		mux.Handle("/debug/", http.DefaultServeMux)
	}

	mux.Handle("/", gwmux)
	srv := &http.Server{
		Addr:    addr,
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{keyPair},
			NextProtos:   []string{"h2"},
		},
	}

	startUpErrOut := make(chan error)
	go func(startUpErrInner chan<- error) {
		conn, err := net.Listen("tcp", c.config.GetServerAddress())
		if err != nil {
			startUpErrInner <- err
			return
		}

		log.Infof("HTTP/gRpc listening on Port: %d\n", c.config.GetServerPort())
		log.Infof("Connecting to Network: %s\n", c.config.GetNetworkString())
		err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))
		if err != nil {
			startUpErrInner <- err
		}
	}(startUpErrOut)

	// listen to context events as well as http server startup errors
	select {
	case err := <-startUpErrOut:
		// this could create an issue if the listeners are blocking.
		// We need to only propagate the error if its an error other than a server closed
		if err != nil && err.Error() != http.ErrServerClosed.Error() {
			startupErr <- err
			return
		}
		// most probably a graceful shutdown
		log.Info(err)
		return
	case <-ctx.Done():
		ctxn, _ := context.WithTimeout(context.Background(), 1*time.Second)
		// gracefully shutdown the server
		// we can only do this because srv is thread safe
		log.Info("Shutting down API server")
		err := srv.Shutdown(ctxn)
		if err != nil {
			panic(err)
		}
		log.Info("API server stopped")
		return
	}
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func loadCertPool() (certPool *x509.CertPool, err error) {
	certPool = x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(insecureCert))
	if !ok {
		return nil, errors.New("could not load certpool")
	}
	return certPool, nil
}

func loadKeyPair() (keyPair tls.Certificate, err error) {
	pair, err := tls.X509KeyPair([]byte(insecureCert), []byte(insecureKey))
	if err != nil {
		return pair, err
	}
	return pair, nil
}

// grpcInterceptor returns a GRPC UnaryInterceptor for all grpc/http requests.
func grpcInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(auth)
}

// auth is the grpc unary interceptor to to check if the tenant ID is passed in the header.
// interceptor will check "authorisation" header. If not set, we return an error.
//
// at this point we are going with one interceptor. Once we have more than one interceptor,
// we can write a wrapper interceptor that will call the chain of interceptor
//
// Note: each handler can access tenantID from the context: ctx.Value(api.TenantKey)
func auth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// if this request is for ping
	if utils.ContainsString(noAuthPaths[:], info.FullMethod) {
		return handler(ctx, req)
	}

	err = errors.NewHTTPError(http.StatusBadRequest, ErrNoAuthHeader)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, err
	}

	auth := md.Get("authorization")
	if len(auth) < 1 {
		return nil, err
	}

	ctx = context.WithValue(ctx, TenantKey, auth[0])
	return handler(ctx, req)
}

// httpResponseInterceptor will intercept if the we return an error from the grpc handler.
// we fetch the http code from the error using errors.GetHTTPDetails.
//
// copied some stuff from the DefaultHTTPError interceptor.
// Note: this is where we marshal the error.
func httpResponseInterceptor(_ context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	const fallback = `{"error": "failed to marshal error message"}`

	w.Header().Set("Content-Type", marshaler.ContentType())
	var errBody struct {
		Error string `protobuf:"bytes,1,name=error" json:"error"`
	}

	code, msg := errors.GetHTTPDetails(err)
	errBody.Error = msg
	buf, merr := marshaler.Marshal(errBody)
	if merr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			log.Infof("Failed to write response: %v", err)
		}
		return
	}

	w.WriteHeader(code)
	if _, err := w.Write(buf); err != nil {
		log.Infof("Failed to write response: %v", err)
	}
}
