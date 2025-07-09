package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net/http"

	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func main() {
	ctx := context.Background()

	log.Println("[Client] Creating X509Source (SPIFFE Workload API client)...")
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr("unix:///tmp/spire-agent/public/api.sock")))
	if err != nil {
		log.Fatalf("[Client] Failed to create X509Source: %v", err)
	}
	defer func() {
		log.Println("[Client] Closing X509Source...")
		source.Close()
	}()

	log.Println("[Client] Configuring TLS with mTLS...")
	tlsConfig := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny())

	// Hook to log handshake state
	tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		log.Printf("[Client] VerifyPeerCertificate called with %d certs", len(rawCerts))
		return nil
	}

	tlsConfig.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		log.Println("[Client] GetClientCertificate called")
		svid, err := source.GetX509SVID()
		if err != nil {
			log.Printf("[Client] Error getting X509SVID: %v", err)
			return nil, err
		}

		tlsCert := tls.Certificate{
			Certificate: [][]byte{},
			PrivateKey:  svid.PrivateKey,
		}

		for _, cert := range svid.Certificates {
			tlsCert.Certificate = append(tlsCert.Certificate, cert.Raw)
		}

		log.Println("[Client] Returning client certificate")
		return &tlsCert, nil
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	url := "https://workload-server:8443"
	log.Printf("[Client] Sending GET request to %s", url)

	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("[Client] Failed to GET: %v", err)
	}
	defer resp.Body.Close()

	log.Println("[Client] Reading response body...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("[Client] Failed to read response body: %v", err)
	}

	log.Printf("[Client] Server response: %s", string(body))
}
