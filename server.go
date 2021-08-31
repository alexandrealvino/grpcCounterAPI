package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"grpcCounterAPI/chat"
	"io/ioutil"
	"net"
	"sync"
)

func main() {
	log.Info("Server is running!")
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal("Failed to listen on port 9000: ", err, lis)
	}
	s := chat.Server{A: 0, Mutex: &sync.Mutex{}}

	serverCert, _ := tls.LoadX509KeyPair("/home/alexandre/GolandProjects/spire/wl/server-cert.pem","/home/alexandre/GolandProjects/spire/wl/server-key.pem")

	root, err := ioutil.ReadFile("/home/alexandre/GolandProjects/spire/wl/bundle.pem")
	if err != nil {
		fmt.Errorf("Failed to load certificates %v", err)
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(root) {
		fmt.Errorf("Failed to append certificates")
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs: cp,
	}

	cred := credentials.NewTLS(cfg)

	grpcServer := grpc.NewServer(grpc.Creds(cred))

	chat.RegisterChatServiceServer(grpcServer, &s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to listen on port 9000: ", err)
	}
}