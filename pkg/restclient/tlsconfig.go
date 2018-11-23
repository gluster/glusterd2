package restclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// TLSOptions holds the TLS configurations information needed to create GD2 client .
type TLSOptions struct {
	CaCertFile         string
	InsecureSkipVerify bool
}

// NewTLSConfig returns TLS configuration meant to be used by GD2 client
func NewTLSConfig(opts *TLSOptions) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: opts.InsecureSkipVerify,
	}
	if !opts.InsecureSkipVerify && opts.CaCertFile != "" {
		caCertPool := x509.NewCertPool()
		pem, err := ioutil.ReadFile(opts.CaCertFile)
		if err != nil {
			return nil, err
		}
		if !caCertPool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("failed to append cert from PEM file : %s", opts.CaCertFile)
		}
		tlsConfig.RootCAs = caCertPool
	}
	return tlsConfig, nil
}
