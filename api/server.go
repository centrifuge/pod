package api

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	_ "net/http/pprof" // we need this side effect that loads the pprof endpoints to defaultServerMux
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	logging "github.com/ipfs/go-log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var log = logging.Logger("api-server")

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

	opts := []grpc.ServerOption{grpc.Creds(creds)}

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
	for {
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
		return nil, centerrors.Wrap(errors.New("could not load certpool"), "")
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
