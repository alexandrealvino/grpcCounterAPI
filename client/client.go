package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/spiffe"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"grpcCounterAPI/chat"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

func main() {
	begin := time.Now()
	goRoutines := 40
	var wg sync.WaitGroup
	out := make([]string, goRoutines + 1)
	for i := 1; i < len(out); i++ {
		wg.Add(1)
		go CreateClient(i, &out[i], &wg, goRoutines)
	}
	wg.Wait()
	println(time.Since(begin).String())
}

func CreateClient(i int, outVal *string, wg *sync.WaitGroup, goRoutines int)  {
	var conn *grpc.ClientConn

	clientCert, _ := tls.LoadX509KeyPair("/home/alexandre/GolandProjects/spire/wl/client/svid.pem","/home/alexandre/GolandProjects/spire/wl/client/key.pem")

	root, err := ioutil.ReadFile("/home/alexandre/GolandProjects/spire/wl/client/bundle.pem")
	if err != nil {
		fmt.Errorf("Failed to load certificates %v", err)
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(root) {
		fmt.Errorf("Failed to append certificates")
	}

	//cfg := &tls.Config{
	//	Certificates: []tls.Certificate{clientCert},
	//	RootCAs: cp,
	//	InsecureSkipVerify: true,
	//}
	var spiffeids []string
	spiffeids = append(spiffeids, "spiffe://example.org/server")

	t := TLSPeer{
		SpiffeIDs: spiffeids,
		TrustRoots: cp,
	}

	lcerts := []tls.Certificate{clientCert}
	cfg2 := t.NewTLSConfig(lcerts)

	//creds := credentials.NewTLS(cfg)
	creds2 := credentials.NewTLS(cfg2)

	//conn, err = grpc.Dial(":9000", grpc.WithTransportCredentials(creds))
	conn, err = grpc.Dial(":9000", grpc.WithTransportCredentials(creds2))
	if err != nil {
		log.Fatal("Could not connect: ", err)
	}
	defer conn.Close()

	c := chat.NewChatServiceClient(conn)
	message := chat.Message{Body: "Hello from client", A: strconv.Itoa(i)}
	now := time.Now().Minute()
	for now < 1 {
		time.Sleep(1000 * time.Millisecond)
		now = time.Now().Minute()
		fmt.Print(now)
	}
	limit := 0
	for limit <= 10000 - goRoutines {
		response, err := c.SayHello(context.Background(), &message)
		if err != nil {
			log.Fatal("Error when calling server: ", err)
		}
		log.Info(response.Body, " count: ", response.A)
		time.Sleep(time.Millisecond)
		limit,_ = strconv.Atoi(response.A)
	}
	*outVal = ""
	wg.Done()
}


// TLSPeer holds settings for creating SPIFFE-compatible TLS
// configurations.
type TLSPeer struct {
	// Slice of permitted SPIFFE IDs
	SpiffeIDs []string
	// Root certificates for validation
	TrustRoots *x509.CertPool
}

// NewTLSConfig creates a SPIFFE-compatible TLS configuration.
// We are opinionated towards mutual TLS. If you don't want
// mutual TLS, you'll need to update the returned config.
//
// `certs` contains one or more certificates to present to the
// other side of the connection, leaf first.
func (t *TLSPeer) NewTLSConfig(certs []tls.Certificate) *tls.Config {
	config := &tls.Config{
		// Disable validation/verification because we perform
		// this step with custom logic in `verifyPeerCertificate`
		ClientAuth:            tls.RequireAndVerifyClientCert,
		InsecureSkipVerify:    true,
		VerifyPeerCertificate: t.verifyPeerCertificate,
		Certificates:          certs,
	}

	return config
}

// verifyPeerCertificate serves callbacks from TLS listeners/dialers. It performs
// SPIFFE-specific validation steps on behalf of the golang TLS library
func (t *TLSPeer) verifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) (err error) {
	// First, parse all received certs
	var certs []*x509.Certificate
	for _, rawCert := range rawCerts {
		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return err
		}

		certs = append(certs, cert)
	}

	// Perform path validation
	// Leaf is the first off the wire:
	// https://tools.ietf.org/html/rfc5246#section-7.4.2
	intermediates := x509.NewCertPool()
	for _, intermediate := range certs[1:] {
		intermediates.AddCert(intermediate)
	}
	err = spiffe.VerifyCertificate(certs[0], intermediates, t.TrustRoots)
	if err != nil {
		return err
	}

	// Look for a known SPIFFE ID in the leaf
	err = spiffe.MatchID(t.SpiffeIDs, certs[0])
	if err != nil {
		return err
	}

	// If we are here, then all is well
	return nil
}