package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "crypto/tls"
	"crypto/x509"

    "github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
    "github.com/spiffe/go-spiffe/v2/workloadapi"
)

func main() {
    ctx := context.Background()

    log.Println("[Server] Creating X509Source (SPIFFE Workload API client)...")
    source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr("unix:///tmp/spire-agent/public/api.sock")))
    if err != nil {
        log.Fatalf("[Server] Failed to create X509Source: %v", err)
    }
    defer func() {
        log.Println("[Server] Closing X509Source...")
        source.Close()
    }()

    log.Println("[Server] Configuring TLS with mTLS...")
    tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeAny())

    // Hook to log handshake state
    tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
    log.Printf("[Server] VerifyPeerCertificate called with %d certs", len(rawCerts))
    return nil
	}
    tlsConfig.GetConfigForClient = func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
        log.Printf("[Server] TLS GetConfigForClient: %+v", hello)
        return nil, nil
    }

    server := &http.Server{
        Addr:      ":8443",
        TLSConfig: tlsConfig,
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log.Println("[Server] Received request")

            if r.TLS != nil {
                log.Printf("[Server] TLS handshake complete. Peer certificates count: %d", len(r.TLS.PeerCertificates))
            } else {
                log.Println("[Server] No TLS connection state")
            }

            clientID := "unknown"
if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
    cert := r.TLS.PeerCertificates[0]
    if len(cert.URIs) > 0 {
        clientID = cert.URIs[0].String()  // Take first URI SAN, which should be the SPIFFE ID
    } else {
        clientID = cert.Subject.CommonName // fallback, if you want
    }
    log.Printf("[Server] Client SPIFFE ID from cert: %s", clientID)
} else {
    log.Println("[Server] Client certificate not found")
}


            response := fmt.Sprintf("Hello, client with SPIFFE ID: %s\n", clientID)
            _, err := w.Write([]byte(response))
            if err != nil {
                log.Printf("[Server] Failed to write response: %v", err)
            } else {
                log.Println("[Server] Response written successfully")
            }
        }),
    }

    log.Println("[Server] Starting HTTPS server on :8443...")
    err = server.ListenAndServeTLS("", "")
    if err != nil {
        log.Fatalf("[Server] Server failed: %v", err)
    }
}
