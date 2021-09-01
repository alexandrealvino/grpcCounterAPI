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
	"os"
	"os/signal"
	"sync"
	"syscall"
	//"istio.io/pkg/filewatcher"
)

func main() {
	log.Info("Server is running!")
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal("Failed to listen on port 9000: ", err, lis)
	}
	s := chat.Server{A: 0, Mutex: &sync.Mutex{}}

	serverCert, _ := tls.LoadX509KeyPair("/home/alexandre/GolandProjects/spire/wl/svid.pem","/home/alexandre/GolandProjects/spire/wl/key.pem")

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

type keypairReloader struct {
	certMu   sync.RWMutex
	cert     *tls.Certificate
	certPath string
	keyPath  string
	//watcher filewatcher.FileWatcher
}

func NewKeypairReloader(certPath, keyPath string) (*keypairReloader, error) {
	result := &keypairReloader{
		certPath: certPath,
		keyPath:  keyPath,
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	//result.watcher.Events(certPath)
	result.cert = &cert
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP)
		for range c {
			log.Printf("Received SIGHUP, reloading TLS certificate and key")
			if err := result.maybeReload(); err != nil {
				log.Printf("Keeping old TLS certificate because the new one could not be loaded: %v", err)
			}
		}
	}()
	return result, nil
}

func (kpr *keypairReloader) maybeReload() error {
	newCert, err := tls.LoadX509KeyPair(kpr.certPath, kpr.keyPath)
	if err != nil {
		return err
	}
	kpr.certMu.Lock()
	defer kpr.certMu.Unlock()
	kpr.cert = &newCert
	return nil
}

func (kpr *keypairReloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		kpr.certMu.RLock()
		defer kpr.certMu.RUnlock()
		return kpr.cert, nil
	}
}