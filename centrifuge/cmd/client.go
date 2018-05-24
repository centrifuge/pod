package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func getDocument(client invoicepb.InvoiceDocumentServiceClient, id []byte) {
	doc, err := client.GetInvoiceDocument(context.Background(), &invoicepb.GetInvoiceDocumentEnvelope{DocumentIdentifier: id})
	if err != nil {
		panic(err)
	}
	log.Infof("Doc: %s\n", doc)
}

func loadCertPool() (certPool *x509.CertPool) {
	certPool = x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(server.InsecureCert))
	if !ok {
		log.Panic("bad certs")
	}
	return
}

// runCmd represents the run command
var runClient = &cobra.Command{
	Use:   "test-client",
	Short: "test client for grpc",
	Long:  `Testbed for interacting with GRPC in native go`,
	Run: func(cmd *cobra.Command, args []string) {
		serverAddr := fmt.Sprintf("%s:%d", viper.GetString("nodeHostname"), viper.GetInt("nodePort"))
		var opts []grpc.DialOption
		cert, err := tls.X509KeyPair([]byte(server.InsecureCert), []byte(server.InsecureKey))
		creds := credentials.NewTLS(&tls.Config{
			RootCAs:            loadCertPool(),
			ServerName:         serverAddr,
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		})

		opts = append(opts, grpc.WithTransportCredentials(creds))

		conn, err := grpc.Dial(serverAddr, opts...)
		if err != nil {
			log.Errorf("fail to dial: %v", err)
		}
		defer conn.Close()
		client := invoicepb.NewInvoiceDocumentServiceClient(conn)

		getDocument(client, []byte("1"))
	},
}

func init() {
	rootCmd.AddCommand(runClient)
}
