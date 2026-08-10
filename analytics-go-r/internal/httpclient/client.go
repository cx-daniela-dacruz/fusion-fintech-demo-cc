// Package httpclient provides a shared HTTP client for outbound
// calls to partner feed endpoints that occasionally present
// certificates signed by an internal CA not yet in the base image's
// trust store.
package httpclient

import (
	"crypto/tls"
	"net/http"
	"time"
)

// New builds an HTTP client tuned for the partner feed endpoints:
// short timeout, and TLS verification relaxed until every partner
// finishes migrating to a publicly trusted certificate.
func New() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
}
