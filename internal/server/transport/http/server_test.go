package http

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewServer_WithoutTLS(t *testing.T) {
	cfg := ServerConfig{
		Address:  ":8080",
		CertFile: "",
		KeyFile:  "",
	}
	router := chi.NewRouter()

	srv := NewServer(cfg, router)

	assert.Equal(t, ":8080", srv.Addr)
	assert.NotNil(t, srv.Handler)
	assert.Nil(t, srv.TLSConfig)
	assert.Equal(t, 15*time.Second, srv.ReadTimeout)
	assert.Equal(t, 15*time.Second, srv.WriteTimeout)
	assert.Equal(t, 60*time.Second, srv.IdleTimeout)
}

func TestNewServer_WithTLS(t *testing.T) {
	cfg := ServerConfig{
		Address:  ":8443",
		CertFile: "/path/to/cert.pem",
		KeyFile:  "/path/to/key.pem",
	}
	router := chi.NewRouter()

	srv := NewServer(cfg, router)

	assert.Equal(t, ":8443", srv.Addr)
	assert.NotNil(t, srv.TLSConfig)
	assert.Equal(t, uint16(tls.VersionTLS12), srv.TLSConfig.MinVersion)
	assert.NotEmpty(t, srv.TLSConfig.CurvePreferences)
	assert.NotEmpty(t, srv.TLSConfig.CipherSuites)
}

func TestNewServer_TLSCipherSuites(t *testing.T) {
	cfg := ServerConfig{
		Address:  ":8443",
		CertFile: "cert.pem",
		KeyFile:  "key.pem",
	}
	router := chi.NewRouter()

	srv := NewServer(cfg, router)

	expectedCiphers := []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}

	assert.Equal(t, expectedCiphers, srv.TLSConfig.CipherSuites)
}

func TestNewServer_TLSCurvePreferences(t *testing.T) {
	cfg := ServerConfig{
		Address:  ":8443",
		CertFile: "cert.pem",
		KeyFile:  "key.pem",
	}
	router := chi.NewRouter()

	srv := NewServer(cfg, router)

	expectedCurves := []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
	}

	assert.Equal(t, expectedCurves, srv.TLSConfig.CurvePreferences)
}

func TestNewServer_PartialTLSConfig(t *testing.T) {
	tests := []struct {
		name      string
		certFile  string
		keyFile   string
		expectTLS bool
	}{
		{
			name:      "only cert",
			certFile:  "cert.pem",
			keyFile:   "",
			expectTLS: false,
		},
		{
			name:      "only key",
			certFile:  "",
			keyFile:   "key.pem",
			expectTLS: false,
		},
		{
			name:      "both provided",
			certFile:  "cert.pem",
			keyFile:   "key.pem",
			expectTLS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{
				Address:  ":8080",
				CertFile: tt.certFile,
				KeyFile:  tt.keyFile,
			}
			router := chi.NewRouter()

			srv := NewServer(cfg, router)

			if tt.expectTLS {
				assert.NotNil(t, srv.TLSConfig)
			} else {
				assert.Nil(t, srv.TLSConfig)
			}
		})
	}
}
