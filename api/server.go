package api

import (
	"io"
	"net"
	"net/http"
	_ "net/http/pprof" // we need this side effect that loads the pprof endpoints to defaultServerMux
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// ErrNoAuthHeader used for requests when header is not passed.
	ErrNoAuthHeader = errors.Error("'authorization' header missing")
)

var (
	log = logging.Logger("api-server")

	// noAuthPaths holds the paths that doesn't require header to be passed.
	noAuthPaths = [...]string{"/health.HealthCheckService/Ping"}
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

	apiAddr := c.config.GetServerAddress()
	grpcAddr, _, err := utils.GetFreeAddrPort()
	if err != nil {
		startupErr <- errors.New("failed to get random port for grpc: %v", err)
		return
	}

	// set http error interceptor
	runtime.HTTPError = httpResponseInterceptor
	opts := []grpc.ServerOption{grpcInterceptor()}

	grpcServer := grpc.NewServer(opts...)

	mux := http.NewServeMux()
	gwmux := runtime.NewServeMux()

	err = registerServices(ctx, c.config, grpcServer, gwmux, grpcAddr, []grpc.DialOption{grpc.WithInsecure()})
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
		Addr:    apiAddr,
		Handler: mux,
	}

	startUpErrOut := make(chan error)
	go func(startUpErrInner chan<- error) {
		conn, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			startUpErrInner <- err
			return
		}

		log.Infof("GRPC Server running at: %s\n", grpcAddr)
		err = grpcServer.Serve(conn)
		if err != nil {
			startUpErrInner <- err
		}
	}(startUpErrOut)

	go func(startUpErrInner chan<- error) {
		log.Infof("HTTP API running at: %s\n", c.config.GetServerAddress())
		log.Infof("Connecting to Network: %s\n", c.config.GetNetworkString())
		err = srv.ListenAndServe()
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
		grpcServer.GracefulStop()
		log.Info("API server stopped")
		return
	}
}

// grpcInterceptor returns a GRPC UnaryInterceptor for all grpc/http requests.
func grpcInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(auth)
}

// auth is the grpc unary interceptor to to check if the account ID is passed in the header.
// interceptor will check "authorisation" header. If not set, we return an error.
//
// at this point we are going with one interceptor. Once we have more than one interceptor,
// we can write a wrapper interceptor that will call the chain of interceptor
//
// Note: each handler can access accountID from the context: ctx.Value(api.AccountHeaderKey)
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

	ctx = context.WithValue(ctx, config.AccountHeaderKey, auth[0])
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
