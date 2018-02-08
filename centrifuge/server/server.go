package server
// LICENSE: Apache
// This is taken from https://github.com/philips/grpc-gateway-example/

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)


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

func getAddress () (addr string) {
	return fmt.Sprintf("%s:%s", viper.GetString("nodeHostname"), viper.GetString("nodePort"))
}

func loadCertPool() (certPool *x509.CertPool) {
	certPool = x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(InsecureCert))
	if !ok {
		panic("bad certs")
	}
	return
}

func loadKeyPair() (keyPair tls.Certificate) {
	var err error
	pair, err := tls.X509KeyPair([]byte(InsecureCert), []byte(InsecureKey))
	if err != nil {
		panic(err)
	}
	return pair
}

// ServeNode exposes the client APIs for interacting with a centrifuge node
func ServeNode() {
	certPool := loadCertPool()
	keyPair := loadKeyPair()
	addr := getAddress()

	creds := credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
		ServerName: addr,
		Certificates: []tls.Certificate{keyPair},
		InsecureSkipVerify: true,
	})

	opts := []grpc.ServerOption{
		grpc.Creds(creds)}

	grpcServer := grpc.NewServer(opts...)
	ctx := context.Background()

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: addr,
		RootCAs:    certPool,
		InsecureSkipVerify:true,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	mux := http.NewServeMux()
	gwmux := runtime.NewServeMux()

	RegisterServices(grpcServer, ctx, gwmux, addr, dopts)

	mux.Handle("/", gwmux)

	conn, err := net.Listen("tcp", fmt.Sprintf("%s:%s", viper.GetString("nodeHostname"), viper.GetString("nodePort")))
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:   addr,
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{keyPair},
			NextProtos:   []string{"h2"},
		},
	}

	log.Printf("grpc on port: %s\n", viper.GetString("nodePort"))
	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return
}