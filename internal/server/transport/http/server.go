package http

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// ServerConfig содержит параметры HTTP-сервера.
type ServerConfig struct {
	Address  string
	CertFile string
	KeyFile  string
}

// NewServer создаёт сконфигурированный HTTP-сервер.
func NewServer(cfg ServerConfig, router chi.Router) *http.Server {
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		srv.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
	}

	return srv
}

// ListenAndServe запускает сервер с TLS или без.
func ListenAndServe(srv *http.Server, certFile, keyFile string) error {
	if certFile != "" && keyFile != "" {
		fmt.Printf("Starting HTTPS server on %s\n", srv.Addr)
		return srv.ListenAndServeTLS(certFile, keyFile)
	}

	fmt.Printf("Starting HTTP server on %s (no TLS)\n", srv.Addr)
	return srv.ListenAndServe()
}
