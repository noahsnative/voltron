package server

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithPort(t *testing.T) {
	t.Run("Should set the server address", func(t *testing.T) {
		sut := New(http.NewServeMux(), WithPort(8080))
		assert.Equal(t, ":8080", sut.Addr)
	})
}

func TestWithCertificate(t *testing.T) {
	t.Run("Should set the server address", func(t *testing.T) {
		cert := tls.Certificate{}
		sut := New(http.NewServeMux(), WithTLSCertificate(&cert))
		assert.NotNil(t, sut.TLSConfig)
		assert.ElementsMatch(t, []tls.Certificate{cert}, sut.TLSConfig.Certificates)
	})
}
