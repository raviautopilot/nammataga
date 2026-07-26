package client

import (
	"crypto/tls"
	"encoding/base64"
	"net/http"
)

// Authenticator dictates how to apply authentication to an http.Request or http.Client.
type Authenticator interface {
	Apply(req *http.Request) error
}

// BearerTokenAuth implements Bearer Token authentication.
type BearerTokenAuth struct {
	Token string
}

// Apply adds the Bearer Token to the Authorization header.
func (a *BearerTokenAuth) Apply(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return nil
}

// BasicAuth implements HTTP Basic authentication.
type BasicAuth struct {
	Username string
	Password string
}

// Apply adds the Basic Auth credentials to the Authorization header.
func (a *BasicAuth) Apply(req *http.Request) error {
	auth := a.Username + ":" + a.Password
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	return nil
}

// APIKeyAuth implements custom API Key authentication.
type APIKeyAuth struct {
	Key   string
	Value string
	In    string // "header" or "query"
}

// Apply injects the API key in the specified header or query param.
func (a *APIKeyAuth) Apply(req *http.Request) error {
	if a.In == "query" {
		q := req.URL.Query()
		q.Add(a.Key, a.Value)
		req.URL.RawQuery = q.Encode()
	} else {
		req.Header.Set(a.Key, a.Value)
	}
	return nil
}

// SSHKeyAuth represents signing request headers via an SSH private key.
type SSHKeyAuth struct {
	KeyName   string
	Signature string
}

// Apply adds custom headers for the signed SSH request.
func (a *SSHKeyAuth) Apply(req *http.Request) error {
	req.Header.Set("X-SSH-Key-Name", a.KeyName)
	req.Header.Set("X-SSH-Signature", a.Signature)
	return nil
}

// ClientCertAuth implements transport layer client certificate (mTLS) configuration.
type ClientCertAuth struct {
	Certificate tls.Certificate
}

// Apply is a no-op on the request; the transport loader handles its TLS configuration.
func (a *ClientCertAuth) Apply(req *http.Request) error {
	return nil
}
