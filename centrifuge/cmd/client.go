package cmd

import (
	"context"
	"fmt"
	"log"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"crypto/x509"
	"crypto/tls"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/grpc"
)

func getDocument(client invoicegrpc.InvoiceDocumentServiceClient, id []byte) {
	doc, err := client.GetInvoiceDocument(context.Background(), &invoicegrpc.GetInvoiceDocumentEnvelope{id})
	if err != nil {
		panic(err)
	}
	log.Printf("Doc: %s\n", doc)
}

func loadCertPool() (certPool *x509.CertPool) {
	certPool = x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(server.InsecureCert))
	if !ok {
		panic("bad certs")
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
			RootCAs:loadCertPool(),
			ServerName: serverAddr,
			Certificates:[]tls.Certificate{cert},
			InsecureSkipVerify: true,
		})

		opts = append(opts, grpc.WithTransportCredentials(creds))

		conn, err := grpc.Dial(serverAddr, opts...)
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()
		client := invoicegrpc.NewInvoiceDocumentServiceClient(conn)

		getDocument(client, []byte("1"))
	},
}

func init() {
	viper.SetDefault("nodeHostname", "localhost")
	viper.SetDefault("nodePort", 8022)
	rootCmd.AddCommand(runClient)
}
